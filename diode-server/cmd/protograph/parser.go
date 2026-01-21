package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ParseProtobuf parses the protobuf file and extracts all entity types from the Entity oneof
func (g *Generator) ParseProtobuf(protoFile string) ([]EntityType, error) {
	// Read and parse the proto file directly by parsing its text content
	entities, err := g.parseProtoFileDirectly(protoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto file directly: %w", err)
	}

	// Parse field information for each entity
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read proto file for field parsing: %w", err)
	}

	for i := range entities {
		fields, err := g.parseEntityFields(string(content), entities[i].Name)
		if err != nil {
			return nil, fmt.Errorf("failed to parse fields for entity %s: %w", entities[i].Name, err)
		}

		entities[i].Fields = fields
	}

	return entities, nil
}

// parseProtoFileDirectly parses the proto file by reading its content and extracting Entity oneof fields
func (g *Generator) parseProtoFileDirectly(protoFile string) ([]EntityType, error) {
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read proto file: %w", err)
	}

	return g.extractEntityTypes(string(content))
}

// extractEntityTypes extracts entity types from the protobuf content by parsing the Entity oneof
func (g *Generator) extractEntityTypes(content string) ([]EntityType, error) {
	var entities []EntityType

	// Find the Entity message definition
	lines := strings.Split(content, "\n")
	inEntityMessage := false
	inOneof := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Find the Entity message
		if strings.HasPrefix(line, "message Entity") {
			inEntityMessage = true
			continue
		}

		// Exit Entity message when we find another message or reach end
		if inEntityMessage && (strings.HasPrefix(line, "message ") || strings.HasPrefix(line, "}")) {
			if strings.HasPrefix(line, "}") && inOneof {
				// End of oneof block
				inOneof = false
				continue
			}
			if strings.HasPrefix(line, "message ") && line != "message Entity" {
				// Start of another message, exit Entity
				break
			}
		}

		if !inEntityMessage {
			continue
		}

		// Find the oneof entity block
		if strings.HasPrefix(line, "oneof entity") {
			inOneof = true
			continue
		}

		// Skip lines until we're in the oneof block
		if !inOneof {
			continue
		}

		// Parse oneof fields
		if strings.Contains(line, " = ") && !strings.HasPrefix(line, "}") {
			entity := g.parseOneofField(line)
			if entity.Name != "" {
				entities = append(entities, entity)
			}
		}
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entity types found in protobuf file")
	}

	return entities, nil
}

// parseOneofField parses a single oneof field line and extracts entity information
func (g *Generator) parseOneofField(line string) EntityType {
	// Example line: "ASN asn = 2;"
	// Example line: "DeviceType device_type = 25;"
	// Example line: "ServicePort service_port = 95 [(netbox_supported) = false];"

	line = strings.TrimSpace(line)
	if !strings.Contains(line, " = ") || !strings.HasSuffix(line, ";") {
		return EntityType{}
	}

	// Remove the semicolon
	line = strings.TrimSuffix(line, ";")

	// Remove field options like [(netbox_supported) = false] before parsing
	// Field options are enclosed in square brackets
	if idx := strings.Index(line, "["); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}

	parts := strings.Split(line, " = ")
	if len(parts) != 2 {
		return EntityType{}
	}

	// Get the left side which contains the type and field name
	leftPart := strings.TrimSpace(parts[0])
	fieldParts := strings.Fields(leftPart)
	if len(fieldParts) != 2 {
		return EntityType{}
	}

	typeName := fieldParts[0]
	fieldName := fieldParts[1]

	// Convert field name from snake_case to the correct Go oneof field name
	oneofFieldName := convertFieldNameToOneofName(fieldName)

	return EntityType{
		Name:        typeName,
		PbType:      fmt.Sprintf("*diodepb.%s", typeName),
		OneofField:  fmt.Sprintf("Entity_%s", oneofFieldName),
		StructField: oneofFieldName, // The field name within the oneof struct
	}
}

