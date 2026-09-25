package reconciler_test

import (
	"context"
	"database/sql"
	"errors"
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

// expectDedupPassthrough stubs WithDedupLocks to invoke fn with the mock
// repository itself, mirroring the real implementation's tx-scoped repo.
func expectDedupPassthrough(m *mocks.Repository) {
	m.EXPECT().WithDedupLocks(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ []string, fn func(reconops.DedupRepository) error) error {
			return fn(m)
		})
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

			expectDedupPassthrough(mockRepository)

			mockRepository.EXPECT().FindPriorIngestionLogByEntityHash(mock.Anything, mock.AnythingOfType("string"), (*string)(nil)).
				Return(tt.mockFindPriorIngestionLogID, tt.mockFindPriorIngestionLog, tt.mockFindPriorIngestionLogError)

			// Mock CreateIngestionLog
			if tt.mockCreateIngestionLog != nil {
				mockRepository.EXPECT().CreateIngestionLog(mock.Anything, tt.ingestionLog, tt.sourceMetadata, mock.AnythingOfType("string")).
					Return(tt.mockCreateIngestionLog.id, tt.mockCreateIngestionLog.error)
			}

			if tt.expectWasDuplicate {
				mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, map[int32]int32{*tt.mockFindPriorIngestionLogID: 1}).
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

	expectDedupPassthrough(mockRepository)
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
	mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, map[int32]int32{10: 1, 11: 1}).
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

