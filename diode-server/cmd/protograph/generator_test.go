package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		wantErr     bool
	}{
		{
			name:        "valid package name",
			packageName: "protograph",
			wantErr:     false,
		},
		{
			name:        "empty package name",
			packageName: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.packageName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, gen)
			assert.Equal(t, tt.packageName, gen.packageName)
		})
	}
}

func TestConvertFieldNameToOneofName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple cases
		{"asn", "Asn"},
		{"device", "Device"},
		{"site", "Site"},

		// Snake case to CamelCase
		{"device_type", "DeviceType"},
		{"asn_range", "AsnRange"},
		{"ip_address", "IpAddress"},
		{"prefix_length", "PrefixLength"},
		{"contact_group", "ContactGroup"},

		// Special cases with acronyms
		{"l2vpn", "L2Vpn"},
		{"l2vpn_termination", "L2VpnTermination"},
		{"mac_address", "MacAddress"},
		{"vm_interface", "VmInterface"},
		{"wireless_lan", "WirelessLan"},
		{"wireless_lan_group", "WirelessLanGroup"},
		{"vlan", "Vlan"},
		{"vlan_group", "VlanGroup"},
		{"vlan_translation_policy", "VlanTranslationPolicy"},
		{"vlan_translation_rule", "VlanTranslationRule"},
		{"vrf", "Vrf"},
		{"rir", "Rir"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertFieldNameToOneofName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseOneofField(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name     string
		line     string
		expected EntityType
	}{
		{
			name: "simple entity type",
			line: "ASN asn = 2;",
			expected: EntityType{
				Name:        "ASN",
				PbType:      "*diodepb.ASN",
				OneofField:  "Entity_Asn",
				StructField: "Asn",
			},
		},
		{
			name: "snake_case field name",
			line: "DeviceType device_type = 25;",
			expected: EntityType{
				Name:        "DeviceType",
				PbType:      "*diodepb.DeviceType",
				OneofField:  "Entity_DeviceType",
				StructField: "DeviceType",
			},
		},
		{
			name: "field with options",
			line: "ServicePort service_port = 95 [(netbox_supported) = false];",
			expected: EntityType{
				Name:        "ServicePort",
				PbType:      "*diodepb.ServicePort",
				OneofField:  "Entity_ServicePort",
				StructField: "ServicePort",
			},
		},
		{
			name: "VLAN entity",
			line: "VLAN vlan = 50;",
			expected: EntityType{
				Name:        "VLAN",
				PbType:      "*diodepb.VLAN",
				OneofField:  "Entity_Vlan",
				StructField: "Vlan",
			},
		},
		{
			name:     "invalid line - no equals",
			line:     "DeviceType device_type",
			expected: EntityType{},
		},
		{
			name:     "invalid line - no semicolon",
			line:     "DeviceType device_type = 25",
			expected: EntityType{},
		},
		{
			name:     "empty line",
			line:     "",
			expected: EntityType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.parseOneofField(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractEntityTypes(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantErr    bool
		checkFirst *EntityType
	}{
		{
			name: "valid proto content",
			content: `
syntax = "proto3";
package diode.v1;

message Entity {
  oneof entity {
    Device device = 1;
    Site site = 2;
    DeviceType device_type = 3;
  }
}

message Device {
  string name = 1;
}
`,
			wantCount: 3,
			wantErr:   false,
			checkFirst: &EntityType{
				Name:        "Device",
				PbType:      "*diodepb.Device",
				OneofField:  "Entity_Device",
				StructField: "Device",
			},
		},
		{
			name: "proto with field options",
			content: `
message Entity {
  oneof entity {
    Device device = 1;
    ServicePort service_port = 2 [(netbox_supported) = false];
  }
}
`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "no Entity message",
			content: `
message Device {
  string name = 1;
}
`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "Entity without oneof",
			content: `
message Entity {
  string name = 1;
}
`,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities, err := gen.extractEntityTypes(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, entities, tt.wantCount)

			if tt.checkFirst != nil && len(entities) > 0 {
				assert.Equal(t, tt.checkFirst.Name, entities[0].Name)
				assert.Equal(t, tt.checkFirst.PbType, entities[0].PbType)
				assert.Equal(t, tt.checkFirst.OneofField, entities[0].OneofField)
				assert.Equal(t, tt.checkFirst.StructField, entities[0].StructField)
			}
		})
	}
}

func TestIsNestedType(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		fieldType string
		isNested  bool
	}{
		// Primitive types
		{"string", false},
		{"int32", false},
		{"int64", false},
		{"uint32", false},
		{"uint64", false},
		{"float", false},
		{"double", false},
		{"bool", false},
		{"bytes", false},
		{"fixed32", false},
		{"fixed64", false},
		{"sfixed32", false},
		{"sfixed64", false},
		{"sint32", false},
		{"sint64", false},

		// Nested message types
		{"Device", true},
		{"Site", true},
		{"DeviceType", true},
		{"IPAddress", true},
		{"CustomMessage", true},
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			result := gen.isNestedType(tt.fieldType)
			assert.Equal(t, tt.isNested, result)
		})
	}
}

