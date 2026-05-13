package reconciler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

// AutoApplyProcessor polls ingestion_logs for QUEUED rows belonging to auto-apply
// tenants and drives them through plan + apply in a single /bulk-plan-apply call.
// It is mutually exclusive with IngestionLogProcessor per tenant — the orchestrator
// starts one or the other based on AUTO_APPLY_CHANGESETS, never both.
//
// State machine per claimed entity:
//
//	QUEUED -> APPLYING -> APPLIED      (plan ok, apply ok)
//	QUEUED -> APPLYING -> FAILED       (plan failed; or plan ok + apply failed)
//	QUEUED -> APPLYING -> NO_CHANGES   (plan ok, empty change_set)
//
// On crash mid-batch, rows stuck in APPLYING are reset to QUEUED at startup by
// ResetApplyingIngestionLogs so the next worker iteration picks them up.
type AutoApplyProcessor struct {
	config       Config
	logger       *slog.Logger
	ops          IngestionProcessorOps
	repo         Repository
	metrics      Metrics
	backpressure BackpressureFunc
	cancel       context.CancelFunc
	mx           sync.Mutex
	batchSize    int32
}

// NewAutoApplyProcessor creates a new auto-apply processor. backpressure may be
// nil; when supplied it yields the poll loop (sleeps one idle interval) while
// the caller-defined condition holds — typically used to let the Redis consume
// loop catch up before AutoApply starts taking NetBox capacity.
func NewAutoApplyProcessor(logger *slog.Logger, cfg Config, repo Repository, ops IngestionProcessorOps, metrics Metrics, backpressure BackpressureFunc) *AutoApplyProcessor {
	batchSize := cfg.AutoApplyProcessorBatchSize
	if batchSize <= 0 {
		batchSize = defaultIngestionLogBatchSize
	}
	return &AutoApplyProcessor{
		config:       cfg,
		logger:       logger,
		ops:          ops,
		repo:         repo,
		metrics:      metrics,
		backpressure: backpressure,
		batchSize:    batchSize,
	}
}

// Name returns the name of the component.
func (p *AutoApplyProcessor) Name() string {
	return "reconciler-auto-apply-processor"
}

// Start begins polling for QUEUED ingestion logs and processing them with combined plan + apply.
func (p *AutoApplyProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting component", "name", p.Name())
	p.mx.Lock()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.mx.Unlock()
	return p.pollLoop(ctx)
}

// Stop stops the auto-apply processor.
func (p *AutoApplyProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.mx.Lock()
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.mx.Unlock()
	return nil
}

func (p *AutoApplyProcessor) pollLoop(ctx context.Context) error {
	concurrency := max(p.config.AutoApplyProcessorConcurrency, 1)

	if concurrency == 1 {
		p.pollWorker(ctx)
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			p.pollWorker(ctx)
		}()
	}
	wg.Wait()
	return nil
}

func (p *AutoApplyProcessor) pollWorker(ctx context.Context) {
	for {
		// Yield to the Redis consume loop when it's behind on draining.
		// AutoApply both writes to NetBox and reads Postgres heavily, so
		// during burst ingest we want consume loop to clear the stream
		// before AutoApply contends for the same resources.
		if p.backpressure != nil && p.backpressure(ctx) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		batch, err := p.repo.ClaimQueuedForAutoApply(ctx, p.batchSize)
		if err != nil {
			p.logger.Error("failed to claim queued ingestion logs for auto-apply", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		if len(batch) == 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		p.processBatch(ctx, batch)

		if len(batch) < int(p.batchSize) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultIngestionLogPollInterval):
			}
		}
	}
}

func (p *AutoApplyProcessor) processBatch(ctx context.Context, batch []ops.QueuedIngestionLog) {
	branchID := ""
	if branch, err := p.ops.DefaultBranch(ctx); err == nil && branch != nil {
		branchID = branch.ID
	}

	results := p.ops.BulkPlanApply(ctx, batch, branchID)

	for i, result := range results {
		item := batch[i]

		attrs := []attribute.KeyValue{
			attribute.String(telemetry.AttributeSDKName, item.IngestionLog.GetSdkName()),
			attribute.String(telemetry.AttributeProducerAppName, item.IngestionLog.GetProducerAppName()),
		}
		metricsCtx := telemetry.ContextWithMetricAttributes(ctx, attrs...)

		switch {
		case result.PlanErr != nil:
			p.logger.Error("plan failed in bulk-plan-apply", "error", result.PlanErr, "ingestionLogID", item.ID)
			p.metrics.RecordChangeSetCreate(metricsCtx, false, 0)
		case result.ApplyErr != nil:
			p.logger.Error("apply failed in bulk-plan-apply", "error", result.ApplyErr, "ingestionLogID", item.ID)
			if result.ChangeSet != nil {
				p.metrics.RecordChangeSetCreate(metricsCtx, true, int64(len(result.ChangeSet.Changes)))
			}
			p.metrics.RecordChangeSetApply(metricsCtx, false, 0)
		default:
			if result.ChangeSet != nil {
				changes := int64(len(result.ChangeSet.Changes))
				p.metrics.RecordChangeSetCreate(metricsCtx, true, changes)
				if changes > 0 {
					p.metrics.RecordChangeSetApply(metricsCtx, true, changes)
				}
			}
		}
	}
}