func TestOpsBulkCreateIngestionLogsInBatchDuplicates(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}

	// Three copies of a brand-new entity and two copies of an entity with a
	// prior: one row is inserted for the new entity, everything else counts
	// as duplicates.
	newLog := func(id string) *pb.IngestionLog {
		return &pb.IngestionLog{Id: id, ObjectType: netbox.SiteObjectType, State: pb.State_QUEUED, Entity: testEntity}
	}
	logs := []*pb.IngestionLog{newLog("new-1"), newLog("new-2"), newLog("dup-1"), newLog("new-3"), newLog("dup-2")}
	hashes := []string{"hash-new", "hash-new", "hash-prior", "hash-new", "hash-prior"}

	prior := &reconops.PriorIngestionLog{ID: 10, IngestionLog: &pb.IngestionLog{Id: "prior", State: pb.State_APPLIED, Entity: testEntity}}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	expectDedupPassthrough(mockRepository)
	mockRepository.EXPECT().FindPriorIngestionLogsByEntityHashes(mock.Anything, mock.MatchedBy(func(hashes []string) bool {
		return len(hashes) == 2
	}), (*string)(nil)).
		Return(map[string]*reconops.PriorIngestionLog{"hash-prior": prior}, nil)
	// Only the first occurrence of hash-new is inserted.
	mockRepository.EXPECT().BulkCreateIngestionLogs(mock.Anything, []*pb.IngestionLog{logs[0]}, mock.Anything, []string{"hash-new"}).
		Return(map[string]int32{"new-1": 42}, nil)
	// The new row absorbs its two in-batch copies; the prior absorbs both of its copies.
	mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, map[int32]int32{42: 2, 10: 2}).
		Return(map[int32]bool{42: false, 10: true}, nil)

	results, err := opsInstance.BulkCreateIngestionLogs(ctx, logs, [][]byte{nil, nil, nil, nil, nil}, hashes)
	require.NoError(t, err)
	require.Len(t, results, 5)

	// First occurrence of the new entity is the inserted row.
	require.False(t, results[0].WasDuplicate)
	require.Equal(t, int32(42), results[0].ID)

	// In-batch copies resolve to the freshly inserted row, not new rows.
	for _, i := range []int{1, 3} {
		require.True(t, results[i].WasDuplicate, "result %d", i)
		require.False(t, results[i].Requeued, "result %d", i)
		require.Equal(t, int32(42), results[i].ID, "result %d", i)
	}

	// Copies of the prior entity dedup against it and pick up the requeue.
	for _, i := range []int{2, 4} {
		require.True(t, results[i].WasDuplicate, "result %d", i)
		require.True(t, results[i].Requeued, "result %d", i)
		require.Equal(t, int32(10), results[i].ID, "result %d", i)
		require.Equal(t, pb.State_QUEUED, results[i].IngestionLog.State, "result %d", i)
	}
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

	expectDedupPassthrough(mockRepository)
	mockRepository.EXPECT().FindPriorIngestionLogsByEntityHashes(mock.Anything, []string{"hash-dup"}, (*string)(nil)).
		Return(map[string]*reconops.PriorIngestionLog{"hash-dup": prior}, nil)
	mockRepository.EXPECT().BulkMarkDuplicates(mock.Anything, map[int32]int32{10: 1}).
		Return(nil, fmt.Errorf("database error"))

	_, err := opsInstance.BulkCreateIngestionLogs(ctx, []*pb.IngestionLog{dupLog}, [][]byte{nil}, []string{"hash-dup"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to bulk mark duplicates")
}

func TestOpsBulkPlanDriftDeviation(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}

	// Both logs were requeued from APPLIED by a duplicate observation.
	// Log 1 re-plans with drift -> new deviation; log 2 re-plans clean ->
	// restored to APPLIED, nothing persisted.
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
		{ID: 2, IngestionLog: &pb.IngestionLog{Id: "log-2", ObjectType: netbox.SiteObjectType, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().BulkPlan(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanResponse{
		Results: []netboxdiodeplugin.BulkPlanResult{
			{ID: "1", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-1", Changes: []netboxdiodeplugin.Change{
				{ID: "ch-1", ChangeType: "update", ObjectType: netbox.SiteObjectType, ObjectPrimaryValue: "test-site-1"},
			}}},
			{ID: "2", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-2"}},
		},
	}, nil)

	newCSID := int32(201)
	mockRepository.EXPECT().BulkCreateDriftDeviations(mock.Anything, mock.MatchedBy(func(driftItems []reconops.DriftDeviationItem) bool {
		return len(driftItems) == 1 &&
			driftItems[0].PriorIngestionLogID == 1 &&
			driftItems[0].NewExternalID != "" &&
			driftItems[0].NewState == pb.State_OPEN &&
			len(driftItems[0].ChangeSet.Changes) == 1
	})).Return([]reconops.DriftDeviationResult{
		{PriorIngestionLogID: 1, NewIngestionLogID: 42, ChangeSetID: &newCSID},
	}, nil)

	mockRepository.EXPECT().BulkPersistChangeSets(mock.Anything, mock.MatchedBy(func(persistItems []reconops.BulkPersistItem) bool {
		return len(persistItems) == 1 &&
			persistItems[0].IngestionLogID == 2 &&
			persistItems[0].NewState == pb.State_APPLIED &&
			len(persistItems[0].ChangeSet.Changes) == 0
	}), mock.Anything).Return([]reconops.BulkPersistResult{{IngestionLogID: 2}}, nil)

	results := opsInstance.BulkPlan(ctx, items, "")
	require.Len(t, results, 2)

	require.NoError(t, results[0].Err)
	require.Equal(t, int32(42), results[0].IngestionLogID)
	require.Equal(t, &newCSID, results[0].ChangeSetID)

	require.NoError(t, results[1].Err)
	require.Equal(t, int32(2), results[1].IngestionLogID)
}

func TestOpsBulkPlanDriftPlanFailureRestoresPrior(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().BulkPlan(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanResponse{
		Results: []netboxdiodeplugin.BulkPlanResult{
			{ID: "1", Errors: []byte(`["boom"]`)},
		},
	}, nil)

	// The prior is restored to APPLIED; no FAILED state, no placeholder change set.
	mockRepository.EXPECT().UpdateIngestionLogStateWithError(mock.Anything, int32(1), pb.State_APPLIED, nil).Return(nil)

	results := opsInstance.BulkPlan(ctx, items, "")
	require.Len(t, results, 1)
	require.Error(t, results[0].Err)
	require.Equal(t, int32(1), results[0].IngestionLogID)
}

func TestOpsBulkPlanFallbackPreservesAppliedState(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}
	// Requeued from APPLIED, re-plans clean; bulk persist fails, so the
	// per-item fallback must restore APPLIED (not derive NO_CHANGES) and must
	// not write a change set.
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, State: pb.State_OPEN, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	mockNetBoxClient.EXPECT().BulkPlan(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanResponse{
		Results: []netboxdiodeplugin.BulkPlanResult{
			{ID: "1", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-1"}},
		},
	}, nil)
	mockRepository.EXPECT().BulkPersistChangeSets(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("database error"))
	mockRepository.EXPECT().UpdateIngestionLogStateWithError(mock.Anything, int32(1), pb.State_APPLIED, nil).Return(nil)

	results := opsInstance.BulkPlan(ctx, items, "")
	require.Len(t, results, 1)
	require.NoError(t, results[0].Err)
	require.Nil(t, results[0].ChangeSetID)
}

func TestOpsBulkPlanApplyDriftDeviation(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	testEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: "test-site-1"},
		},
	}

	// Both requeued from APPLIED with drift: log 1 applies cleanly -> new
	// APPLIED deviation; log 2 fails apply -> new FAILED deviation annotated
	// with the apply error.
	items := []reconops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &pb.IngestionLog{Id: "log-1", ObjectType: netbox.SiteObjectType, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
		{ID: 2, IngestionLog: &pb.IngestionLog{Id: "log-2", ObjectType: netbox.SiteObjectType, Entity: testEntity}, RequeuedFromState: pb.State_APPLIED},
	}

	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	change := netboxdiodeplugin.Change{ID: "ch-1", ChangeType: "update", ObjectType: netbox.SiteObjectType, ObjectPrimaryValue: "test-site-1"}
	mockNetBoxClient.EXPECT().BulkPlanApply(mock.Anything, mock.Anything).Return(&netboxdiodeplugin.BulkPlanApplyResponse{
		Results: []netboxdiodeplugin.BulkPlanApplyResult{
			{ID: "1", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-1", Changes: []netboxdiodeplugin.Change{change}}},
			{
				ID: "2", ChangeSet: &netboxdiodeplugin.ChangeSet{ID: "cs-2", Changes: []netboxdiodeplugin.Change{change}},
				Errors: &netboxdiodeplugin.BulkPlanApplyErrors{Apply: []byte(`["apply boom"]`)},
			},
		},
	}, nil)

	cs1, cs2 := int32(201), int32(202)
	mockRepository.EXPECT().BulkCreateDriftDeviations(mock.Anything, mock.MatchedBy(func(driftItems []reconops.DriftDeviationItem) bool {
		return len(driftItems) == 2 &&
			driftItems[0].PriorIngestionLogID == 1 && driftItems[0].NewState == pb.State_APPLIED &&
			driftItems[1].PriorIngestionLogID == 2 && driftItems[1].NewState == pb.State_FAILED
	})).Return([]reconops.DriftDeviationResult{
		{PriorIngestionLogID: 1, NewIngestionLogID: 41, ChangeSetID: &cs1},
		{PriorIngestionLogID: 2, NewIngestionLogID: 42, ChangeSetID: &cs2},
	}, nil)

	// Apply-error annotation targets the new deviation, not the restored prior.
	mockRepository.EXPECT().UpdateIngestionLogStateWithError(mock.Anything, int32(42), pb.State_FAILED, mock.Anything).Return(nil)

	results := opsInstance.BulkPlanApply(ctx, items, "")
	require.Len(t, results, 2)

	require.NoError(t, results[0].PlanErr)
	require.NoError(t, results[0].ApplyErr)
	require.Equal(t, int32(41), results[0].IngestionLogID)

	require.NoError(t, results[1].PlanErr)
	require.Error(t, results[1].ApplyErr)
	require.Equal(t, int32(42), results[1].IngestionLogID)
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

func TestOpsBranchRefresherRetriesInitialFailure(t *testing.T) {
	// The initial fetch races OAuth client bootstrap and can come back 403.
	// Retry with backoff instead of waiting a full DefaultBranchRefreshInterval,
	// which would leave the cache cold — and processors idle — for 5 minutes.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, logger, nil)

	want := &netboxdiodeplugin.Branch{ID: "branch-1", Name: "Branch One"}
	mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).
		Return((*netboxdiodeplugin.Branch)(nil), errors.New("get default branch failed with status 403: Invalid token")).Once()
	mockNetBoxClient.EXPECT().GetDefaultBranch(mock.Anything).Return(want, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opsInstance.Start(ctx)

	// Well inside DefaultBranchRefreshInterval, so only the retry can satisfy this.
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
