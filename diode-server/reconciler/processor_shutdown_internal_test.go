package reconciler

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func testChangeSet() *changeset.ChangeSet {
	return &changeset.ChangeSet{
		Changes: []changeset.Change{{ChangeType: "create", ObjectType: "dcim.device"}},
	}
}

// These tests cover the graceful-shutdown contract shared by AutoApplyProcessor
// and IngestionLogProcessor. They live in the internal `reconciler` package
// (not `reconciler_test`) so they can shrink the package-level
// defaultProcessorGracefulShutdownTimeout var instead of waiting 30s.

func newShutdownTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newShutdownTestIngestionLog() *reconcilerpb.IngestionLog {
	return &reconcilerpb.IngestionLog{
		DataType:        "dcim.device",
		State:           reconcilerpb.State_QUEUED,
		Entity:          &diodepb.Entity{},
		SdkName:         "test-sdk",
		ProducerAppName: "test-app",
	}
}

// withShortGracefulShutdown swaps the package timeout to a short value for
// the duration of a test so the force-cancel path is observable in <1s.
func withShortGracefulShutdown(t *testing.T, d time.Duration) {
	t.Helper()
	prev := defaultProcessorGracefulShutdownTimeout
	defaultProcessorGracefulShutdownTimeout = d
	t.Cleanup(func() { defaultProcessorGracefulShutdownTimeout = prev })
}

// --- AutoApplyProcessor ---

// TestAutoApplyProcessor_StopWaitsForInFlightBatch verifies that an in-flight
// /bulk-plan-apply call is allowed to finish (and record its result) before
// Stop() returns. Without graceful drain, Stop would cancel the work ctx
// immediately and leave rows orphaned in APPLYING.
func TestAutoApplyProcessor_StopWaitsForInFlightBatch(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}

	// First claim returns a batch; subsequent claims return empty so workers idle.
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	releaseBulk := make(chan struct{})
	var bulkCompleted atomic.Bool

	mockOps.On("BulkPlanApply", mock.Anything, batch, "").Return(
		[]ops.BulkPlanApplyResult{{IngestionLogID: 1, ChangeSet: testChangeSet()}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		select {
		case <-releaseBulk:
		case <-time.After(5 * time.Second):
			t.Errorf("test fixture timed out waiting for release")
		}
		// Verify the work ctx is still alive while we're inside the call.
		// (Graceful drain must not cancel it.)
		// assert (not require) because we're inside a mock Run callback that
		// runs on a separate goroutine - require.X's t.FailNow only stops
		// the calling goroutine, leaving the test in a half-failed state.
		assert.NoError(t, ctx.Err(), "work ctx was canceled while batch was in-flight")
		bulkCompleted.Store(true)
	}).Once()

	// RecordChangeSetCreate fires after BulkPlanApply returns; its presence
	// proves the post-HTTP loop ran, which is what graceful drain enables.
	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()
	mockMetrics.On("RecordChangeSetApply", mock.Anything, true, int64(1)).Once()

	p := NewAutoApplyProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(context.Background()) }()

	select {
	case <-bulkStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("BulkPlanApply never invoked")
	}

	stopReturned := make(chan struct{})
	go func() {
		require.NoError(t, p.Stop())
		close(stopReturned)
	}()

	// Stop must not return while BulkPlanApply is still in flight.
	select {
	case <-stopReturned:
		t.Fatal("Stop returned before in-flight batch completed")
	case <-time.After(150 * time.Millisecond):
	}
	require.False(t, bulkCompleted.Load(), "BulkPlanApply completed before release")

	close(releaseBulk)

	select {
	case <-stopReturned:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return after batch was released")
	}
	require.True(t, bulkCompleted.Load(), "BulkPlanApply did not complete")

	select {
	case err := <-startDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

// TestAutoApplyProcessor_ParentCancelWaitsForInFlightBatch verifies the same
// drain semantics apply when the parent ctx is canceled (the
// orchestrator-driven tenant-restart path), not just direct Stop() calls.
func TestAutoApplyProcessor_ParentCancelWaitsForInFlightBatch(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	releaseBulk := make(chan struct{})

	mockOps.On("BulkPlanApply", mock.Anything, batch, "").Return(
		[]ops.BulkPlanApplyResult{{IngestionLogID: 1, ChangeSet: testChangeSet()}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		select {
		case <-releaseBulk:
		case <-time.After(5 * time.Second):
			t.Errorf("test fixture timed out")
		}
		require.NoError(t, ctx.Err(), "work ctx canceled mid-batch from parent-cancel path")
	}).Once()

	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()
	mockMetrics.On("RecordChangeSetApply", mock.Anything, true, int64(1)).Once()

	p := NewAutoApplyProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	parentCtx, parentCancel := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(parentCtx) }()

	select {
	case <-bulkStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("BulkPlanApply never invoked")
	}

	parentCancel()

	// Start must not return while BulkPlanApply is still in flight, even
	// though the parent ctx was canceled.
	select {
	case <-startDone:
		t.Fatal("Start returned before in-flight batch completed")
	case <-time.After(150 * time.Millisecond):
	}

	close(releaseBulk)

	select {
	case err := <-startDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after batch was released")
	}
}

