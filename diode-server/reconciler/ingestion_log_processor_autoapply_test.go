package reconciler_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func newAutoApplyTestLog() *reconcilerpb.IngestionLog {
	return &reconcilerpb.IngestionLog{
		DataType:        "extras.customfield",
		ObjectType:      "extras.customfield",
		State:           reconcilerpb.State_QUEUED,
		Entity:          &diodepb.Entity{},
		SdkName:         "test-sdk",
		ProducerAppName: "orb-pro-credentials-bootstrap",
	}
}

func newAutoApplyTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// TestIngestionLogProcessor_AutoApplyPolicyMatch_AppliesItem verifies that
// when WithAutoApplyPolicy is set and the policy returns true for a record,
// the processor calls ApplyChangeSetForLog and records a successful apply
// metric — bypassing the OPEN review queue.
func TestIngestionLogProcessor_AutoApplyPolicyMatch_AppliesItem(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	log := newAutoApplyTestLog()
	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: log}}
	cs := &changeset.ChangeSet{
		Changes: []changeset.Change{{ChangeType: changeset.ChangeTypeCreate, ObjectType: "extras.customfield"}},
	}

	claimed := make(chan struct{}, 1)
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()
	mockOps.On("BulkPlan", mock.Anything, batch, "").Return([]ops.BulkGenerateChangeSetResult{
		{IngestionLogID: 1, ChangeSetID: int32Ptr(10), ChangeSet: cs},
	}).Once()
	mockOps.On("ApplyChangeSetForLog", mock.Anything, batch[0], cs, "").Return(nil).Once()

	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()
	mockMetrics.On("RecordChangeSetApply", mock.Anything, true, int64(1)).Once()

	policy := reconciler.AutoApplyPolicy(func(_ *reconcilerpb.IngestionLog, _ *changeset.ChangeSet) bool {
		return true
	})

	p := reconciler.NewIngestionLogProcessor(
		newAutoApplyTestLogger(),
		reconciler.Config{},
		repo,
		mockOps,
		mockMetrics,
		nil,
		reconciler.WithAutoApplyPolicy(policy),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("batch was never claimed")
	}

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}
}

// TestIngestionLogProcessor_AutoApplyPolicyNoMatch_DoesNotApply verifies that
// when the policy returns false for a record, the processor does NOT call
// ApplyChangeSetForLog — leaving the row in OPEN for review, as it would
// without any policy.
func TestIngestionLogProcessor_AutoApplyPolicyNoMatch_DoesNotApply(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	log := newAutoApplyTestLog()
	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: log}}
	cs := &changeset.ChangeSet{
		Changes: []changeset.Change{{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "extras.customfield"}},
	}

	claimed := make(chan struct{}, 1)
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	mockOps.On("DefaultBranch", mock.Anything).Return(nil, nil).Maybe()
	mockOps.On("BulkPlan", mock.Anything, batch, "").Return([]ops.BulkGenerateChangeSetResult{
		{IngestionLogID: 1, ChangeSetID: int32Ptr(10), ChangeSet: cs},
	}).Once()

	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Once()

	policy := reconciler.AutoApplyPolicy(func(_ *reconcilerpb.IngestionLog, _ *changeset.ChangeSet) bool {
		return false
	})

	p := reconciler.NewIngestionLogProcessor(
		newAutoApplyTestLogger(),
		reconciler.Config{},
		repo,
		mockOps,
		mockMetrics,
		nil,
		reconciler.WithAutoApplyPolicy(policy),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("batch was never claimed")
	}

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}

	mockOps.AssertNotCalled(t, "ApplyChangeSetForLog")
}
