package graph

import (
	"encoding/json"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
)

// extractMetadata extracts metadata from an entity.
// Returns a JSON-encoded map of metadata key-value pairs.
//
// TODO: Add support for merging request-level metadata (IngestRequest.metadata) with
// entity-level metadata. This would require:
// 1. Pass IngestRequest.metadata to Service (via SetSourceMetadata or UpsertEntity parameter)
// 2. Merge source metadata with entity metadata (entity takes precedence for duplicate keys)
// 3. Wire up the call in IngestionProcessor.CreateIngestionLogs before UpsertEntity
func (s *Service) extractMetadata(entity *diodepb.Entity) (json.RawMessage, error) {
	if entity == nil || entity.GetEntity() == nil {
		return json.RawMessage("{}"), nil
	}

	result := matching.ExtractEntityMetadata(entity)
	if len(result) == 0 {
		return json.RawMessage("{}"), nil
	}

	return json.Marshal(result)
}

// ensureDiodeID sets source_match.diode_id to the externalID.
// This allows future lookups by diode_id to find the node directly by externalID.
// Always overwrites any client-provided diode_id to ensure consistency.
func (s *Service) ensureDiodeID(metadata json.RawMessage, externalID string) json.RawMessage {
	var metaMap map[string]any
	if err := json.Unmarshal(metadata, &metaMap); err != nil {
		s.logger.Debug("failed to unmarshal metadata, creating new map", "error", err)
		metaMap = make(map[string]any)
	}

	// Get or create source_match
	sourceMatch, ok := metaMap[sourceMatchKey].(map[string]any)
	if !ok {
		sourceMatch = make(map[string]any)
	}

	// Always set diode_id to externalID (system controls identity)
	sourceMatch[diodeIDKey] = externalID
	metaMap[sourceMatchKey] = sourceMatch

	if updated, err := json.Marshal(metaMap); err == nil {
		return updated
	}

	return metadata
}
