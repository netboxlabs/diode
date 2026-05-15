package reconciler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	pb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/gen/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	pluginmocks "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestOpsGenerateChangeSet(t *testing.T) {
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
				ObjectType: netbox.SiteObjectType,
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
					deviationName:    strPtr("Site test-site-1 reported"),
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
				ObjectType: netbox.SiteObjectType,
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
					deviationName:    strPtr("Site test-site-1 reported"),
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
			ops := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

			for _, m := range tt.generateDiff {
				if m.err == nil {
					mockNetBoxClient.EXPECT().GenerateDiff(ctx, mock.Anything).Return(&netboxdiodeplugin.ChangeSetResult{
						ChangeSet: &netboxdiodeplugin.ChangeSet{
							Changes: m.changeSet,
							Branch: &netboxdiodeplugin.Branch{
								ID: m.branchID,
							},
						},
					}, nil)
				} else {
					mockNetBoxClient.EXPECT().GenerateDiff(ctx, mock.Anything).Return(nil, m.err)
				}
			}
			for _, m := range tt.createChangeSets {
				mockRepository.EXPECT().CreateChangeSet(ctx, mock.MatchedBy(func(c changeset.ChangeSet) bool {
					if !strPtrEq(c.DeviationName, m.deviationName) {
						t.Logf("deviation name mismatch: %v != %v", *c.DeviationName, *m.deviationName)
						return false
					}
					if tt.branchID != "" && !strPtrEq(c.BranchID, &tt.branchID) {
						return false
					}
					return true
				}), m.ingestionLogDBID).Return(&m.id, nil)
			}
			for _, m := range tt.updateIngestionLogStateWithErrors {
				mockRepository.EXPECT().UpdateIngestionLogStateWithError(ctx, m.ingestionLogDBID, m.state, mock.Anything).Run(func(_ context.Context, _ int32, _ pb.State, err error) {
					// the error given must marshal to JSON for storage in the database
					_, jsonErr := json.Marshal(err)
					require.NoError(t, jsonErr)
				}).Return(m.err)
			}

			csid, cs, err := ops.GenerateChangeSet(ctx, tt.logDBID, tt.log, tt.branchID)
			if tt.hasError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMessage)
				if cse, ok := err.(*changeset.Error); ok {
					require.True(t, json.Valid(cse.Details))
				}
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

func TestOpsCreateIngestionLog(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{
				Name: "test-site-1",
			},
		},
	}

	testIngestionLog := &pb.IngestionLog{
		Id:         "8a8ae517-85b9-466e-890c-aadb0771cc9e",
		ObjectType: netbox.SiteObjectType,
		State:      pb.State_QUEUED,
		RequestId:  "1abf059c-496f-4037-83c2-0e9b1d021e85",
		Entity:     testEntity,
	}

	testSourceMetadata := []byte(`{"source": "test"}`)

	type mockCreateIngestionLog struct {
		id    *int32
		error error
	}

	tests := []struct {
		name string

		ingestionLog   *pb.IngestionLog
		sourceMetadata []byte

		// Mock expectations
		mockCreateIngestionLog *mockCreateIngestionLog

		mockFindPriorIngestionLogID    *int32
		mockFindPriorIngestionLog      *pb.IngestionLog
		mockFindPriorIngestionLogError error

		expectedError      string
		expectWasDuplicate bool
	}{
		{
			name:           "no duplicate found - successful creation",
			ingestionLog:   testIngestionLog,
			sourceMetadata: testSourceMetadata,

			mockCreateIngestionLog: &mockCreateIngestionLog{
				id:    int32Ptr(1234),
				error: nil,
			},
			mockFindPriorIngestionLogError: sql.ErrNoRows,

			expectedError:      "",
			expectWasDuplicate: false,
		},
		{
			name:           "duplicate found",
			ingestionLog:   testIngestionLog,
			sourceMetadata: testSourceMetadata,

			mockFindPriorIngestionLogID:    int32Ptr(5678),
			mockFindPriorIngestionLog:      testIngestionLog,
			mockFindPriorIngestionLogError: nil,

			expectedError:      "",
			expectWasDuplicate: true,
		},
		{
			name:           "create ingestion log fails",
			ingestionLog:   testIngestionLog,
			sourceMetadata: testSourceMetadata,

			mockCreateIngestionLog: &mockCreateIngestionLog{
				error: fmt.Errorf("database error"),
			},

			expectedError:      "database error",
			expectWasDuplicate: false,
		},
		{
			name:           "duplicate search fails with non-NoRows error",
			ingestionLog:   testIngestionLog,
			sourceMetadata: testSourceMetadata,

			mockFindPriorIngestionLogError: fmt.Errorf("database connection error"),

			expectedError: "failed to search for prior deviation: database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository := mocks.NewRepository(t)
			mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
			opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

			// GetDefaultBranch is only called by the background refresher
			// (started via ops.Start). This test does not start the refresher,
			// so DefaultBranch reads a cold cache and returns (nil, nil) —
			// matching the previous "nil branch" expectation here without
			// needing a mock call.

			mockRepository.EXPECT().FindPriorIngestionLogByEntityHash(mock.Anything, mock.AnythingOfType("string"), (*string)(nil)).
				Return(tt.mockFindPriorIngestionLogID, tt.mockFindPriorIngestionLog, tt.mockFindPriorIngestionLogError)

			// Mock CreateIngestionLog
			if tt.mockCreateIngestionLog != nil {
				mockRepository.EXPECT().CreateIngestionLog(mock.Anything, tt.ingestionLog, tt.sourceMetadata, mock.AnythingOfType("string")).
					Return(tt.mockCreateIngestionLog.id, tt.mockCreateIngestionLog.error)
			}

			if tt.expectWasDuplicate {
				mockRepository.EXPECT().IncrementDuplicateCount(mock.Anything, *tt.mockFindPriorIngestionLogID).Return(nil)
			}

			result, err := opsInstance.CreateIngestionLog(ctx, tt.ingestionLog, tt.sourceMetadata)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.IngestionLog)
			if tt.mockCreateIngestionLog != nil {
				require.Equal(t, *tt.mockCreateIngestionLog.id, result.ID)
			}
			require.Equal(t, tt.ingestionLog, result.IngestionLog)
			require.Equal(t, tt.expectWasDuplicate, result.WasDuplicate)
		})
	}
}