// TestAutoApplyProcessor_ForceCancelAfterTimeout verifies that an in-flight
// batch that exceeds the drain timeout is force-canceled (work ctx canceled)
// rather than blocking shutdown indefinitely.
func TestAutoApplyProcessor_ForceCancelAfterTimeout(t *testing.T) {
	withShortGracefulShutdown(t, 100*time.Millisecond)

	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedForAutoApply", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	var ctxCanceledDuringBulk atomic.Bool

	mockOps.On("BulkPlanApply", mock.Anything, batch, "").Return(
		[]ops.BulkPlanApplyResult{{IngestionLogID: 1}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		// Wait for the work ctx to be canceled by force-cancel. With a 100ms
		// drain timeout, this should fire within ~150ms.
		select {
		case <-ctx.Done():
			ctxCanceledDuringBulk.Store(true)
		case <-time.After(5 * time.Second):
			t.Errorf("work ctx was never canceled")
		}
	}).Once()

	// The post-HTTP loop runs synchronously after BulkPlanApply returns;
	// since we return a result with no ChangeSet, no metric is recorded.

	p := NewAutoApplyProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(context.Background()) }()

	<-bulkStarted

	start := time.Now()
	require.NoError(t, p.Stop())
	elapsed := time.Since(start)

	require.True(t, ctxCanceledDuringBulk.Load(), "work ctx was not canceled inside BulkPlanApply")
	require.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "Stop returned before drain timeout elapsed")
	require.Less(t, elapsed, 2*time.Second, "Stop took too long after force-cancel")

	select {
	case <-startDone:
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

// TestAutoApplyProcessor_StopBeforeStartIsSafe verifies the
// never-started-yet case (e.g., Stop called from a defer after early-exit
// from constructor wiring) does not panic.
func TestAutoApplyProcessor_StopBeforeStartIsSafe(t *testing.T) {
	p := NewAutoApplyProcessor(newShutdownTestLogger(), Config{}, nil, nil, nil, nil)
	require.NoError(t, p.Stop())
	require.NoError(t, p.Stop()) // idempotent
}

// --- IngestionLogProcessor ---

// TestIngestionLogProcessor_StopWaitsForInFlightBatch mirrors the AutoApply
// case but for BulkPlan: in-flight plan must finish before Stop returns so
// rows reach OPEN/NO_CHANGES rather than staying claimed.
func TestIngestionLogProcessor_StopWaitsForInFlightBatch(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	releaseBulk := make(chan struct{})
	var bulkCompleted atomic.Bool

	mockOps.On("BulkPlan", mock.Anything, batch, "").Return(
		[]ops.BulkGenerateChangeSetResult{{IngestionLogID: 1, ChangeSet: testChangeSet()}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		select {
		case <-releaseBulk:
		case <-time.After(5 * time.Second):
			t.Errorf("test fixture timed out")
		}
		// assert (not require) because we're inside a mock Run callback that
		// runs on a separate goroutine - require.X's t.FailNow only stops
		// the calling goroutine, leaving the test in a half-failed state.
		assert.NoError(t, ctx.Err(), "work ctx was canceled while batch was in-flight")
		bulkCompleted.Store(true)
	}).Once()

	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()

	p := NewIngestionLogProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(context.Background()) }()

	<-bulkStarted

	stopReturned := make(chan struct{})
	go func() {
		require.NoError(t, p.Stop())
		close(stopReturned)
	}()

	select {
	case <-stopReturned:
		t.Fatal("Stop returned before in-flight batch completed")
	case <-time.After(150 * time.Millisecond):
	}
	require.False(t, bulkCompleted.Load())

	close(releaseBulk)

	select {
	case <-stopReturned:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return after batch was released")
	}
	require.True(t, bulkCompleted.Load())

	select {
	case err := <-startDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

// TestIngestionLogProcessor_ParentCancelWaitsForInFlightBatch covers the
// orchestrator-driven cancellation path for IngestionLogProcessor.
func TestIngestionLogProcessor_ParentCancelWaitsForInFlightBatch(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	releaseBulk := make(chan struct{})

	mockOps.On("BulkPlan", mock.Anything, batch, "").Return(
		[]ops.BulkGenerateChangeSetResult{{IngestionLogID: 1, ChangeSet: testChangeSet()}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		select {
		case <-releaseBulk:
		case <-time.After(5 * time.Second):
			t.Errorf("test fixture timed out")
		}
		require.NoError(t, ctx.Err(), "work ctx canceled mid-batch from parent-cancel path")
	}).Once()

	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()

	p := NewIngestionLogProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	parentCtx, parentCancel := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(parentCtx) }()

	<-bulkStarted
	parentCancel()

	select {
	case <-startDone:
		t.Fatal("Start returned before in-flight batch completed")
	case <-time.After(150 * time.Millisecond):
	}

	close(releaseBulk)

	select {
	case err := <-startDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after batch was released")
	}
}

// TestIngestionLogProcessor_ForceCancelAfterTimeout: drain-timeout path.
func TestIngestionLogProcessor_ForceCancelAfterTimeout(t *testing.T) {
	withShortGracefulShutdown(t, 100*time.Millisecond)

	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: newShutdownTestIngestionLog()}}
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return(batch, nil).Once()
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()

	bulkStarted := make(chan struct{})
	var ctxCanceledDuringBulk atomic.Bool

	mockOps.On("BulkPlan", mock.Anything, batch, "").Return(
		[]ops.BulkGenerateChangeSetResult{{IngestionLogID: 1}},
	).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(bulkStarted)
		select {
		case <-ctx.Done():
			ctxCanceledDuringBulk.Store(true)
		case <-time.After(5 * time.Second):
			t.Errorf("work ctx was never canceled")
		}
	}).Once()

	// No ChangeSet in the result, so no metric is recorded.

	p := NewIngestionLogProcessor(newShutdownTestLogger(), Config{}, repo, mockOps, mockMetrics, nil)

	startDone := make(chan error, 1)
	go func() { startDone <- p.Start(context.Background()) }()

	<-bulkStarted

	start := time.Now()
	require.NoError(t, p.Stop())
	elapsed := time.Since(start)

	require.True(t, ctxCanceledDuringBulk.Load(), "work ctx was not canceled inside BulkPlan")
	require.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	require.Less(t, elapsed, 2*time.Second)

	select {
	case <-startDone:
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

// TestIngestionLogProcessor_StopBeforeStartIsSafe: defensive check.
func TestIngestionLogProcessor_StopBeforeStartIsSafe(t *testing.T) {
	p := NewIngestionLogProcessor(newShutdownTestLogger(), Config{}, nil, nil, nil, nil)
	require.NoError(t, p.Stop())
	require.NoError(t, p.Stop())
}
