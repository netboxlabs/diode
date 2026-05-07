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

	if res.ChangeSet == nil {
		return nil, fmt.Errorf("no change set returned")
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

	deviationName := genDeviationName(changes, entity.ObjectType)
	cs := &changeset.ChangeSet{ID: res.ID, Changes: changes, DeviationName: deviationName}
	if res.ChangeSet.Branch != nil {
		branchID := fmt.Sprintf("%s (%s)", res.ChangeSet.Branch.Name, res.ChangeSet.Branch.ID)
		cs.BranchID = &branchID
	}
	return cs, nil
}

// ConvertBulkPlanResult converts a single BulkPlanResult to a changeset.ChangeSet.
// Returns nil changeset if the result contains errors.
func ConvertBulkPlanResult(result netboxdiodeplugin.BulkPlanResult, objectType string) (*changeset.ChangeSet, error) {
	if len(result.Errors) > 0 && string(result.Errors) != "null" {
		return nil, fmt.Errorf("bulk plan error for entity %s: %s", result.ID, string(result.Errors))
	}

	if result.ChangeSet == nil {
		return nil, fmt.Errorf("no change set returned for entity %s", result.ID)
	}

	changes := make([]changeset.Change, 0, len(result.ChangeSet.Changes))
	for _, change := range result.ChangeSet.Changes {
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

	deviationName := genDeviationName(changes, objectType)
	cs := &changeset.ChangeSet{ID: result.ChangeSet.ID, Changes: changes, DeviationName: deviationName}
	if result.ChangeSet.Branch != nil {
		branchID := fmt.Sprintf("%s (%s)", result.ChangeSet.Branch.Name, result.ChangeSet.Branch.ID)
		cs.BranchID = &branchID
	}
	return cs, nil
}

// FindPrimaryChange attempts to find the earliest change that refers to the primary object.
func FindPrimaryChange(changes []changeset.Change, objectType string) *changeset.Change {
	// The main complexity is that there is no identifier that binds the entities to the
	// changes that are about them.  Additionally, a change may create an entity, then a later
	// change may update it (this should still be considered a discovery/create).

	firstRefID := make(map[string]*changeset.Change)
	for _, change := range changes {
		if change.ObjectType == objectType && change.RefID != nil {
			if _, ok := firstRefID[*change.RefID]; !ok {
				firstRefID[*change.RefID] = &change
			}
		}
	}

	for i := len(changes) - 1; i >= 0; i-- {
		change := &changes[i]
		if change.ObjectType == objectType {
			// if the object was created in a previous change in the change set
			if change.RefID != nil {
				// return the change that created it.
				return firstRefID[*change.RefID]
			}
			// update to a specific existing object
			return change
		}
	}
	return nil
}

func genDeviationName(changes []changeset.Change, objectType string) *string {
	typeName, err := netbox.GetObjectTypeName(objectType)
	if err != nil {
		typeName = objectType
	}

	primaryChange := FindPrimaryChange(changes, objectType)
	if primaryChange == nil {
		deviationName := fmt.Sprintf("%s unchanged", typeName)
		return &deviationName
	}

	deviationName := typeName
	if primaryChange.ObjectPrimaryValue != "" {
		deviationName += " " + primaryChange.ObjectPrimaryValue
	}

	switch primaryChange.ChangeType {
	case changeset.ChangeTypeUpdate:
		deviationName += " modified"
	case changeset.ChangeTypeCreate:
		deviationName += " discovered"
	case changeset.ChangeTypeNoop:
		// TODO: this means some subbordinate or related object was modified,
		// but the primary object itself was not modified.
		// for now we consider this a modification of the primary object.
		deviationName += " modified"
	default:
		deviationName += " (unrecognized change type " + primaryChange.ChangeType + ")"
	}

	return &deviationName
}

func deviationNameForDiffFailure(entity IngestEntity) string {
	objectType := entity.ObjectType
	objectTypeName, err := netbox.GetObjectTypeName(objectType)
	if err != nil {
		return fmt.Sprintf("Unresolved %s reported", objectType)
	}

	if e, ok := entity.Entity.(*diodepb.Entity); ok {
		primaryValue, err := netbox.GetPrimaryValue(e)
		if err != nil {
			return fmt.Sprintf("Unresolved %s reported", objectType)
		}
		return fmt.Sprintf("%s %s reported", objectTypeName, primaryValue)
	}

	return fmt.Sprintf("Unresolved %s reported", objectType)
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
