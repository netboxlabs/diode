package graph

import (
	"encoding/json"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
)

// extractMetadata extracts metadata from an entity, merging with request-level
// metadata (e.g. run_id from IngestRequest.metadata). Entity-level metadata
// takes precedence over request-level metadata for duplicate keys.
func (s *Service) extractMetadata(entity *diodepb.Entity) (json.RawMessage, error) {
	// Start with request-level metadata as the base
	result := make(map[string]any)
	for k, v := range s.requestMetadata {
		result[k] = v
	}

	// Entity-level metadata takes precedence
	if entity != nil && entity.GetEntity() != nil {
		for k, v := range matching.ExtractEntityMetadata(entity) {
			result[k] = v
		}
	}

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
