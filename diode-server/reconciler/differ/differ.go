package differ

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
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

// Diff compares ingested entity with the intended state in NetBox and returns a change set
func Diff(ctx context.Context, entity IngestEntity, branchID string, netboxAPI netboxdiodeplugin.NetBoxAPI) (*changeset.ChangeSet, error) {
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
		intended, err := retrieveObjectState(ctx, netboxAPI, obj, branchID)
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

	// process objectsToReconcile and prepare change set to return
	changes := make([]changeset.Change, 0)

	for _, obj := range objectsToReconcile {
		operation := changeset.ChangeTypeCreate
		var objectID *int

		id := obj.ID()
		if id > 0 {
			objectID = &id
			operation = changeset.ChangeTypeUpdate
		}

		changes = append(changes, changeset.Change{
			ChangeID:      uuid.NewString(),
			ChangeType:    operation,
			ObjectType:    obj.DataType(),
			ObjectID:      objectID,
			ObjectVersion: nil,
			Data:          obj.Data(),
		})
	}

	cs := &changeset.ChangeSet{ChangeSetID: uuid.NewString(), ChangeSet: changes}
	if branchID != "" {
		cs.BranchID = &branchID
	}
	return cs, nil
}

func retrieveObjectState(ctx context.Context, netboxAPI netboxdiodeplugin.NetBoxAPI, change netbox.ComparableData, branchID string) (netbox.ComparableData, error) {
	params := netboxdiodeplugin.RetrieveObjectStateQueryParams{
		ObjectID:   0,
		ObjectType: change.DataType(),
		BranchID:   branchID,
		Params:     change.ObjectStateQueryParams(),
	}
	resp, err := netboxAPI.RetrieveObjectState(ctx, params)
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

	protoEntity, ok := ingestEntity.Entity.(*diodepb.Entity)
	if !ok {
		return nil, fmt.Errorf("ingest entity is not a proto entity")
	}

	if err = dw.FromProtoEntity(protoEntity); err != nil {
		return nil, err
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
