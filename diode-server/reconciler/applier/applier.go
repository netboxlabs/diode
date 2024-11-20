package applier

import (
	"context"
	"log/slog"

	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// ApplyChangeSet applies a change set to NetBox
func ApplyChangeSet(ctx context.Context, logger *slog.Logger, cs changeset.ChangeSet, nbClient netboxdiodeplugin.NetBoxAPI) error {
	changes := make([]netboxdiodeplugin.Change, 0)
	for _, change := range cs.ChangeSet {
		changes = append(changes, netboxdiodeplugin.Change{
			ChangeID:      change.ChangeID,
			ChangeType:    change.ChangeType,
			ObjectType:    change.ObjectType,
			ObjectID:      change.ObjectID,
			ObjectVersion: change.ObjectVersion,
			Data:          change.Data,
		})
	}

	req := netboxdiodeplugin.ChangeSetRequest{
		ChangeSetID: cs.ChangeSetID,
		ChangeSet:   changes,
	}

	resp, err := nbClient.ApplyChangeSet(ctx, req)
	if err != nil {
		return err
	}

	logger.Debug("apply change set response", "response", resp)

	return nil
}
