package applier

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	diodeErrors "github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// Compile the regex once at package level for better performance
var branchIDRegex = regexp.MustCompile(`^.*\((.*)\)$`)

func toApplyRequest(cs changeset.ChangeSet) netboxdiodeplugin.ApplyChangeSetRequest {
	changes := make([]netboxdiodeplugin.Change, 0, len(cs.Changes))
	for _, change := range cs.Changes {
		changes = append(changes, netboxdiodeplugin.Change{
			ID:                 change.ID,
			ChangeType:         change.ChangeType,
			ObjectType:         change.ObjectType,
			ObjectID:           change.ObjectID,
			ObjectVersion:      change.ObjectVersion,
			ObjectPrimaryValue: change.ObjectPrimaryValue,
			RefID:              change.RefID,
			Data:               change.After,
			NewRefs:            change.NewRefs,
		})
	}

	req := netboxdiodeplugin.ApplyChangeSetRequest{
		ID:      cs.ID,
		Changes: changes,
	}
	if cs.BranchID != nil {
		branchIDStr := *cs.BranchID
		if matches := branchIDRegex.FindStringSubmatch(branchIDStr); len(matches) > 1 {
			req.BranchID = strings.TrimSpace(matches[1])
		} else {
			req.BranchID = branchIDStr
		}
	}
	return req
}

// ApplyChangeSet applies a change set to NetBox
func ApplyChangeSet(ctx context.Context, logger *slog.Logger, cs changeset.ChangeSet, nbClient netboxdiodeplugin.NetBoxAPI) error {
	req := toApplyRequest(cs)

	resp, err := nbClient.ApplyChangeSet(ctx, req)
	if err != nil {
		return err
	}

	logger.Debug("apply change set response", "response", resp)

	return nil
}

// BulkApplyResult holds the outcome of applying one change set within a bulk call.
type BulkApplyResult struct {
	Index int
	Err   error
}

// BulkApplyChangeSets applies multiple change sets in a single HTTP call.
// Returns one BulkApplyResult per input change set, preserving order.
func BulkApplyChangeSets(ctx context.Context, logger *slog.Logger, changeSets []changeset.ChangeSet, branchID string, nbClient netboxdiodeplugin.NetBoxAPI) []BulkApplyResult {
	results := make([]BulkApplyResult, len(changeSets))

	requests := make([]netboxdiodeplugin.ApplyChangeSetRequest, len(changeSets))
	for i, cs := range changeSets {
		requests[i] = toApplyRequest(cs)
	}

	resp, err := nbClient.BulkApply(ctx, netboxdiodeplugin.BulkApplyRequest{
		ChangeSets: requests,
		BranchID:   branchID,
	})
	if err != nil {
		for i := range results {
			results[i] = BulkApplyResult{Index: i, Err: err}
		}
		return results
	}

	for i := range changeSets {
		results[i].Index = i
		if i >= len(resp.Results) {
			results[i].Err = fmt.Errorf("no result returned for change set at index %d", i)
			continue
		}
		r := resp.Results[i]
		if len(r.Errors) > 0 && string(r.Errors) != "null" {
			results[i].Err = changeset.NewError("apply failed", diodeErrors.ErrCodeOpsApplyChangeSet, r.Errors)
		}
	}

	logger.Debug("bulk apply complete", "total", len(changeSets), "results", len(resp.Results))
	return results
}
