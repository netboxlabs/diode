package differ

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// IngestEntity represents an ingest entity
type IngestEntity struct {
	RequestID  string        `json:"request_id"`
	ObjectType string        `json:"object_type"`
	Entity     proto.Message `json:"entity"`
	State      int           `json:"state"`
}

// Diff compares ingested entity with the intended state in NetBox and returns a change set
func Diff(ctx context.Context, entity IngestEntity, branchID string, netboxAPI netboxdiodeplugin.NetBoxAPI) (*changeset.ChangeSet, error) {
	req := netboxdiodeplugin.GenerateDiffRequest{
		ObjectType: entity.ObjectType,
		BranchID:   branchID,
		Entity:     entity.Entity,
	}

	res, err := netboxAPI.GenerateDiff(ctx, req)
	if err != nil {
		return nil, err
	}

	changes := make([]changeset.Change, 0)
	for _, change := range res.ChangeSet.Changes {
		changes = append(changes, changeset.Change{
			ID:                 change.ID,
			ChangeType:         change.ChangeType,
			ObjectType:         change.ObjectType,
			ObjectID:           change.ObjectID,
			ObjectVersion:      change.ObjectVersion,
			RefID:              change.RefID,
			ObjectPrimaryValue: change.ObjectPrimaryValue,
			After:              change.Data,
			Before:             change.Before,
			NewRefs:            change.NewRefs,
		})
	}

	deviationName := genDeviationName(changes)
	cs := &changeset.ChangeSet{ID: res.ID, Changes: changes, DeviationName: deviationName}
	if res.ChangeSet.Branch != nil {
		branchID := fmt.Sprintf("%s (%s)", res.ChangeSet.Branch.Name, res.ChangeSet.Branch.ID)
		cs.BranchID = &branchID
	}
	return cs, nil
}

func genDeviationName(objects []changeset.Change) *string {
	if len(objects) == 0 {
		deviationName := fmt.Sprintf("%s unchanged", typeName)
		return &deviationName
	}

	primaryObject := objects[len(objects)-1]
	objectTypeName, err := netbox.GetObjectTypeName(primaryObject.ObjectType)
	if err != nil {
		objectTypeName = "<unrecognized type " + primaryObject.ObjectType + ">"
	}

	deviationName := fmt.Sprintf("%s %s", objectTypeName, primaryObject.ObjectPrimaryValue)

	if primaryObject.ChangeType == changeset.ChangeTypeUpdate {
		deviationName += " modified"
	} else if primaryObject.ChangeType == changeset.ChangeTypeCreate {
		deviationName += " created"
	} else {
		deviationName += " (unrecognized change type " + primaryObject.ChangeType + ""
	}

	return &deviationName
}

func deviationNameForDiffFailure(entity IngestEntity) string {
	objectType := entity.ObjectType
	objectTypeName, err := netbox.GetObjectTypeName(objectType)
	if err != nil {
		return fmt.Sprintf("Unknown %s discovered", objectType)
	}

	if e, ok := entity.Entity.(*diodepb.Entity); ok {
		primaryValue, err := netbox.GetPrimaryValue(e)
		if err != nil {
			return fmt.Sprintf("Unknown %s discovered", objectType)
		}
		return fmt.Sprintf("%s %s discovered", objectTypeName, primaryValue)
	}

	return fmt.Sprintf("Unknown %s discovered", objectType)
}

// FailedDiffChangeSet generates a placeholder change set for a failed diff
func FailedDiffChangeSet(entity IngestEntity, branchID string) *changeset.ChangeSet {
	deviationName := deviationNameForDiffFailure(entity)
	cs := &changeset.ChangeSet{
		ID:            uuid.NewString(),
		DeviationName: &deviationName,
	}
	if branchID != "" {
		cs.BranchID = &branchID
	}
	return cs
}
