package changeset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
)

const (
	// ChangeTypeCreate is the change type for a creation
	ChangeTypeCreate = "create"

	// ChangeTypeUpdate is the change type for an update
	ChangeTypeUpdate = "update"
)

// IngestEntity represents an ingest entity
type IngestEntity struct {
	RequestID string `json:"request_id"`
	DataType  string `json:"data_type"`
	Entity    any    `json:"entity"`
	State     int    `json:"state"`
}

// ObjectState represents a object state
type ObjectState struct {
	ObjectID       int    `json:"object_id"`
	ObjectType     string `json:"object_type"`
	ObjectChangeID int    `json:"object_change_id"`
	Object         any    `json:"object"`
}

// ChangeSet represents a change set
type ChangeSet struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
}

// Change represents a change for the change set
type Change struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"`
	ObjectType    string `json:"object_type"`
	ObjectID      *int   `json:"object_id,omitempty"`
	ObjectVersion *int   `json:"object_version,omitempty"`
	Data          any    `json:"data"`
}

// Prepare prepares a change set
func Prepare(entity IngestEntity, netboxAPI netboxdiodeplugin.NetBoxAPI) (*ChangeSet, error) {
	// extract ingested entity (actual)
	actual, err := extractIngestEntityData(entity)
	if err != nil {
		return nil, err
	}

	// get root object and all its nested objects (actual)
	actualNestedObjects, err := actual.NestedObjects()
	if err != nil {
		return nil, err
	}

	// map out root object and all its nested objects (actual)
	actualNestedObjectsMap := make(map[string]netbox.ComparableData)
	for _, obj := range actualNestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	// retrieve root object all its nested objects from NetBox (intended)
	intendedNestedObjectsMap := make(map[string]netbox.ComparableData)
	for _, obj := range actualNestedObjects {
		intended, err := retrieveObjectState(netboxAPI, obj)
		if err != nil {
			return nil, err
		}
		intendedNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = intended
	}

	// map out retrieved root object and all its nested objects (current)
	var current netbox.ComparableData
	for _, obj := range actualNestedObjects {
		if obj.DataType() == entity.DataType {
			current = intendedNestedObjectsMap[fmt.Sprintf("%p", obj.Data())]
			break
		}
	}

	objectsToReconcile, err := actual.Patch(current, intendedNestedObjectsMap)
	if err != nil {
		return nil, err
	}

	// process objectsToReconcile and prepare changeset to return
	changes := make([]Change, 0)

	for _, obj := range objectsToReconcile {
		operation := ChangeTypeCreate
		var objectID *int

		id := obj.ID()
		if id > 0 {
			objectID = &id
			operation = ChangeTypeUpdate
		}

		changes = append(changes, Change{
			ChangeID:      uuid.NewString(),
			ChangeType:    operation,
			ObjectType:    obj.DataType(),
			ObjectID:      objectID,
			ObjectVersion: nil,
			Data:          obj.Data(),
		})
	}

	return &ChangeSet{ChangeSetID: uuid.NewString(), ChangeSet: changes}, nil
}

func retrieveObjectState(netboxAPI netboxdiodeplugin.NetBoxAPI, change netbox.ComparableData) (netbox.ComparableData, error) {
	resp, err := netboxAPI.RetrieveObjectState(context.Background(), change.DataType(), 0, change.QueryString())
	if err != nil {
		return nil, err
	}

	if resp.Object.IsValid() {
		objectState := &ObjectState{
			ObjectID:       resp.ObjectID,
			ObjectType:     change.DataType(),
			ObjectChangeID: resp.ObjectChangeID,
			Object:         resp.Object,
		}

		return extractNetBoxObjectStateData(*objectState)
	}

	return nil, nil
}

func extractIngestEntityData(ingestEntity IngestEntity) (netbox.ComparableData, error) {
	if ingestEntity.Entity == nil {
		return nil, fmt.Errorf("ingest entity is nil")
	}

	dw, err := netbox.NewDataWrapper(ingestEntity.DataType)
	if err != nil {
		return nil, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &dw,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			netbox.IpamIPAddressAssignedObjectHookFunc(),
		),
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(ingestEntity.Entity); err != nil {
		return nil, fmt.Errorf("failed to decode ingest entity %w", err)
	}

	if !dw.IsValid() {
		return nil, fmt.Errorf("invalid ingest entity")
	}

	return dw, nil
}

func extractNetBoxObjectStateData(obj ObjectState) (netbox.ComparableData, error) {
	if obj.Object == nil {
		return nil, fmt.Errorf("object state is nil")
	}

	dw, err := netbox.NewDataWrapper(obj.ObjectType)
	if err != nil {
		return nil, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:    &dw,
		MatchName: netbox.IpamIPAddressAssignedObjectMatchName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			netbox.IpamIPAddressAssignedObjectHookFunc(),
		),
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(obj.Object); err != nil {
		return nil, fmt.Errorf("failed to decode object entity %w", err)
	}

	if !dw.IsValid() {
		return nil, fmt.Errorf("invalid object state")
	}

	dw.Normalise()

	return dw, nil
}
