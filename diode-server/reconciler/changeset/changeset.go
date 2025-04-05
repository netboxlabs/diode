package changeset

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/andybalholm/brotli"

	"github.com/netboxlabs/diode/diode-server/errors"
)

const (
	// ChangeTypeCreate is the change type for a creation
	ChangeTypeCreate = "create"

	// ChangeTypeUpdate is the change type for an update
	ChangeTypeUpdate = "update"

	// ChangeTypeNoop is the change type for a no-op
	ChangeTypeNoop = "noop"
)

// ChangeSet represents a change set
type ChangeSet struct {
	ID            string   `json:"id"`
	Changes       []Change `json:"changes"`
	BranchID      *string  `json:"branch_id,omitempty"`
	DeviationName *string  `json:"deviation_name,omitempty"`
}

// Change represents a change for the change set
type Change struct {
	ID                 string          `json:"id"`
	ChangeType         string          `json:"change_type"`
	ObjectType         string          `json:"object_type"`
	ObjectPrimaryValue string          `json:"object_primary_value"`
	ObjectID           *int            `json:"object_id,omitempty"`
	RefID              *string         `json:"ref_id,omitempty"`
	ObjectVersion      *int            `json:"object_version,omitempty"`
	Before             json.RawMessage `json:"before"`
	After              json.RawMessage `json:"after"`
	NewRefs            []string        `json:"new_refs,omitempty"`
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

// Error represents an error when diffing or applying change set
type Error struct {
	Message string           `json:"message"`
	Code    errors.ErrorCode `json:"code"`
	Details json.RawMessage  `json:"details"`
}

// Error returns the ChangeSetError message
func (e *Error) Error() string {
	return fmt.Sprintf("%s - %s - %s", e.Message, e.Code, string(e.Details))
}

// NewError creates a new Error
func NewError(message string, code errors.ErrorCode, details []byte) error {
	return &Error{
		Message: message,
		Code:    code,
		Details: details,
	}
}