// waitForBranch polls DefaultBranch until predicate returns true or timeout
// fires. Returns the last value seen. Used so tests don't depend on the
// refresher goroutine's exact scheduling.
func waitForBranch(t *testing.T, ops *reconciler.Ops, predicate func(*netboxdiodeplugin.Branch) bool) *netboxdiodeplugin.Branch {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		b, _ := ops.DefaultBranch(context.Background())
		if predicate(b) {
			return b
		}
		time.Sleep(10 * time.Millisecond)
	}
	b, _ := ops.DefaultBranch(context.Background())
	return b
}

func TestOpsDefaultBranchColdCache(t *testing.T) {
	// Without Start, the refresher never runs; DefaultBranch returns (nil, nil).
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	branch, err := opsInstance.DefaultBranch(context.Background())
	require.NoError(t, err)
	require.Nil(t, branch)
	require.False(t, opsInstance.HasBranchLoaded())
}

func TestOpsBranchRefresherSeedsCache(t *testing.T) {
	// After Start, the refresher fetches once and DefaultBranch returns the
	// fetched value without doing HTTP itself.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	want := &netboxdiodeplugin.Branch{ID: "branch-1", Name: "Branch One"}
	mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).Return(want, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opsInstance.Start(ctx)

	got := waitForBranch(t, opsInstance, func(b *netboxdiodeplugin.Branch) bool { return b != nil })
	require.Equal(t, want, got)
	require.True(t, opsInstance.HasBranchLoaded())
}

func TestOpsBranchRefresher404IsRecorded(t *testing.T) {
	// ErrDefaultBranchNotFound resolves the cache to nil (known-absent) — not
	// a transient error, so we record it instead of leaving the cache cold.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).Return((*netboxdiodeplugin.Branch)(nil), netboxdiodeplugin.ErrDefaultBranchNotFound)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opsInstance.Start(ctx)

	// Wait until the refresher has recorded ANY result.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !opsInstance.HasBranchLoaded() {
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, opsInstance.HasBranchLoaded())

	branch, err := opsInstance.DefaultBranch(context.Background())
	require.NoError(t, err)
	require.Nil(t, branch)
}

func TestOpsBranchRefresherKeepsLastOnError(t *testing.T) {
	// A transient error during refresh leaves the previous successful value
	// in place — the whole point of decoupling consume loop from NetBox auth.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	good := &netboxdiodeplugin.Branch{ID: "branch-1", Name: "Branch One"}
	// First call (refresher start): success.
	firstCall := mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).Return(good, nil).Once()
	// Second call (triggered via RefreshDefaultBranch): error.
	mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).Return((*netboxdiodeplugin.Branch)(nil), fmt.Errorf("transient network error")).Maybe().NotBefore(firstCall)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opsInstance.Start(ctx)

	got := waitForBranch(t, opsInstance, func(b *netboxdiodeplugin.Branch) bool { return b != nil })
	require.Equal(t, good, got)

	// Signal the refresher to attempt again; the second mock call returns an
	// error. The cache should retain the prior good value.
	_, err := opsInstance.RefreshDefaultBranch(context.Background())
	require.NoError(t, err)

	// Give the refresher a moment to run.
	time.Sleep(50 * time.Millisecond)

	branch, err := opsInstance.DefaultBranch(context.Background())
	require.NoError(t, err)
	require.Equal(t, good, branch, "should retain last-known-good value on transient error")
}

// TestOpsDefaultBranch404Caching previously tested cache-on-fetch semantics
// of DefaultBranch. After the refresher refactor, DefaultBranch no longer
// performs HTTP itself — TestOpsBranchRefresher{SeedsCache,404IsRecorded,
// KeepsLastOnError} above cover the equivalent behaviour.
