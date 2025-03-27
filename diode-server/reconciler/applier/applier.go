package applier

import (
	"context"
	"log/slog"
	"regexp"
	"strings"

	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// Compile the regex once at package level for better performance
var branchIDRegex = regexp.MustCompile(`^.*\((.*)\)$`)

// ApplyChangeSet applies a change set to NetBox
func ApplyChangeSet(ctx context.Context, logger *slog.Logger, cs changeset.ChangeSet, nbClient netboxdiodeplugin.NetBoxAPI) error {
	changes := make([]netboxdiodeplugin.Change, 0)
	for _, change := range cs.Changes {
		changes = append(changes, netboxdiodeplugin.Change{
			ID:            change.ID,
			ChangeType:    change.ChangeType,
			ObjectType:    change.ObjectType,
			ObjectID:      change.ObjectID,
			ObjectVersion: change.ObjectVersion,
			Data:          change.After,
		})
	}

	req := netboxdiodeplugin.ApplyChangeSetRequest{
		ID:      cs.ID,
		Changes: changes,
	}
	if cs.BranchID != nil {
		branchIDStr := *cs.BranchID
		// Check if the branch ID is in the format "branch_name (branch_id)"
		if matches := branchIDRegex.FindStringSubmatch(branchIDStr); len(matches) > 1 {
			// Extract the branch_id from within the parentheses (captured group)
			req.BranchID = strings.TrimSpace(matches[1])
		} else {
			// Use the original branch ID as is
			req.BranchID = branchIDStr
		}
	}

	resp, err := nbClient.ApplyChangeSet(ctx, req)
	if err != nil {
		return err
	}

	logger.Debug("apply change set response", "response", resp)

	return nil
}
