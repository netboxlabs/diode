package applier_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	nbClientMock "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/applier"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

func TestApplyChangeSet(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	mockNetBoxAPI := new(nbClientMock.NetBoxAPI)
	cs := changeset.ChangeSet{
		ID: "00000000-0000-0000-0000-000000000000",
		Changes: []changeset.Change{
			{
				ID:         "00000000-0000-0000-0000-000000000001",
				ChangeType: "create",
				ObjectType: "dcim.site",
				After:      json.RawMessage(`{"name": "Site A", "slug": "site-a", "status": "active"}`),
			},
		},
		BranchID: func() *string { s := "branch_name (123)"; return &s }(),
	}

	req := netboxdiodeplugin.ApplyChangeSetRequest{
		ID: "00000000-0000-0000-0000-000000000000",
		Changes: []netboxdiodeplugin.Change{
			{
				ID:            "00000000-0000-0000-0000-000000000001",
				ChangeType:    "create",
				ObjectType:    "dcim.site",
				ObjectID:      nil,
				ObjectVersion: nil,
				Data:          json.RawMessage(`{"name": "Site A", "slug": "site-a", "status": "active"}`),
			},
		},
		BranchID: "123",
	}

	resp := &netboxdiodeplugin.ChangeSetResult{
		ChangeSet: &netboxdiodeplugin.ChangeSet{
			ID: "00000000-0000-0000-0000-000000000000",
		},
	}

	mockNetBoxAPI.On("ApplyChangeSet", ctx, req).Return(resp, nil)

	err := applier.ApplyChangeSet(ctx, logger, cs, mockNetBoxAPI)
	assert.NoError(t, err)
	mockNetBoxAPI.AssertExpectations(t)
}
