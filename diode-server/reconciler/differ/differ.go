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
	RequestID  string `json:"request_id"`
	ObjectType string `json:"object_type"`
	Entity     any    `json:"entity"`
	State      int    `json:"state"`
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
		if obj.ObjectType() == entity.ObjectType {
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
		change := changeset.Change{
			ChangeID:           uuid.NewString(),
			ChangeType:         changeset.ChangeTypeCreate,
			ObjectType:         obj.ObjectType(),
			ObjectPrimaryValue: obj.ObjectPrimaryValue(),
			ObjectID:           nil,
			ObjectVersion:      nil,
			After:              obj.Data(),
		}

		id := obj.ID()
		if id > 0 {
			change.ObjectID = &id
			change.ChangeType = changeset.ChangeTypeUpdate
			change.Before = obj.IntendedData()
		}

		changes = append(changes, change)
	}

	deviationName := genDeviationName(objectsToReconcile)

	cs := &changeset.ChangeSet{ChangeSetID: uuid.NewString(), ChangeSet: changes, DeviationName: deviationName}
	if branchID != "" {
		cs.BranchID = &branchID
	}
	return cs, nil
}

func retrieveObjectState(ctx context.Context, netboxAPI netboxdiodeplugin.NetBoxAPI, change netbox.ComparableData, branchID string) (netbox.ComparableData, error) {
	params := netboxdiodeplugin.RetrieveObjectStateQueryParams{
		ObjectID:   0,
		ObjectType: change.ObjectType(),
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
			ObjectType:     change.ObjectType(),
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

	dw, err := netbox.NewDataWrapper(ingestEntity.ObjectType)
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

func genDeviationName(objects []netbox.ComparableData) *string {
	if len(objects) == 0 {
		return nil
	}

	primaryObject := objects[len(objects)-1]
	deviationName := fmt.Sprintf("%s %s", primaryObject.ObjectTypeName(), primaryObject.ObjectPrimaryValue())

	if primaryObject.ID() > 0 {
		deviationName += " modified"
	} else {
		deviationName += " discovered"
	}

	return &deviationName
}

func deviationNameForDiffFailure(entity IngestEntity) string {
	e, err := extractIngestEntityData(entity)
	if err != nil {
		return fmt.Sprintf("Unknown %s discovered", entity.ObjectType)
	}

	return fmt.Sprintf("%s %s discovered", e.ObjectTypeName(), e.ObjectPrimaryValue())
}

// FailedDiffChangeSet generates a placeholder change set for a failed diff
func FailedDiffChangeSet(entity IngestEntity, branchID string) *changeset.ChangeSet {
	deviationName := deviationNameForDiffFailure(entity)
	cs := &changeset.ChangeSet{
		ChangeSetID:   uuid.NewString(),
		DeviationName: &deviationName,
	}
	if branchID != "" {
		cs.BranchID = &branchID
	}
	return cs
}
