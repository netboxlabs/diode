package reconciler_test

import (
	"context"
	"database/sql"
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
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	reconops "github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

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
		mockRequeued                   bool

		expectedError      string
		expectWasDuplicate bool
		expectRequeued     bool
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
			name:           "duplicate found - prior requeued for re-plan",
			ingestionLog:   testIngestionLog,
			sourceMetadata: testSourceMetadata,

			mockFindPriorIngestionLogID: int32Ptr(5678),
			mockFindPriorIngestionLog: &pb.IngestionLog{
				Id:         "8a8ae517-85b9-466e-890c-aadb0771cc9e",
				ObjectType: netbox.SiteObjectType,
				State:      pb.State_APPLIED,
				RequestId:  "1abf059c-496f-4037-83c2-0e9b1d021e85",
				Entity:     testEntity,
			},
			mockFindPriorIngestionLogError: nil,
			mockRequeued:                   true,

			expectedError:      "",
			expectWasDuplicate: true,
			expectRequeued:     true,
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
				mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, []int32{*tt.mockFindPriorIngestionLogID}).
					Return(map[int32]bool{*tt.mockFindPriorIngestionLogID: tt.mockRequeued}, nil)
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
			if tt.expectWasDuplicate {
				require.Equal(t, tt.mockFindPriorIngestionLog, result.IngestionLog)
			} else {
				require.Equal(t, tt.ingestionLog, result.IngestionLog)
			}
			require.Equal(t, tt.expectWasDuplicate, result.WasDuplicate)
			require.Equal(t, tt.expectRequeued, result.Requeued)
			if tt.expectRequeued {
				require.Equal(t, pb.State_QUEUED, result.IngestionLog.State)
			}
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

func TestOpsBulkCreateIngestionLogsDuplicateRequeue(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}

	newLog := &pb.IngestionLog{Id: "new-log", ObjectType: netbox.SiteObjectType, State: pb.State_QUEUED, Entity: testEntity}
	dupOfApplied := &pb.IngestionLog{Id: "dup-of-applied", ObjectType: netbox.SiteObjectType, State: pb.State_QUEUED, Entity: testEntity}
	dupOfOpen := &pb.IngestionLog{Id: "dup-of-open", ObjectType: netbox.SiteObjectType, State: pb.State_QUEUED, Entity: testEntity}

	priorApplied := &reconops.PriorIngestionLog{ID: 10, IngestionLog: &pb.IngestionLog{Id: "prior-applied", State: pb.State_APPLIED, Entity: testEntity}}
	priorOpen := &reconops.PriorIngestionLog{ID: 11, IngestionLog: &pb.IngestionLog{Id: "prior-open", State: pb.State_OPEN, Entity: testEntity}}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	// Ops dedupes hashes via a map, so the argument order is nondeterministic
	mockRepository.EXPECT().FindPriorIngestionLogsByEntityHashes(mock.Anything, mock.MatchedBy(func(hashes []string) bool {
		want := map[string]struct{}{"hash-new": {}, "hash-applied": {}, "hash-open": {}}
		if len(hashes) != len(want) {
			return false
		}
		for _, h := range hashes {
			if _, ok := want[h]; !ok {
				return false
			}
		}
		return true
	}), (*string)(nil)).
		Return(map[string]*reconops.PriorIngestionLog{"hash-applied": priorApplied, "hash-open": priorOpen}, nil)
	mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, []int32{10, 11}).
		Return(map[int32]bool{10: true, 11: false}, nil)
	mockRepository.EXPECT().BulkCreateIngestionLogs(mock.Anything, []*pb.IngestionLog{newLog}, mock.Anything, []string{"hash-new"}).
		Return(map[string]int32{"new-log": 1}, nil)

	results, err := opsInstance.BulkCreateIngestionLogs(ctx,
		[]*pb.IngestionLog{newLog, dupOfApplied, dupOfOpen},
		[][]byte{nil, nil, nil},
		[]string{"hash-new", "hash-applied", "hash-open"})
	require.NoError(t, err)
	require.Len(t, results, 3)

	require.False(t, results[0].WasDuplicate)
	require.False(t, results[0].Requeued)
	require.Equal(t, int32(1), results[0].ID)

	require.True(t, results[1].WasDuplicate)
	require.True(t, results[1].Requeued)
	require.Equal(t, int32(10), results[1].ID)
	require.Equal(t, pb.State_QUEUED, results[1].IngestionLog.State)

	require.True(t, results[2].WasDuplicate)
	require.False(t, results[2].Requeued)
	require.Equal(t, int32(11), results[2].ID)
	require.Equal(t, pb.State_OPEN, results[2].IngestionLog.State)
}

