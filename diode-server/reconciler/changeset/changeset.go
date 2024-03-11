package changeset

import (
	"encoding/json"
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/netboxlabs/diode/diode-server/netbox"
)

const (
	// CreateChangeType is the change type for a create
	CreateChangeType = "create"

	// UpdateChangeType is the change type for an update
	UpdateChangeType = "update"
)

// IngestEntity represents an ingest entity
type IngestEntity struct {
	RequestID string `json:"request_id"`
	DataType  string `json:"data_type"`
	Data      any    `json:"data"`
	State     int    `json:"state"`
}

// ObjectState represents an object state
type ObjectState struct {
	ObjectID       int    `json:"object_id"`
	ObjectType     string `json:"object_type"`
	ObjectChangeID int    `json:"object_change_id"`
	Object         any    `json:"object"`
}

// ChangeSet represents a apply change set
type ChangeSet struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
}

// Change represents a change to apply
type Change struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"`
	ObjectType    string `json:"object_type"`
	ObjectID      *int   `json:"object_id,omitempty"`
	ObjectVersion *int   `json:"object_version,omitempty"`
	Data          any    `json:"data"`
}

// MakeChangeSet creates a change set based on ingested entities and existing object states
func MakeChangeSet(ingestEntities []IngestEntity, objectStates map[string]*ObjectState) (*ChangeSet, error) {
	var changes []Change

	for _, ingestEntity := range ingestEntities {
		objectState := objectStates[ingestEntity.RequestID]

		change, err := PrepareChange(ingestEntity, objectState)
		if err != nil {
			return nil, err
		}

		changes = append(changes, *change)
	}

	return &ChangeSet{ChangeSetID: uuid.NewString(), ChangeSet: changes}, nil
}

// PrepareChange prepares a change based on ingested entity and existing object state
func PrepareChange(ingestEntity IngestEntity, netBoxObjectState *ObjectState) (*Change, error) {
	ingestEntityData, err := extractIngestEntityData(ingestEntity)
	if err != nil {
		return nil, err
	}

	changeType := CreateChangeType

	var objectID, objectVersion *int

	changeData := ingestEntityData.Data()

	if netBoxObjectState != nil {
		changeType = UpdateChangeType
		objectID = &netBoxObjectState.ObjectID
		objectVersion = &netBoxObjectState.ObjectChangeID

		objectStateData, err := extractNetBoxObjectStateData(*netBoxObjectState)
		if err != nil {
			return nil, err
		}

		patchData, err := createPatch(objectStateData, ingestEntityData)
		if err != nil {
			return nil, err
		}

		changeData = patchData.Data()
	}

	return &Change{
		ChangeID:      uuid.NewString(),
		ChangeType:    changeType,
		ObjectType:    ingestEntity.DataType,
		ObjectID:      objectID,
		ObjectVersion: objectVersion,
		Data:          changeData,
	}, nil
}

func extractIngestEntityData(ingestEntity IngestEntity) (netbox.ComparableData, error) {
	if ingestEntity.Data == nil {
		return nil, fmt.Errorf("ingest entity data is nil")
	}

	dw, err := netbox.NewDataWrapper(ingestEntity.DataType)
	if err != nil {
		return nil, err
	}

	if err := mapstructure.Decode(ingestEntity.Data, &dw); err != nil {
		return nil, fmt.Errorf("failed to decode ingest entity data %w", err)
	}

	if !dw.IsValid() {
		return nil, fmt.Errorf("invalid ingest entity data")
	}

	return dw, nil
}

func extractNetBoxObjectStateData(obj ObjectState) (netbox.ComparableData, error) {
	if obj.Object == nil {
		return nil, fmt.Errorf("object state data is nil")
	}

	dw, err := netbox.NewDataWrapper(obj.ObjectType)
	if err != nil {
		return nil, err
	}

	if err := mapstructure.Decode(obj.Object, &dw); err != nil {
		return nil, fmt.Errorf("failed to decode object state data %w", err)
	}

	if !dw.IsValid() {
		return nil, fmt.Errorf("invalid object state data")
	}

	return dw, nil
}

func createPatch(original, modified netbox.ComparableData) (netbox.ComparableData, error) {
	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	modifiedJSON, err := json.Marshal(modified)
	if err != nil {
		return nil, err
	}

	patchJSON, err := jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
	if err != nil {
		return nil, err
	}

	patch := reflect.New(reflect.TypeOf(original)).Interface().(netbox.ComparableData)
	if err := json.Unmarshal(patchJSON, &patch); err != nil {
		return nil, err
	}

	return patch, nil
}