func TestCalculateFieldConfidence(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name    string
		field   EntityField
		minConf float64
		maxConf float64
	}{
		// High confidence fields
		{
			name:    "id field",
			field:   EntityField{Name: "id", Type: "int64"},
			minConf: 0.9,
			maxConf: 0.9,
		},
		{
			name:    "name field",
			field:   EntityField{Name: "name", Type: "string"},
			minConf: 0.9,
			maxConf: 0.9,
		},
		{
			name:    "serial field",
			field:   EntityField{Name: "serial", Type: "string"},
			minConf: 0.9,
			maxConf: 0.9,
		},
		{
			name:    "mac_address field",
			field:   EntityField{Name: "mac_address", Type: "string"},
			minConf: 0.9,
			maxConf: 0.9,
		},
		{
			name:    "slug field",
			field:   EntityField{Name: "slug", Type: "string"},
			minConf: 0.9,
			maxConf: 0.9,
		},

		// Medium confidence fields
		{
			name:    "status field",
			field:   EntityField{Name: "status", Type: "string"},
			minConf: 0.7,
			maxConf: 0.7,
		},
		{
			name:    "role field",
			field:   EntityField{Name: "role", Type: "string"},
			minConf: 0.7,
			maxConf: 0.7,
		},
		{
			name:    "site field",
			field:   EntityField{Name: "site", Type: "Site", IsNested: true},
			minConf: 0.7,
			maxConf: 0.7,
		},

		// Nested object (non-medium confidence)
		{
			name:    "nested custom field",
			field:   EntityField{Name: "custom_object", Type: "CustomType", IsNested: true},
			minConf: 0.5,
			maxConf: 0.5,
		},

		// Default low confidence
		{
			name:    "description field",
			field:   EntityField{Name: "description", Type: "string"},
			minConf: 0.3,
			maxConf: 0.3,
		},
		{
			name:    "comments field",
			field:   EntityField{Name: "comments", Type: "string"},
			minConf: 0.3,
			maxConf: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := gen.calculateFieldConfidence(tt.field)
			assert.GreaterOrEqual(t, conf, tt.minConf)
			assert.LessOrEqual(t, conf, tt.maxConf)
		})
	}
}

func TestDetermineMatchType(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name      string
		field     EntityField
		matchType string
	}{
		// Exact match types
		{
			name:      "IP address field",
			field:     EntityField{Name: "ip_address", Type: "string"},
			matchType: "exact",
		},
		{
			name:      "address field",
			field:     EntityField{Name: "address", Type: "string"},
			matchType: "exact",
		},
		{
			name:      "serial field",
			field:     EntityField{Name: "serial", Type: "string"},
			matchType: "exact",
		},
		{
			name:      "mac field",
			field:     EntityField{Name: "mac_address", Type: "string"},
			matchType: "exact",
		},
		{
			name:      "boolean field",
			field:     EntityField{Name: "enabled", Type: "bool"},
			matchType: "exact",
		},

		// Fuzzy match types
		{
			name:      "name field",
			field:     EntityField{Name: "name", Type: "string"},
			matchType: "fuzzy",
		},
		{
			name:      "slug field",
			field:     EntityField{Name: "slug", Type: "string"},
			matchType: "fuzzy",
		},

		// Numeric match types
		{
			name:      "int32 field",
			field:     EntityField{Name: "count", Type: "int32"},
			matchType: "numeric",
		},
		{
			name:      "float field",
			field:     EntityField{Name: "weight", Type: "float"},
			matchType: "numeric",
		},

		// Exists match type (nested)
		{
			name:      "nested field",
			field:     EntityField{Name: "device", Type: "Device", IsNested: true},
			matchType: "exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.determineMatchType(tt.field)
			assert.Equal(t, tt.matchType, result)
		})
	}
}

