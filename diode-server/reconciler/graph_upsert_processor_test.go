package reconciler_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/graph"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

// fakeUpserter is a minimal GraphEntityUpserter that returns canned responses
// per call. It tracks how many times it was invoked so tests can assert the
// processor walked the full batch.
type fakeUpserter struct {
	mu      sync.Mutex
	results []error
	calls   int
	entity  []*diodepb.Entity
	reqMeta []map[string]any
}

func (f *fakeUpserter) UpsertEntity(_ context.Context, entity *diodepb.Entity, requestMetadata ...map[string]any) (*graph.Node, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entity = append(f.entity, entity)
	if len(requestMetadata) > 0 {
		f.reqMeta = append(f.reqMeta, requestMetadata[0])
	} else {
		f.reqMeta = append(f.reqMeta, nil)
	}
	idx := f.calls
	f.calls++
	if idx < len(f.results) {
		if err := f.results[idx]; err != nil {
			return nil, err
		}
	}
	return &graph.Node{}, nil
}

func (f *fakeUpserter) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeUpserter) recordedMetadata() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]map[string]any, len(f.reqMeta))
	copy(out, f.reqMeta)
	return out
}

func newGraphUpsertTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newGraphUpsertCandidate(id int32, objectType string, sourceMetadata []byte) ops.QueuedIngestionLog {
	return ops.QueuedIngestionLog{
		ID: id,
		IngestionLog: &reconcilerpb.IngestionLog{
			ObjectType:      objectType,
			Entity:          &diodepb.Entity{Entity: &diodepb.Entity_Site{Site: &diodepb.Site{Name: "site-" + objectType}}},
			SdkName:         "test-sdk",
			ProducerAppName: "test-app",
		},
		SourceMetadata: sourceMetadata,
	}
}

func TestGraphUpsertProcessor_Name(t *testing.T) {
	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, nil, nil, factory, nil)
	assert.Equal(t, "reconciler-graph-upsert-processor", p.Name())
}

func TestGraphUpsertProcessor_StartStop(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, mock.Anything, mock.Anything).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

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

func TestGraphUpsertProcessor_StopViaMethod(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, mock.Anything, mock.Anything).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

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

func TestGraphUpsertProcessor_SuccessfulBatchMarksAllRows(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{
		newGraphUpsertCandidate(1, "dcim.site", []byte(`{"run_id":"abc"}`)),
		newGraphUpsertCandidate(2, "dcim.site", []byte(`{"run_id":"abc"}`)),
	}

	claimed := make(chan struct{}, 1)
	marked := make(chan struct{}, 1)

	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	repo.On("MarkGraphUpserted", mock.Anything, []int32{1, 2}).Return(nil).Once().Run(func(_ mock.Arguments) {
		select {
		case marked <- struct{}{}:
		default:
		}
	})
	mockMetrics.On("RecordGraphUpsert", mock.Anything, true, "dcim.site", mock.AnythingOfType("float64")).Times(2)

	upserter := &fakeUpserter{}
	factory := func() reconciler.GraphEntityUpserter { return upserter }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("batch was never claimed")
	}
	select {
	case <-marked:
	case <-time.After(5 * time.Second):
		t.Fatal("batch was never marked upserted")
	}

	cancel()
	<-done

	assert.Equal(t, 2, upserter.callCount(), "upserter should have been called for each row")
	// Verify request metadata was unmarshaled and passed through.
	for _, meta := range upserter.recordedMetadata() {
		require.NotNil(t, meta)
		assert.Equal(t, "abc", meta["run_id"])
	}

	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "ReleaseGraphUpsertClaims", mock.Anything, mock.Anything)
}

