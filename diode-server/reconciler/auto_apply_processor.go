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
// It is mutually exclusive with IngestionLogProcessor per tenant - the orchestrator
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
//
// Shutdown is graceful: see IngestionLogProcessor for the workCtx/pollCtx split
// and defaultProcessorGracefulShutdownTimeout. The drain window is what keeps
// in-flight /bulk-plan-apply calls from leaving rows orphaned in APPLYING on a
// tenant-config restart.
type AutoApplyProcessor struct {
	config       Config
	logger       *slog.Logger
	ops          IngestionProcessorOps
	repo         Repository
	metrics      Metrics
	backpressure BackpressureFunc

	// mx protects the lifecycle fields below: workCancel, pollCancel, done.
	// Set by Start, read by Stop/shutdown/watchParent.
	mx         sync.Mutex
	workCancel context.CancelFunc
	pollCancel context.CancelFunc
	done       chan struct{}

	batchSize int32
}

// NewAutoApplyProcessor creates a new auto-apply processor. backpressure may be
// nil; when supplied it yields the poll loop (sleeps one idle interval) while
// the caller-defined condition holds - typically used to let the Redis consume
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

	workCtx, workCancel := context.WithCancel(context.WithoutCancel(ctx))
	pollCtx, pollCancel := context.WithCancel(workCtx)
	done := make(chan struct{})

	p.mx.Lock()
	p.workCancel = workCancel
	p.pollCancel = pollCancel
	p.done = done
	p.mx.Unlock()

	var watchWG sync.WaitGroup
	watchWG.Add(1)
	go func() {
		defer watchWG.Done()
		p.watchParent(ctx, done)
	}()

	err := p.pollLoop(pollCtx, workCtx)
	workCancel()
	close(done)
	watchWG.Wait()
	return err
}

// Stop stops the auto-apply processor. Waits up to
// defaultProcessorGracefulShutdownTimeout for in-flight /bulk-plan-apply calls
// to drain (so APPLYING rows reach a terminal state), then force-cancels.
func (p *AutoApplyProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.shutdown()
	return nil
}

func (p *AutoApplyProcessor) shutdown() {
	p.mx.Lock()
	pollCancel := p.pollCancel
	workCancel := p.workCancel
	done := p.done
	p.mx.Unlock()

	if pollCancel == nil || done == nil {
		return
	}
	pollCancel()

	select {
	case <-done:
	case <-time.After(defaultProcessorGracefulShutdownTimeout):
		p.logger.Warn("auto-apply processor did not drain within timeout, forcing cancel",
			"name", p.Name(), "timeout", defaultProcessorGracefulShutdownTimeout)
		if workCancel != nil {
			workCancel()
		}
		<-done
	}
}

func (p *AutoApplyProcessor) watchParent(parentCtx context.Context, done <-chan struct{}) {
	select {
	case <-parentCtx.Done():
		p.shutdown()
	case <-done:
	}
}

func (p *AutoApplyProcessor) pollLoop(pollCtx, workCtx context.Context) error {
	concurrency := max(p.config.AutoApplyProcessorConcurrency, 1)

	if concurrency == 1 {
		p.pollWorker(pollCtx, workCtx)
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			p.pollWorker(pollCtx, workCtx)
		}()
	}
	wg.Wait()
	return nil
}

func (p *AutoApplyProcessor) pollWorker(pollCtx, workCtx context.Context) {
	for {
		// Exit between batches if graceful shutdown has been signaled.
		select {
		case <-pollCtx.Done():
			return
		default:
		}

		// Yield to the Redis consume loop when it's behind on draining.
		// AutoApply both writes to NetBox and reads Postgres heavily, so
		// during burst ingest we want consume loop to clear the stream
		// before AutoApply contends for the same resources.
		if p.backpressure != nil && p.backpressure(pollCtx) {
			select {
			case <-pollCtx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		// Claim + process run on workCtx so they survive pollCtx cancel.
		// Only force-cancellation of workCtx (after drain timeout) aborts
		// them, which is what prevents APPLYING orphans on tenant restart.
		batch, err := p.repo.ClaimQueuedForAutoApply(workCtx, p.batchSize)
		if err != nil {
			p.logger.Error("failed to claim queued ingestion logs for auto-apply", "error", err)
			select {
			case <-pollCtx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		if len(batch) == 0 {
			select {
			case <-pollCtx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		p.processBatch(workCtx, batch)

		if len(batch) < int(p.batchSize) {
			select {
			case <-pollCtx.Done():
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
