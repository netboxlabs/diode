package main

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
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
	Name       string  // Field name (e.g., "name", "serial")
	Type       string  // Field type (e.g., "string", "int32", "Device")
	IsOptional bool    // Whether field is optional
	IsRepeated bool    // Whether field is repeated (array/slice)
	IsNested   bool    // Whether field is a nested message type
	JSONPath   string  // JSON path for accessing field (e.g., "name", "site.name")
	Confidence float64 // Generated confidence score based on field characteristics
	MatchType  string  // Suggested match type based on field type
}

// Generator handles the code generation process
type Generator struct {
	packageName string
	templates   *template.Template
}

// NewGenerator creates a new code generator instance
func NewGenerator(packageName string) (*Generator, error) {
	tmpl, err := template.New("entity_mappings").Parse(entityMappingTemplate)
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
	data := struct {
		Package  string
		Entities []EntityType
	}{
		Package:  g.packageName,
		Entities: entities,
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
