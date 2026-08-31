package reconciler_test

import (
	"context"
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
