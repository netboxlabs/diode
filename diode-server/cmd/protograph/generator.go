package main

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"

	"github.com/netboxlabs/diode/diode-server/strcase"
)

// EntityType represents a parsed entity type from the protobuf
type EntityType struct {
	Name        string        // e.g., "Device"
	PbType      string        // e.g., "*diodepb.Device"
	OneofField  string        // e.g., "Entity_Device"
	StructField string        // e.g., "Device" (field name within the oneof struct)
	Fields      []EntityField // Parsed fields for this entity type
}

// EntityField represents a field within an entity type
type EntityField struct {
	Name       string  // Field name (e.g., "name", "serial") - snake_case from proto
	GoName     string  // Go field name (e.g., "Name", "Serial") - PascalCase
	Type       string  // Field type (e.g., "string", "int32", "Device")
	IsOptional bool    // Whether field is optional
	IsRepeated bool    // Whether field is repeated (array/slice)
	IsNested   bool    // Whether field is a nested message type
	JSONPath   string  // JSON path for accessing field (e.g., "name", "site.name")
	Confidence float64 // Generated confidence score based on field characteristics
	MatchType  string  // Suggested match type based on field type
}

// FieldTypeMapping represents a mapping from field name to entity type
type FieldTypeMapping struct {
	FieldName  string // Field name (snake_case or PascalCase)
	EntityType string // Target entity type
}

// Generator handles the code generation process
type Generator struct {
	packageName string
	templates   *template.Template
}

// NewGenerator creates a new code generator instance
func NewGenerator(packageName string) (*Generator, error) {
	funcMap := template.FuncMap{
		"snakeCase": strcase.ToSnakeCase,
		"lowerCase": strings.ToLower,
	}

	tmpl, err := template.New("entity_mappings").Funcs(funcMap).Parse(entityMappingTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Generator{
		packageName: packageName,
		templates:   tmpl,
	}, nil
}

// GenerateCode generates Go code for entity mappings
func (g *Generator) GenerateCode(entities []EntityType) (string, error) {
	// Build field-to-type mappings from all entity fields
	fieldMappings := g.buildFieldTypeMappings(entities)

	data := struct {
		Package       string
		Entities      []EntityType
		FieldMappings []FieldTypeMapping
	}{
		Package:       g.packageName,
		Entities:      entities,
		FieldMappings: fieldMappings,
	}

	var buf bytes.Buffer
	if err := g.templates.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, return the unformatted code for debugging
		return buf.String(), fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// buildFieldTypeMappings extracts field-to-entity-type mappings from all entities
func (g *Generator) buildFieldTypeMappings(entities []EntityType) []FieldTypeMapping {
	// Use a map to deduplicate mappings
	seen := make(map[string]string)

	// Build set of known entity types for validation
	knownTypes := make(map[string]bool)
	for _, e := range entities {
		knownTypes[e.Name] = true
	}

	for _, entity := range entities {
		for _, field := range entity.Fields {
			// Only include nested fields that reference known entity types
			if field.IsNested && knownTypes[field.Type] {
				// Add snake_case mapping (proto field name)
				if _, exists := seen[field.Name]; !exists {
					seen[field.Name] = field.Type
				}
				// Add PascalCase mapping (Go field name)
				if field.GoName != "" {
					if _, exists := seen[field.GoName]; !exists {
						seen[field.GoName] = field.Type
					}
				}
			}
		}
	}

	// Convert map to sorted slice for deterministic output
	mappings := make([]FieldTypeMapping, 0, len(seen))
	for fieldName, entityType := range seen {
		mappings = append(mappings, FieldTypeMapping{
			FieldName:  fieldName,
			EntityType: entityType,
		})
	}

	// Sort for deterministic output
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].FieldName < mappings[j].FieldName
	})

	return mappings
}
