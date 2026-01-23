package protograph

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

func TestCreateEntityFromInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		wantNil  bool
		wantType string
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:     "device pointer",
			input:    &diodepb.Device{},
			wantNil:  false,
			wantType: "Device",
		},
		{
			name:     "site pointer",
			input:    &diodepb.Site{},
			wantNil:  false,
			wantType: "Site",
		},
		{
			name:     "interface pointer",
			input:    &diodepb.Interface{},
			wantNil:  false,
			wantType: "Interface",
		},
		{
			name:    "unknown type",
			input:   "string",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateEntityFromInterface(tt.input)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetEntityTypeName(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, ""},
		{"device", &diodepb.Device{}, "Device"},
		{"site", &diodepb.Site{}, "Site"},
		{"interface", &diodepb.Interface{}, "Interface"},
		{"manufacturer", &diodepb.Manufacturer{}, "Manufacturer"},
		{"unknown", "string", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEntityTypeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsKnownEntityType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil", nil, false},
		{"device", &diodepb.Device{}, true},
		{"site", &diodepb.Site{}, true},
		{"string", "string", false},
		{"int", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKnownEntityType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllEntityTypes(t *testing.T) {
	types := GetAllEntityTypes()
	assert.NotEmpty(t, types)
	assert.Contains(t, types, "Device")
	assert.Contains(t, types, "Site")
	assert.Contains(t, types, "Interface")
}

func TestGetEdgeTypesForField(t *testing.T) {
	tests := []struct {
		name       string
		fieldName  string
		sourceType string
		wantFwd    string
		wantRev    string
	}{
		{
			name:       "site field from device",
			fieldName:  "Site",
			sourceType: "Device",
			wantFwd:    "BELONGS_TO_SITE",
			wantRev:    "HAS_DEVICE",
		},
		{
			name:       "device_type field",
			fieldName:  "DeviceType",
			sourceType: "Device",
			wantFwd:    "BELONGS_TO_DEVICE_TYPE",
			wantRev:    "HAS_DEVICE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEdgeTypesForField(tt.fieldName, tt.sourceType)
			assert.Equal(t, tt.wantFwd, result.Forward)
			assert.Equal(t, tt.wantRev, result.Reverse)
		})
	}
}

func TestGetEdgeTypeForField(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  string
	}{
		{"Site", "BELONGS_TO_SITE"},
		{"DeviceType", "BELONGS_TO_DEVICE_TYPE"},
		{"Manufacturer", "BELONGS_TO_MANUFACTURER"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := GetEdgeTypeForField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetReverseEdgeTypeForField(t *testing.T) {
	tests := []struct {
		sourceType string
		expected   string
	}{
		{"Device", "HAS_DEVICE"},
		{"Site", "HAS_SITE"},
		{"Interface", "HAS_INTERFACE"},
	}

	for _, tt := range tests {
		t.Run(tt.sourceType, func(t *testing.T) {
			result := GetReverseEdgeTypeForField(tt.sourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNodeTypeForField(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  string
	}{
		{"site", "Site"},
		{"Site", "Site"},
		{"device_type", "DeviceType"},
		{"DeviceType", "DeviceType"},
		{"UnknownField", "UnknownField"}, // Returns as-is if not found
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := GetNodeTypeForField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