func TestOpsBulkCreateIngestionLogsMarkDuplicatesError(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}
	dupLog := &pb.IngestionLog{Id: "dup-log", ObjectType: netbox.SiteObjectType, State: pb.State_QUEUED, Entity: testEntity}
	prior := &reconops.PriorIngestionLog{ID: 10, IngestionLog: &pb.IngestionLog{Id: "prior", State: pb.State_APPLIED, Entity: testEntity}}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockRepository.EXPECT().FindPriorIngestionLogsByEntityHashes(mock.Anything, []string{"hash-dup"}, (*string)(nil)).
		Return(map[string]*reconops.PriorIngestionLog{"hash-dup": prior}, nil)
	mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, []int32{10}).
		Return(nil, fmt.Errorf("database error"))

	_, err := opsInstance.BulkCreateIngestionLogs(ctx, []*pb.IngestionLog{dupLog}, [][]byte{nil}, []string{"hash-dup"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to bulk mark duplicates")
}

func TestOpsBulkPlanUnchangedMarker(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}

	// Two requeued logs re-plan to an empty diff: log 1 has an applied change
	// set with changes (expect an "unchanged" marker), log 2 has none (first
	// plan; expect state update only).
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, Entity: testEntity}},
		{ID: 2, IngestionLog: &pb.IngestionLog{Id: "log-2", ObjectType: netbox.SiteObjectType, Entity: testEntity}},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().BulkPlan(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanResponse{
		Results: []netboxdiodeplugin.BulkPlanResult{
			{ID: "1", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-1"}},
			{ID: "2", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-2"}},
		},
	}, nil)
	mockRepository.EXPECT().LatestChangeSetsHaveChanges(mock.Anything, []int32{1, 2}).
		Return(map[int32]bool{1: true}, nil)

	csID1, csID2 := int32(101), int32(102)
	mockRepository.EXPECT().BulkPersistChangeSets(mock.Anything, mock.MatchedBy(func(persistItems []reconops.BulkPersistItem) bool {
		if len(persistItems) != 2 {
			return false
		}
		markerOK := persistItems[0].PersistEmptyChangeSet &&
			persistItems[0].NewState == pb.State_NO_CHANGES &&
			persistItems[0].ChangeSet.DeviationName != nil &&
			*persistItems[0].ChangeSet.DeviationName == "Site unchanged"
		noMarkerOK := !persistItems[1].PersistEmptyChangeSet &&
			persistItems[1].NewState == pb.State_NO_CHANGES
		return markerOK && noMarkerOK
	}), mock.Anything).Return([]reconops.BulkPersistResult{
		{IngestionLogID: 1, ChangeSetID: &csID1},
		{IngestionLogID: 2, ChangeSetID: &csID2},
	}, nil)

	results := opsInstance.BulkPlan(ctx, items, "")
	require.Len(t, results, 2)
	require.NoError(t, results[0].Err)
	require.NoError(t, results[1].Err)
}

func TestOpsBulkPlanUnchangedMarkerLookupFailure(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, Entity: testEntity}},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().BulkPlan(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanResponse{
		Results: []netboxdiodeplugin.BulkPlanResult{
			{ID: "1", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-1"}},
		},
	}, nil)
	mockRepository.EXPECT().LatestChangeSetsHaveChanges(mock.Anything, []int32{1}).
		Return(nil, fmt.Errorf("database error"))

	// Marker is skipped on lookup failure; plan still persists the state.
	mockRepository.EXPECT().BulkPersistChangeSets(mock.Anything, mock.MatchedBy(func(persistItems []reconops.BulkPersistItem) bool {
		return len(persistItems) == 1 && !persistItems[0].PersistEmptyChangeSet && persistItems[0].NewState == pb.State_NO_CHANGES
	}), mock.Anything).Return([]reconops.BulkPersistResult{{IngestionLogID: 1}}, nil)

	results := opsInstance.BulkPlan(ctx, items, "")
	require.Len(t, results, 1)
	require.NoError(t, results[0].Err)
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
