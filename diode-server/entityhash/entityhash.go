package entityhash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/gowebpki/jcs"
	"google.golang.org/protobuf/encoding/protojson"

	diodepb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// EntityFingerprinter provides methods for generating consistent hashes of entities
type EntityFingerprinter struct {
	marshaler protojson.MarshalOptions
}

// NewEntityFingerprinter creates a new entity fingerprinter with consistent options
func NewEntityFingerprinter() *EntityFingerprinter {
	return &EntityFingerprinter{
		marshaler: protojson.MarshalOptions{
			UseProtoNames:   true,  // Use snake_case field names for consistency
			EmitUnpopulated: false, // Don't include zero/empty values
		},
	}
}

// GenerateEntityHash creates a SHA256 hash for an entity that includes the object type
// and canonicalized entity content, excluding the timestamp from the outer envelope
func (f *EntityFingerprinter) GenerateEntityHash(entity *diodepb.Entity) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("entity cannot be nil")
	}

	// Extract the inner entity content (excluding timestamp)
	entityContent := entity.GetEntity()
	if entityContent == nil {
		return "", fmt.Errorf("entity content cannot be nil")
	}

	// Serialize the inner entity content to JSON (excluding timestamp)
	entityJSON, err := f.marshaler.Marshal(&diodepb.Entity{Entity: entityContent})
	if err != nil {
		return "", fmt.Errorf("failed to marshal entity: %w", err)
	}

	return f.GenerateEntityHashFromJSON(entityJSON)
}

// GenerateEntityHashFromJSON creates a SHA256 hash for an entity.
// This should already exclude the timestamp from the entity envelope.
func (f *EntityFingerprinter) GenerateEntityHashFromJSON(entityJSON []byte) (string, error) {
	// Canonicalize the JSON using RFC 8785 JCS
	canonicalJSON, err := jcs.Transform(entityJSON)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize JSON: %w", err)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(canonicalJSON)
	return hex.EncodeToString(hash[:]), nil
}
