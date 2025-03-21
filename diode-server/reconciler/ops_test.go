package reconciler_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	pb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	pluginmocks "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestProOpsGenerateChangeSet(t *testing.T) {
	type mockGenerateDiff struct {
		changeSet []netboxdiodeplugin.Change
		branchID  string
		err       error
	}

	type mockCreateChangeSet struct {
		ingestionLogDBID int32
		deviationName    *string
		id               int32
	}

	type mockUpdateIngestionLogStateWithError struct {
		ingestionLogDBID int32
		state            pb.State
		err              error
	}

	tests := []struct {
		name string

		logDBID  int32
		log      *pb.IngestionLog
		branchID string

		generateDiff                      []mockGenerateDiff
		createChangeSets                  []mockCreateChangeSet
		updateIngestionLogStateWithErrors []mockUpdateIngestionLogStateWithError

		errorMessage string
		hasError     bool
	}{
		{
			name:    "diff failure generates a placeholder change with deviation type",
			logDBID: 1234,
			log: &pb.IngestionLog{
				Id:         "8a8ae517-85b9-466e-890c-aadb0771cc9e",
				ObjectType: netbox.DcimSiteObjectType,
				State:      pb.State_QUEUED,
				RequestId:  "1abf059c-496f-4037-83c2-0e9b1d021e85",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-1",
						},
					},
				},
			},
			generateDiff: []mockGenerateDiff{
				{
					err: fmt.Errorf("Client.Timeout exceeded while awaiting headers"),
				},
			},
			updateIngestionLogStateWithErrors: []mockUpdateIngestionLogStateWithError{
				{
					ingestionLogDBID: 1234,
					state:            pb.State_FAILED,
				},
			},
			createChangeSets: []mockCreateChangeSet{
				{
					ingestionLogDBID: 1234,
					deviationName:    strPtr("Site test-site-1 discovered"),
					id:               1235,
				},
			},
			hasError:     true,
			errorMessage: "Client.Timeout exceeded while awaiting headers",
		},
		{
			name:     "placeholder change reflects branch",
			branchID: "branch-1",
			logDBID:  1234,
			log: &pb.IngestionLog{
				Id:         "8a8ae517-85b9-466e-890c-aadb0771cc9e",
				ObjectType: netbox.DcimSiteObjectType,
				State:      pb.State_QUEUED,
				RequestId:  "1abf059c-496f-4037-83c2-0e9b1d021e85",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-1",
						},
					},
				},
			},
			generateDiff: []mockGenerateDiff{
				{
					err: fmt.Errorf("Client.Timeout exceeded while awaiting headers"),
				},
			},
			updateIngestionLogStateWithErrors: []mockUpdateIngestionLogStateWithError{
				{
					ingestionLogDBID: 1234,
					state:            pb.State_FAILED,
				},
			},
			createChangeSets: []mockCreateChangeSet{
				{
					ingestionLogDBID: 1234,
					deviationName:    strPtr("Site test-site-1 discovered"),
					id:               1235,
				},
			},
			hasError:     true,
			errorMessage: "Client.Timeout exceeded while awaiting headers",
		},
	}

	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository := mocks.NewRepository(t)
			mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
			ops := reconciler.NewOps(mockRepository, mockNetBoxClient, logger)

			for _, m := range tt.generateDiff {
				if m.err == nil {
					mockNetBoxClient.EXPECT().GenerateDiff(ctx, mock.Anything).Return(&netboxdiodeplugin.GenerateDiffResponse{
						ChangeSet: m.changeSet,
						Branch: &netboxdiodeplugin.Branch{
							ID: m.branchID,
						},
					}, nil)
				} else {
					mockNetBoxClient.EXPECT().GenerateDiff(ctx, mock.Anything).Return(nil, m.err)
				}
			}
			for _, m := range tt.createChangeSets {
				mockRepository.EXPECT().CreateChangeSet(ctx, mock.MatchedBy(func(c changeset.ChangeSet) bool {
					if !strPtrEq(c.DeviationName, m.deviationName) {
						return false
					}
					if tt.branchID != "" && !strPtrEq(c.BranchID, &tt.branchID) {
						return false
					}
					return true
				}), m.ingestionLogDBID).Return(&m.id, nil)
			}
			for _, m := range tt.updateIngestionLogStateWithErrors {
				mockRepository.EXPECT().UpdateIngestionLogStateWithError(ctx, m.ingestionLogDBID, m.state, mock.Anything).Return(m.err)
			}

			csid, cs, err := ops.GenerateChangeSet(ctx, tt.logDBID, tt.log, tt.branchID)
			if tt.hasError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, csid)
				require.NotNil(t, cs)
				// TODO(ltucker): positive tests
			}
		})
	}
}

func strPtrEq(a *string, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func strPtr(s string) *string {
	return &s
}