func TestGenerateJSONPath(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name     string
		field    EntityField
		expected string
	}{
		{
			name:     "simple field",
			field:    EntityField{Name: "name", Type: "string", IsNested: false},
			expected: "name",
		},
		{
			name:     "nested field",
			field:    EntityField{Name: "device", Type: "Device", IsNested: true},
			expected: "device.name",
		},
		{
			name:     "another nested field",
			field:    EntityField{Name: "site", Type: "Site", IsNested: true},
			expected: "site.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateJSONPath(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseFieldLine(t *testing.T) {
	gen, err := NewGenerator("test")
	require.NoError(t, err)

	tests := []struct {
		name    string
		line    string
		want    EntityField
		wantErr bool
	}{
		{
			name: "simple string field",
			line: "string name = 1;",
			want: EntityField{
				Name:       "name",
				Type:       "string",
				IsOptional: false,
				IsRepeated: false,
				IsNested:   false,
			},
			wantErr: false,
		},
		{
			name: "optional field",
			line: "optional string description = 2;",
			want: EntityField{
				Name:       "description",
				Type:       "string",
				IsOptional: true,
				IsRepeated: false,
				IsNested:   false,
			},
			wantErr: false,
		},
		{
			name: "repeated field",
			line: "repeated string tags = 3;",
			want: EntityField{
				Name:       "tags",
				Type:       "string",
				IsOptional: false,
				IsRepeated: true,
				IsNested:   false,
			},
			wantErr: false,
		},
		{
			name: "nested message field",
			line: "Device device = 4;",
			want: EntityField{
				Name:       "device",
				Type:       "Device",
				IsOptional: false,
				IsRepeated: false,
				IsNested:   true,
			},
			wantErr: false,
		},
		{
			name: "field with comment",
			line: "string name = 1; // This is the name",
			want: EntityField{
				Name:       "name",
				Type:       "string",
				IsOptional: false,
				IsRepeated: false,
				IsNested:   false,
			},
			wantErr: false,
		},
		{
			name:    "invalid line",
			line:    "not a valid field",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gen.parseFieldLine(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.IsOptional, got.IsOptional)
			assert.Equal(t, tt.want.IsRepeated, got.IsRepeated)
			assert.Equal(t, tt.want.IsNested, got.IsNested)
		})
	}
}

func TestGenerateCode(t *testing.T) {
	gen, err := NewGenerator("protograph")
	require.NoError(t, err)

	entities := []EntityType{
		{
			Name:        "Device",
			PbType:      "*diodepb.Device",
			OneofField:  "Entity_Device",
			StructField: "Device",
		},
		{
			Name:        "Site",
			PbType:      "*diodepb.Site",
			OneofField:  "Entity_Site",
			StructField: "Site",
		},
	}

	code, err := gen.GenerateCode(entities)
	require.NoError(t, err)

	// Verify the generated code contains expected elements
	assert.Contains(t, code, "package protograph")
	assert.Contains(t, code, "CreateEntityFromInterface")
	assert.Contains(t, code, "GetEntityTypeName")
	assert.Contains(t, code, "IsKnownEntityType")
	assert.Contains(t, code, "GetAllEntityTypes")
	assert.Contains(t, code, "Entity_Device")
	assert.Contains(t, code, "Entity_Site")
	assert.Contains(t, code, "*diodepb.Device")
	assert.Contains(t, code, "*diodepb.Site")

	// Verify it's valid Go code (no syntax errors from formatting)
	assert.True(t, strings.HasPrefix(code, "// Code generated"))
}