// convertFieldNameToOneofName converts proto field names to Go oneof field names
// e.g., "asn" -> "Asn", "asn_range" -> "AsnRange", "device_type" -> "DeviceType"
func convertFieldNameToOneofName(fieldName string) string {
	// Handle special cases for acronyms first
	switch fieldName {
	case "l2vpn":
		return "L2Vpn"
	case "l2vpn_termination":
		return "L2VpnTermination"
	case "mac_address":
		return "MacAddress"
	case "vm_interface":
		return "VmInterface"
	case "wireless_lan":
		return "WirelessLan"
	case "wireless_lan_group":
		return "WirelessLanGroup"
	case "vlan":
		return "Vlan"
	case "vlan_group":
		return "VlanGroup"
	case "vlan_translation_policy":
		return "VlanTranslationPolicy"
	case "vlan_translation_rule":
		return "VlanTranslationRule"
	case "vrf":
		return "Vrf"
	case "rir":
		return "Rir"
	}

	// Convert snake_case to CamelCase for regular cases
	parts := strings.Split(fieldName, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

// parseEntityFields extracts field information from a specific message type
func (g *Generator) parseEntityFields(content, entityName string) ([]EntityField, error) {
	var fields []EntityField

	// Find the message definition for this entity
	lines := strings.Split(content, "\n")
	inMessage := false
	messagePattern := fmt.Sprintf("message %s", entityName)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Find the message start
		if strings.HasPrefix(line, messagePattern+" ") || line == messagePattern+" {" {
			inMessage = true
			continue
		}

		// Exit message when we find closing brace or another message
		if inMessage && (strings.HasPrefix(line, "}") || strings.HasPrefix(line, "message ")) {
			if strings.HasPrefix(line, "}") {
				break
			}
			inMessage = false
			continue
		}

		// Parse field definitions
		if inMessage && strings.Contains(line, "=") && !strings.HasPrefix(line, "//") && !strings.Contains(line, "oneof") {
			field, err := g.parseFieldLine(line)
			if err != nil {
				continue // Skip malformed lines
			}

			// Analyze field characteristics
			field.Confidence = g.calculateFieldConfidence(field)
			field.MatchType = g.determineMatchType(field)
			field.JSONPath = g.generateJSONPath(field)

			fields = append(fields, field)
		}
	}

	return fields, nil
}

// parseFieldLine parses a single field definition line from protobuf
func (g *Generator) parseFieldLine(line string) (EntityField, error) {
	// Remove comments and clean up the line
	if commentIdx := strings.Index(line, "//"); commentIdx != -1 {
		line = line[:commentIdx]
	}
	line = strings.TrimSpace(line)

	// Regular expression to match protobuf field definitions
	// Matches: [optional/repeated] type name = number [options];
	fieldRegex := regexp.MustCompile(`(optional|repeated)?\s*([A-Za-z_][A-Za-z0-9_]*)\s+([a-z_][a-z_0-9]*)\s*=\s*\d+`)
	matches := fieldRegex.FindStringSubmatch(line)

	if len(matches) < 4 {
		return EntityField{}, fmt.Errorf("invalid field line: %s", line)
	}

	field := EntityField{
		Name:       matches[3],
		Type:       matches[2],
		IsOptional: matches[1] == "optional" || strings.Contains(line, "oneof"),
		IsRepeated: matches[1] == "repeated",
		IsNested:   g.isNestedType(matches[2]),
	}

	return field, nil
}

// isNestedType determines if a field type is a nested message type
func (g *Generator) isNestedType(fieldType string) bool {
	// Primitive types in protobuf
	primitiveTypes := map[string]bool{
		"string":   true,
		"int32":    true,
		"int64":    true,
		"uint32":   true,
		"uint64":   true,
		"float":    true,
		"double":   true,
		"bool":     true,
		"bytes":    true,
		"fixed32":  true,
		"fixed64":  true,
		"sfixed32": true,
		"sfixed64": true,
		"sint32":   true,
		"sint64":   true,
	}

	return !primitiveTypes[fieldType]
}

// calculateFieldConfidence assigns confidence scores based on field characteristics
func (g *Generator) calculateFieldConfidence(field EntityField) float64 {
	// High confidence fields (unique identifiers)
	highConfidenceFields := map[string]bool{
		"id":          true,
		"name":        true,
		"serial":      true,
		"mac_address": true,
		"address":     true,
		"slug":        true,
	}

	// Medium confidence fields (important attributes)
	mediumConfidenceFields := map[string]bool{
		"status":       true,
		"role":         true,
		"type":         true,
		"model":        true,
		"manufacturer": true,
		"site":         true,
		"location":     true,
		"device":       true,
		"interface":    true,
	}

	if highConfidenceFields[field.Name] {
		return 0.9
	}
	if mediumConfidenceFields[field.Name] {
		return 0.7
	}
	if field.IsNested {
		return 0.5 // Nested objects have moderate confidence
	}
	return 0.3 // Default low confidence
}

// determineMatchType suggests appropriate match type based on field characteristics
func (g *Generator) determineMatchType(field EntityField) string {
	// Determine match type based on field name and type
	switch {
	case strings.Contains(field.Name, "address") || strings.Contains(field.Name, "ip"):
		return "exact" // IP addresses should match exactly
	case field.Name == "name" || field.Name == "slug":
		return "fuzzy" // Names can have fuzzy matching
	case strings.Contains(field.Name, "serial") || strings.Contains(field.Name, "mac"):
		return "exact" // Serial numbers and MAC addresses should match exactly
	case field.Type == "string":
		return "exact" // Default string matching
	case field.Type == "int32" || field.Type == "int64" || field.Type == "float" || field.Type == "double":
		return "numeric" // Numeric fields
	case field.Type == "bool":
		return "exact" // Boolean fields
	case field.IsNested:
		return "exists" // Nested objects - just check existence
	default:
		return "exact" // Default to exact matching
	}
}

// generateJSONPath creates JSON path for accessing nested fields
func (g *Generator) generateJSONPath(field EntityField) string {
	if field.IsNested {
		return fmt.Sprintf("%s.name", field.Name) // Assume nested objects have name fields
	}
	return field.Name
}
