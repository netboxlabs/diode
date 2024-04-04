package changeset

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch/v5"
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
	Data      any    `json:"data"`
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
	ingested, err := extractIngestEntityData(entity)
	if err != nil {
		return nil, err
	}

	createObjectsMap := make(map[string]map[string]netbox.ComparableData)

	for _, ingestedObject := range ingested.AllData() {
		existingObject, err := retrieveObjectState(netboxAPI, ingestedObject)
		if err != nil {
			return nil, err
		}

		changesMap := make(map[string]netbox.ComparableData)
		changesMap["actual"] = ingestedObject
		changesMap["intended"] = existingObject

		createObjectsMap[ingestedObject.DataType()] = changesMap
	}

	updateObjectsMap := make(map[string]netbox.ComparableData)

	if createObjectsMap[entity.DataType]["intended"] != nil {
		for _, obj := range createObjectsMap[entity.DataType]["intended"].AllData() {
			updateObjectsMap[obj.DataType()] = obj
		}
	}

	var objectsToReconcile []netbox.ComparableData

	if len(updateObjectsMap) > 0 {
		// update existing object
		objectsToReconcileUpdate, err := objectToUpdate(ingested, createObjectsMap, updateObjectsMap)
		if err != nil {
			return nil, err
		}
		objectsToReconcile = objectsToReconcileUpdate
	} else {
		// create new object
		objectsToReconcileCreate, err := objectToCreate(ingested, createObjectsMap)
		if err != nil {
			return nil, err
		}
		objectsToReconcile = objectsToReconcileCreate
	}

	changesList := make([]Change, 0)

	for _, obj := range objectsToReconcile {

		operation := ChangeTypeCreate
		var objectID *int

		id := obj.ID()
		if id > 0 {
			objectID = &id
			operation = ChangeTypeUpdate
		}

		changesList = append(changesList, Change{
			ChangeID:      uuid.NewString(),
			ChangeType:    operation,
			ObjectType:    obj.DataType(),
			ObjectID:      objectID,
			ObjectVersion: nil,
			Data:          obj.Data(),
		})
	}

	return &ChangeSet{ChangeSetID: uuid.NewString(), ChangeSet: changesList}, nil
}

func objectToCreate(ingested netbox.ComparableData, createObjectsMap map[string]map[string]netbox.ComparableData) ([]netbox.ComparableData, error) {
	objectsToReconcile := make([]netbox.ComparableData, 0)

	for _, ingestedObject := range ingested.AllData() {
		isRootObject := ingestedObject.DataType() == ingested.DataType()

		actualObject := createObjectsMap[ingestedObject.DataType()]["actual"]
		intendedObject := createObjectsMap[ingestedObject.DataType()]["intended"]

		if intendedObject != nil {
			if !isRootObject {
				actualObject.ReplaceData(intendedObject)
				continue
			}

			patch, err := createPatch(intendedObject, actualObject)
			if err != nil {
				return nil, err
			}

			if reflect.DeepEqual(patch.Data(), intendedObject.Data()) {
				actualObject.ReplaceData(intendedObject)
				continue
			}
		} else {
			actualObject.SetDefaults()
		}

		objectsToReconcile = append(objectsToReconcile, actualObject)
	}
	return objectsToReconcile, nil
}

func objectToUpdate(ingested netbox.ComparableData, createObjectsMap map[string]map[string]netbox.ComparableData, updateObjectsMap map[string]netbox.ComparableData) ([]netbox.ComparableData, error) {
	objectsToReconcile := make([]netbox.ComparableData, 0)

	for _, ingestedObject := range ingested.AllData() {
		isRootObject := ingestedObject.DataType() == ingested.DataType()

		actualObject := createObjectsMap[ingestedObject.DataType()]["actual"]
		intendedObject := createObjectsMap[ingestedObject.DataType()]["intended"]
		currentObject := updateObjectsMap[ingestedObject.DataType()]

		isPlaceholder := actualObject.IsPlaceholder()

		if isPlaceholder {
			actualObject.ReplaceData(currentObject)
			intendedObject.ReplaceData(currentObject)
			continue
		}

		if intendedObject == nil {
			actualObject.SetDefaults()
		}

		if !isPlaceholder && !isRootObject {
			if intendedObject == nil {
				currentObject.ReplaceData(actualObject)
			} else {
				actualObject.ReplaceData(intendedObject)
				currentObject.ReplaceData(intendedObject)
			}
		}

		patch, err := createPatch(currentObject, actualObject)
		if err != nil {
			return nil, err
		}

		comparePatchWith := actualObject
		if isRootObject {
			comparePatchWith = intendedObject
		}
		updateRequired := !reflect.DeepEqual(patch.Data(), comparePatchWith.Data())

		if updateRequired {
			objectsToReconcile = append(objectsToReconcile, patch)
			continue
		}

		if isRootObject && !updateRequired && len(objectsToReconcile) > 0 {
			objectsToReconcile = append(objectsToReconcile, currentObject)
			continue
		}

		if !updateRequired && intendedObject == nil {
			objectsToReconcile = append(objectsToReconcile, currentObject)
			continue
		}
	}

	return objectsToReconcile, nil
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

	patchJSON, err := jsonpatch.MergeMergePatches(originalJSON, modifiedJSON)
	if err != nil {
		return nil, err
	}

	patch, err := cloneObject(original)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(patchJSON, &patch); err != nil {
		return nil, err
	}

	return patch, nil
}

func cloneObject(obj netbox.ComparableData) (netbox.ComparableData, error) {
	j, _ := json.Marshal(obj)

	clone, err := netbox.NewDataWrapper(obj.DataType())
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(j, &clone); err != nil {
		return nil, err
	}

	return clone, nil
}