func TestGraphUpsertProcessor_FailedRowsReleaseClaims(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{
		newGraphUpsertCandidate(10, "dcim.site", nil),
		newGraphUpsertCandidate(11, "dcim.site", nil),
		newGraphUpsertCandidate(12, "dcim.site", nil),
	}

	claimed := make(chan struct{}, 1)
	mixed := make(chan struct{}, 1)

	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	// Row 11 fails; rows 10 and 12 succeed.
	repo.On("MarkGraphUpserted", mock.Anything, []int32{10, 12}).Return(nil).Once()
	repo.On("ReleaseGraphUpsertClaims", mock.Anything, []int32{11}).Return(nil).Once().Run(func(_ mock.Arguments) {
		select {
		case mixed <- struct{}{}:
		default:
		}
	})

	mockMetrics.On("RecordGraphUpsert", mock.Anything, true, "dcim.site", mock.AnythingOfType("float64")).Times(2)
	mockMetrics.On("RecordGraphUpsert", mock.Anything, false, "dcim.site", mock.AnythingOfType("float64")).Once()

	upserter := &fakeUpserter{results: []error{nil, errors.New("graph DB unreachable"), nil}}
	factory := func() reconciler.GraphEntityUpserter { return upserter }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	<-claimed
	<-mixed

	cancel()
	<-done

	repo.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestGraphUpsertProcessor_NilEntityIsTreatedAsSuccess(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{{
		ID:           42,
		IngestionLog: &reconcilerpb.IngestionLog{ObjectType: "dcim.site"},
	}}

	claimed := make(chan struct{}, 1)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	repo.On("MarkGraphUpserted", mock.Anything, []int32{42}).Return(nil).Once()

	upserter := &fakeUpserter{}
	factory := func() reconciler.GraphEntityUpserter { return upserter }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	<-claimed
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	assert.Equal(t, 0, upserter.callCount(), "nil entity must skip the graph service entirely")
	repo.AssertNotCalled(t, "ReleaseGraphUpsertClaims", mock.Anything, mock.Anything)
}

func TestGraphUpsertProcessor_ClaimErrorIsRetried(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	errCh := make(chan struct{}, 1)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return(nil, errors.New("db down")).Once().Run(func(_ mock.Arguments) {
		select {
		case errCh <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	<-errCh
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit after claim error")
	}
}

func TestGraphUpsertProcessor_BackpressureSkipsProcessing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	var backpressureActive atomic.Bool
	backpressureActive.Store(true)
	backpressure := func(_ context.Context) bool { return backpressureActive.Load() }

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, backpressure)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	repo.AssertNotCalled(t, "ClaimGraphUpsertCandidates", mock.Anything, mock.Anything, mock.Anything)
}

func TestGraphUpsertProcessor_BackpressureReleasedResumesProcessing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	batch := []ops.QueuedIngestionLog{newGraphUpsertCandidate(1, "dcim.site", nil)}

	var backpressureActive atomic.Bool
	backpressureActive.Store(true)
	backpressure := func(_ context.Context) bool { return backpressureActive.Load() }

	claimed := make(chan struct{}, 1)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return(batch, nil).Once().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()
	repo.On("MarkGraphUpserted", mock.Anything, []int32{1}).Return(nil).Maybe()
	mockMetrics.On("RecordGraphUpsert", mock.Anything, true, "dcim.site", mock.AnythingOfType("float64")).Maybe()

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, backpressure)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	time.Sleep(100 * time.Millisecond)
	repo.AssertNotCalled(t, "ClaimGraphUpsertCandidates", mock.Anything, mock.Anything, mock.Anything)

	backpressureActive.Store(false)

	select {
	case <-claimed:
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not resume after backpressure released")
	}

	cancel()
	<-done
}

func TestGraphUpsertProcessor_PerWorkerFactoryIsInvokedOncePerWorker(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe()

	var factoryCalls atomic.Int32
	factory := func() reconciler.GraphEntityUpserter {
		factoryCalls.Add(1)
		return &fakeUpserter{}
	}

	cfg := reconciler.Config{GraphUpsertProcessorConcurrency: 3}
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), cfg, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	// Give all three workers a chance to spin up before cancelling.
	assert.Eventually(t, func() bool { return factoryCalls.Load() == 3 }, 2*time.Second, 20*time.Millisecond,
		"factory should be invoked exactly once per worker — graph.Service is not safe for concurrent use")

	cancel()
	<-done
}

func TestGraphUpsertProcessor_DefaultsFillUnsetConfigValues(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockMetrics := mocks.NewMetrics(t)

	// With BatchSize/MaxAttempts unset, the processor must fall back to the
	// package defaults — 100/5 — when issuing claim queries.
	claimed := make(chan struct{}, 1)
	repo.On("ClaimGraphUpsertCandidates", mock.Anything, int32(100), int32(5)).
		Return([]ops.QueuedIngestionLog{}, nil).Maybe().Run(func(_ mock.Arguments) {
		select {
		case claimed <- struct{}{}:
		default:
		}
	})

	factory := func() reconciler.GraphEntityUpserter { return &fakeUpserter{} }
	p := reconciler.NewGraphUpsertProcessor(newGraphUpsertTestLogger(), reconciler.Config{}, repo, mockMetrics, factory, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	select {
	case <-claimed:
	case <-time.After(2 * time.Second):
		t.Fatal("claim never called with default args")
	}
	cancel()
	<-done
}
