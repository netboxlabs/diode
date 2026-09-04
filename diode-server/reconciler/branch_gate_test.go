package reconciler_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

// A cold default-branch cache is indistinguishable from "no default branch" at
// the point of use: both yield an empty branchID, which plans and applies
// against main. These tests pin the fail-closed behaviour — while the cache is
// cold the processors must not claim work at all, so nothing can be placed on
// the wrong branch.
//
// Regression test for the boot race where the initial default-branch fetch
// loses to OAuth client bootstrap and returns 403, leaving the cache cold.

func TestIngestionLogProcessor_ColdBranchCacheClaimsNothing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	mockOps.On("HasBranchLoaded").Return(false)

	// No ClaimQueuedIngestionLogs expectation is registered, so the mock fails
	// the test if the processor claims anything while the cache is cold.

	p := reconciler.NewIngestionLogProcessor(
		newIngestionLogProcessorTestLogger(),
		reconciler.Config{},
		repo,
		mockOps,
		mockMetrics,
		nil,
	)

	runBriefly(t, p.Start)

	repo.AssertNotCalled(t, "ClaimQueuedIngestionLogs")
	mockOps.AssertNotCalled(t, "BulkPlan")
	mockOps.AssertNotCalled(t, "BulkPlanApply")
}

func TestAutoApplyProcessor_ColdBranchCacheClaimsNothing(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	mockOps.On("HasBranchLoaded").Return(false)

	p := reconciler.NewAutoApplyProcessor(
		newIngestionLogProcessorTestLogger(),
		reconciler.Config{},
		repo,
		mockOps,
		mockMetrics,
		nil,
	)

	runBriefly(t, p.Start)

	repo.AssertNotCalled(t, "ClaimQueuedForAutoApply")
	mockOps.AssertNotCalled(t, "BulkPlanApply")
}

// runBriefly starts a processor, gives it long enough to run several poll
// iterations, then cancels and waits for a clean exit.
func runBriefly(t *testing.T, start func(context.Context) error) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- start(ctx) }()

	time.Sleep(250 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}
}

// lockedBuffer collects log output safely across poll workers.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// The gate is re-evaluated on every poll iteration by every worker, and a
// misconfigured client keeps the cache cold indefinitely, so the warning has
// to be emitted once per cold period rather than once per iteration.
func TestColdBranchCacheWarnsOncePerColdPeriod(t *testing.T) {
	repo := mocks.NewRepository(t)
	mockOps := mocks.NewIngestionProcessorOps(t)
	mockMetrics := mocks.NewMetrics(t)

	mockOps.On("HasBranchLoaded").Return(false)

	out := &lockedBuffer{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))

	p := reconciler.NewIngestionLogProcessor(
		logger,
		reconciler.Config{},
		repo,
		mockOps,
		mockMetrics,
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Start(ctx) }()

	// Several idle intervals, so an unguarded warning would repeat.
	time.Sleep(defaultIdleIntervalsForWarnTest)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("processor did not exit")
	}

	n := strings.Count(out.String(), "default branch not yet known")
	require.Equal(t, 1, n, "cold-cache warning should be logged once per cold period, got %d", n)
}

// Long enough for several poll iterations at the idle interval.
const defaultIdleIntervalsForWarnTest = 3200 * time.Millisecond
