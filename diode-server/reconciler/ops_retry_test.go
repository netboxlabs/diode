package reconciler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	pluginmocks "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func retryTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func retryTestItems() []ops.QueuedIngestionLog {
	return []ops.QueuedIngestionLog{
		{ID: 1, IngestionLog: &reconcilerpb.IngestionLog{Entity: &diodepb.Entity{}, ObjectType: "dcim.device"}},
	}
}

// TestBulkPlanApplyTransportErrorMarksForRetry verifies that a whole-batch
// transport failure routes every entity through MarkIngestionLogRetry with the
// configured policy (seconds derived from the RetryPolicy durations) rather than
// a bare FAILED write, and still persists a failure-placeholder change set.
func TestBulkPlanApplyTransportErrorMarksForRetry(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	policy := reconciler.RetryPolicy{Enabled: true, MaxRetries: 7, BaseBackoff: 45 * time.Second, MaxBackoff: 10 * time.Minute}
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, retryTestLogger(), nil, reconciler.WithRetryPolicy(policy))

	mockNetBoxClient.EXPECT().BulkPlanApply(mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("netbox unavailable"))

	// 45s base, 600s (10m) cap -> passed to MarkIngestionLogRetry as seconds.
	mockRepository.EXPECT().
		MarkIngestionLogRetry(mock.Anything, int32(1), int32(7), int64(45), int64(600), mock.Anything).
		Return(nil).Once()
	mockRepository.EXPECT().
		CreateChangeSet(mock.Anything, mock.Anything, int32(1)).
		Return(int32Ptr(99), nil).Once()

	results := opsInstance.BulkPlanApply(context.Background(), retryTestItems(), "")

	require.Len(t, results, 1)
	require.Error(t, results[0].PlanErr)
}

// TestBulkPlanApplyApplyErrorMarksForRetry verifies that an apply-phase failure
// (plan ok, apply failed) persists the change set and then advances retry
// accounting via MarkIngestionLogRetry.
func TestBulkPlanApplyApplyErrorMarksForRetry(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	policy := reconciler.RetryPolicy{Enabled: true, MaxRetries: 3, BaseBackoff: 30 * time.Second, MaxBackoff: time.Hour}
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, retryTestLogger(), nil, reconciler.WithRetryPolicy(policy))

	resp := &netboxdiodeplugin.BulkPlanApplyResponse{
		Results: []netboxdiodeplugin.BulkPlanApplyResult{
			{
				ID: "1",
				ChangeSet: &netboxdiodeplugin.ChangeSet{
					ID:      "cs-1",
					Changes: []netboxdiodeplugin.Change{{ID: "c1", ChangeType: "update", ObjectType: "dcim.device"}},
				},
				Errors: &netboxdiodeplugin.BulkPlanApplyErrors{Apply: json.RawMessage(`"constraint violation"`)},
			},
		},
	}
	mockNetBoxClient.EXPECT().BulkPlanApply(mock.Anything, mock.Anything).Return(resp, nil)

	// DefaultLimits.MaxChangeSetsPerIngestionLog() == 5.
	mockRepository.EXPECT().
		BulkPersistChangeSets(mock.Anything, mock.Anything, int32(5)).
		Return([]ops.BulkPersistResult{{IngestionLogID: 1, ChangeSetID: int32Ptr(5)}}, nil).Once()
	mockRepository.EXPECT().
		MarkIngestionLogRetry(mock.Anything, int32(1), int32(3), int64(30), int64(3600), mock.Anything).
		Return(nil).Once()

	results := opsInstance.BulkPlanApply(context.Background(), retryTestItems(), "")

	require.Len(t, results, 1)
	require.NoError(t, results[0].PlanErr)
	require.Error(t, results[0].ApplyErr)
}

// TestBulkPlanApplyRetryDisabledWritesFailed verifies that with retry disabled
// (the default), a failed apply is written as terminal FAILED via
// UpdateIngestionLogStateWithError — exactly as before the feature — and
// MarkIngestionLogRetry is never called.
func TestBulkPlanApplyRetryDisabledWritesFailed(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	// Enabled defaults to false.
	policy := reconciler.RetryPolicy{MaxRetries: 5, BaseBackoff: 30 * time.Second, MaxBackoff: time.Hour}
	opsInstance := reconciler.NewOps(mockRepository, mockNetBoxClient, retryTestLogger(), nil, reconciler.WithRetryPolicy(policy))

	mockNetBoxClient.EXPECT().BulkPlanApply(mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("netbox unavailable"))

	mockRepository.EXPECT().
		UpdateIngestionLogStateWithError(mock.Anything, int32(1), reconcilerpb.State_FAILED, mock.Anything).
		Return(nil).Once()
	mockRepository.EXPECT().
		CreateChangeSet(mock.Anything, mock.Anything, int32(1)).
		Return(int32Ptr(99), nil).Once()

	results := opsInstance.BulkPlanApply(context.Background(), retryTestItems(), "")

	require.Len(t, results, 1)
	require.Error(t, results[0].PlanErr)
	mockRepository.AssertNotCalled(t, "MarkIngestionLogRetry")
}

// TestConfigRetryPolicy verifies the config->RetryPolicy mapping, including the
// Enabled gate (off unless ENABLE_FAILED_RETRY is set) and seconds->durations.
func TestConfigRetryPolicy(t *testing.T) {
	cfg := reconciler.Config{
		EnableFailedRetry:             true,
		FailedRetryMaxRetries:         9,
		FailedRetryBaseBackoffSeconds: 15,
		FailedRetryMaxBackoffSeconds:  1800,
	}
	p := cfg.RetryPolicy()
	require.True(t, p.Enabled)
	require.Equal(t, int32(9), p.MaxRetries)
	require.Equal(t, 15*time.Second, p.BaseBackoff)
	require.Equal(t, 30*time.Minute, p.MaxBackoff)

	require.False(t, reconciler.Config{}.RetryPolicy().Enabled)
}
