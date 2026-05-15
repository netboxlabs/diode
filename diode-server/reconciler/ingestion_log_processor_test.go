package reconciler_test

import (
	"context"
	"errors"
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
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func newIngestionLogProcessorTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newTestIngestionLog() *reconcilerpb.IngestionLog {
	return &reconcilerpb.IngestionLog{
		DataType:        "dcim.device",
		State:           reconcilerpb.State_QUEUED,
		Entity:          &diodepb.Entity{},
		SdkName:         "test-sdk",
		ProducerAppName: "test-app",
	}
}

func TestIngestionLogProcessor_Name(t *testing.T) {
	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, nil, nil, nil, nil)
	assert.Equal(t, "reconciler-ingestion-log-processor", p.Name())
}

func TestIngestionLogProcessor_StartStop(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}
}

func TestIngestionLogProcessor_StopViaMethod(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	repo.On("ClaimQueuedIngestionLogs", mock.Anything, mock.Anything).Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

	ctx := context.Background()
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(50 * time.Millisecond)
	require.NoError(t, p.Stop())

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}
}

func TestIngestionLogProcessor_ProcessesItems(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	ingestionLog := newTestIngestionLog()
	batch := []ops.QueuedIngestionLog{
		{ID: 1, IngestionLog: ingestionLog},
		{ID: 2, IngestionLog: ingestionLog},
	}

	csID := int32Ptr(10)
	cs := &changeset.ChangeSet{
		Changes: []changeset.Change{{ChangeType: "create", ObjectType: "dcim.device"}},
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
		{IngestionLogID: 1, ChangeSetID: csID, ChangeSet: cs},
		{IngestionLogID: 2, ChangeSetID: csID, ChangeSet: cs},
	}).Once()
	mockMetrics.On("RecordChangeSetCreate", mock.Anything, true, int64(1)).Times(2)

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

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

	mockOps.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestIngestionLogProcessor_GenerateChangeSetError(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	ingestionLog := newTestIngestionLog()
	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: ingestionLog}}

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
		{IngestionLogID: 1, Err: errors.New("netbox error")},
	}).Once()
	mockMetrics.On("RecordChangeSetCreate", mock.Anything, false, int64(0)).Once()

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	<-claimed
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	mockOps.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestIngestionLogProcessor_BackpressureSkipsProcessing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	var backpressureActive atomic.Bool
	backpressureActive.Store(true)
	backpressure := func(_ context.Context) bool {
		return backpressureActive.Load()
	}

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, backpressure)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	repo.AssertNotCalled(t, "ClaimQueuedIngestionLogs", mock.Anything, mock.Anything)
}

func TestIngestionLogProcessor_BackpressureReleasedResumesProcessing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	ingestionLog := newTestIngestionLog()
	batch := []ops.QueuedIngestionLog{{ID: 1, IngestionLog: ingestionLog}}

	var backpressureActive atomic.Bool
	backpressureActive.Store(true)
	backpressure := func(_ context.Context) bool {
		return backpressureActive.Load()
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
		{IngestionLogID: 1},
	}).Once()

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, backpressure)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(100 * time.Millisecond)
	repo.AssertNotCalled(t, "ClaimQueuedIngestionLogs", mock.Anything, mock.Anything)

	backpressureActive.Store(false)

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not resume after backpressure released")
	}

	cancel()
	<-done

	mockOps.AssertExpectations(t)
}

func TestIngestionLogProcessor_NilBackpressureIsIgnored(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	claimed := make(chan struct{}, 1)
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("processor never polled with nil backpressure")
	}

	cancel()
	<-done
}

func TestIngestionLogProcessor_ClaimError(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	errCh := make(chan struct{}, 1)
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return(nil, errors.New("db connection error")).Once().Run(func(_ mock.Arguments) {
		select {
		case errCh <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimQueuedIngestionLogs", mock.Anything, int32(100)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	p := reconciler.NewIngestionLogProcessor(newIngestionLogProcessorTestLogger(), reconciler.Config{}, repo, mockOps, mockMetrics, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	<-errCh
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}
}
