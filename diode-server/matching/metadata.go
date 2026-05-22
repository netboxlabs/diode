package matching

import (
	"reflect"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// metadataProvider is implemented by all entity types that have metadata.
type metadataProvider interface {
	GetMetadata() *structpb.Struct
}

// ExtractEntityMetadata extracts the metadata field from an entity.
// Each entity type (Device, Site, etc.) has a GetMetadata() method returning *structpb.Struct.
// Returns nil if the entity has no metadata or metadata extraction fails.
func ExtractEntityMetadata(entity *diodepb.Entity) map[string]any {
	if entity == nil || entity.GetEntity() == nil {
		return nil
	}

	// entity.GetEntity() returns a wrapper like *Entity_Device
	// which has a single field containing the actual entity (e.g., *Device)
	wrapperVal := reflect.ValueOf(entity.GetEntity()).Elem()
	if wrapperVal.NumField() == 0 {
		return nil
	}

	// Get the actual entity and check if it implements metadataProvider
	actualEntity := wrapperVal.Field(0).Interface()
	if mp, ok := actualEntity.(metadataProvider); ok {
		if meta := mp.GetMetadata(); meta != nil {
			return meta.AsMap()
		}
	}

	return nil
}

// FlattenMetadata converts a nested map to a flat map with dot-notation keys.
// Example: {"source": {"id": "123"}} -> {"source.id": "123"}
func FlattenMetadata(data map[string]any, prefix string) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nested, ok := value.(map[string]any); ok {
			for k, v := range FlattenMetadata(nested, fullKey) {
				result[k] = v
			}
		} else {
			result[fullKey] = value
		}
	}
	return result
}

// BuildNestedFilter converts a dot-notation key back to a nested map structure.
// Example: "source.id", "123" -> {"source": {"id": "123"}}
func BuildNestedFilter(key string, value any) map[string]any {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		return map[string]any{key: value}
	}

	// Build nested structure from inside out
	result := map[string]any{parts[len(parts)-1]: value}
	for i := len(parts) - 2; i >= 0; i-- {
		result = map[string]any{parts[i]: result}
	}
	return result
}
