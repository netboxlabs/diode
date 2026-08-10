// Protograph generates Go code for entity type mappings from protobuf definitions.
// This eliminates manual switch statements and ensures complete entity coverage.
//
// Usage:
//
//	go run ./cmd/protograph -proto=../../diode-proto/diode/v1/ingester.proto -output=reconciler/entity_mappings_generated.go
//	make gen-entity-mappings
//	go generate ./reconciler/
//
// Generated Functions:
//   - CreateEntityFromInterface() - Convert interface{} to *diodepb.Entity
//   - GetEntityTypeName() - Get entity type name for logging
//   - IsKnownEntityType() - Validate entity types
//   - GetAllEntityTypes() - List all supported entity types
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		protoFile   = flag.String("proto", "", "Path to the protobuf file")
		outputFile  = flag.String("output", "gen/protograph/entity_mappings.go", "Output file name")
		packageName = flag.String("package", "protograph", "Go package name for generated code")
	)
	flag.Parse()

	if *protoFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -proto=<protobuf_file> [-output=<output_file>] [-package=<package_name>]\n", os.Args[0])
		os.Exit(1)
	}

	// Resolve absolute paths
	protoPath, err := filepath.Abs(*protoFile)
	if err != nil {
		log.Fatalf("Failed to resolve proto file path: %v", err)
	}

	outputPath, err := filepath.Abs(*outputFile)
	if err != nil {
		log.Fatalf("Failed to resolve output file path: %v", err)
	}

	fmt.Printf("Generating entity mappings from %s to %s\n", protoPath, outputPath)

	// Initialize the generator
	generator, err := NewGenerator(*packageName)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Parse the protobuf file and extract entity types
	entities, err := generator.ParseProtobuf(protoPath)
	if err != nil {
		log.Fatalf("Failed to parse protobuf file: %v", err)
	}

	fmt.Printf("Found %d entity types\n", len(entities))

	// Generate the Go code
	code, err := generator.GenerateCode(entities)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	// Write the generated code to file
	if err := os.WriteFile(outputPath, []byte(code), 0o644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Successfully generated entity mappings in %s\n", outputPath)
}
