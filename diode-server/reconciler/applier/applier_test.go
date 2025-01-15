package applier_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/netbox"
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
		ChangeSetID: "00000000-0000-0000-0000-000000000000",
		ChangeSet: []changeset.Change{
			{
				ChangeID:   "00000000-0000-0000-0000-000000000001",
				ChangeType: "create",
				ObjectType: "dcim.site",
				After: &netbox.DcimSite{
					Name:   "Site A",
					Slug:   "site-a",
					Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
				},
			},
		},
	}

	req := netboxdiodeplugin.ChangeSetRequest{
		ChangeSetID: "00000000-0000-0000-0000-000000000000",
		ChangeSet: []netboxdiodeplugin.Change{
			{
				ChangeID:      "00000000-0000-0000-0000-000000000001",
				ChangeType:    "create",
				ObjectType:    "dcim.site",
				ObjectID:      nil,
				ObjectVersion: nil,
				Data: &netbox.DcimSite{
					Name:   "Site A",
					Slug:   "site-a",
					Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
				},
			},
		},
	}

	resp := &netboxdiodeplugin.ChangeSetResponse{
		ChangeSetID: "00000000-0000-0000-0000-000000000000",
		Result:      "success",
	}

	mockNetBoxAPI.On("ApplyChangeSet", ctx, req).Return(resp, nil)

	err := applier.ApplyChangeSet(ctx, logger, cs, mockNetBoxAPI)
	assert.NoError(t, err)
	mockNetBoxAPI.AssertExpectations(t)
}

func strPtr(s string) *string { return &s }
