package changeset

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/andybalholm/brotli"
)

const (
	// ChangeTypeCreate is the change type for a creation
	ChangeTypeCreate = "create"

	// ChangeTypeUpdate is the change type for an update
	ChangeTypeUpdate = "update"
)

// ChangeSet represents a change set
type ChangeSet struct {
	ChangeSetID   string   `json:"change_set_id"`
	ChangeSet     []Change `json:"change_set"`
	BranchID      *string  `json:"branch_id,omitempty"`
	DeviationName *string  `json:"deviation_name,omitempty"`
}

// Change represents a change for the change set
type Change struct {
	ChangeID           string          `json:"change_id"`
	ChangeType         string          `json:"change_type"`
	ObjectType         string          `json:"object_type"`
	ObjectPrimaryValue string          `json:"object_primary_value"`
	ObjectID           *int            `json:"object_id,omitempty"`
	ObjectVersion      *int            `json:"object_version,omitempty"`
	Before             json.RawMessage `json:"before"`
	After              json.RawMessage `json:"after"`
}

// CompressChangeSet compresses a change set
func CompressChangeSet(cs *ChangeSet) ([]byte, error) {
	csJSON, err := json.Marshal(cs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal change set JSON: %v", err)
	}

	var brotliBuf bytes.Buffer
	brotliWriter := brotli.NewWriter(&brotliBuf)
	if _, err = brotliWriter.Write(csJSON); err != nil {
		return nil, fmt.Errorf("failed to compress change set: %v", err)
	}
	if err = brotliWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to compress change set: %v", err)
	}

	return brotliBuf.Bytes(), nil
}
