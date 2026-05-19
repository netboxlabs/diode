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

const (
	defaultIngestionLogPollInterval = 100 * time.Millisecond
	defaultIngestionLogIdleInterval = time.Second
	defaultIngestionLogBatchSize    = 100
)

// defaultProcessorGracefulShutdownTimeout bounds how long Stop / parent-ctx
// cancellation waits for in-flight batches to drain before force-canceling
// the work ctx and aborting any still-running HTTP/DB calls. Without this
// window, a canceled batch leaves ingestion_logs rows orphaned in the
// in-progress state (OPEN for plan-only, APPLYING for auto-apply) until
// the next pod restart calls ResetApplyingIngestionLogs.
//
// Var (not const) so package-internal tests can shrink it.
var defaultProcessorGracefulShutdownTimeout = 30 * time.Second

// BackpressureFunc returns true when the processor should back off to reduce
// Postgres contention (e.g., during heavy Redis stream drain).
type BackpressureFunc func(ctx context.Context) bool

// IngestionLogProcessor polls ingestion_logs for QUEUED rows and generates change sets.
// Applying change sets is handled by a separate ApplyProcessor.
//
// Lifecycle invariants (shared with AutoApplyProcessor):
//
//   - workCtx is detached from the parent ctx so an in-flight batch finishes
//     its HTTP round-trip + DB state updates even when the parent is canceled
//     (e.g., orchestrator restarting this processor for a tenant config update).
//   - pollCtx is a child of workCtx and is the graceful-shutdown signal:
//     canceling it tells workers to exit between batches but lets the current
//     batch run to completion.
//   - Stop() and parent-ctx cancellation both trigger graceful drain via
//     pollCtx cancel, wait up to defaultProcessorGracefulShutdownTimeout for
//     workers to exit naturally, then force-cancel workCtx to abort anything
//     still in flight.
type IngestionLogProcessor struct {
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

// NewIngestionLogProcessor creates a new ingestion log processor.
func NewIngestionLogProcessor(logger *slog.Logger, cfg Config, repo Repository, ops IngestionProcessorOps, metrics Metrics, backpressure BackpressureFunc) *IngestionLogProcessor {
	batchSize := cfg.IngestionLogProcessorBatchSize
	if batchSize <= 0 {
		batchSize = defaultIngestionLogBatchSize
	}
	return &IngestionLogProcessor{
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
func (p *IngestionLogProcessor) Name() string {
	return "reconciler-ingestion-log-processor"
}

// Start starts polling for QUEUED ingestion logs and processing them.
func (p *IngestionLogProcessor) Start(ctx context.Context) error {
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

// Stop stops the ingestion log processor. Waits up to
// defaultProcessorGracefulShutdownTimeout for in-flight batches to drain, then
// force-cancels the work context.
func (p *IngestionLogProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.shutdown()
	return nil
}

// shutdown is the shared graceful-drain path used by both the explicit Stop()
// call and the parent-ctx cancellation path in watchParent.
func (p *IngestionLogProcessor) shutdown() {
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
		p.logger.Warn("ingestion log processor did not drain within timeout, forcing cancel",
			"name", p.Name(), "timeout", defaultProcessorGracefulShutdownTimeout)
		if workCancel != nil {
			workCancel()
		}
		<-done
	}
}

// watchParent mirrors parent-ctx cancellation onto the shutdown path so that
// orchestrator-driven restarts get the same graceful drain as a direct Stop().
func (p *IngestionLogProcessor) watchParent(parentCtx context.Context, done <-chan struct{}) {
	select {
	case <-parentCtx.Done():
		p.shutdown()
	case <-done:
	}
}

func (p *IngestionLogProcessor) pollLoop(pollCtx, workCtx context.Context) error {
	concurrency := max(p.config.IngestionLogProcessorConcurrency, 1)

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

func (p *IngestionLogProcessor) pollWorker(pollCtx, workCtx context.Context) {
	for {
		// Exit between batches if graceful shutdown has been signaled.
		select {
		case <-pollCtx.Done():
			return
		default:
		}

		if p.backpressure != nil && p.backpressure(pollCtx) {
			select {
			case <-pollCtx.Done():
				return
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		// Claim + process run on workCtx so they survive pollCtx cancel.
		// Only force-cancellation of workCtx (after the drain timeout)
		// aborts them.
		batch, err := p.repo.ClaimQueuedIngestionLogs(workCtx, p.batchSize)
		if err != nil {
			p.logger.Error("failed to claim queued ingestion logs", "error", err)
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

func (p *IngestionLogProcessor) processBatch(ctx context.Context, batch []ops.QueuedIngestionLog) {
	branchID := ""
	if branch, err := p.ops.DefaultBranch(ctx); err == nil && branch != nil {
		branchID = branch.ID
	}

	results := p.ops.BulkPlan(ctx, batch, branchID)

	for i, result := range results {
		item := batch[i]

		attrs := []attribute.KeyValue{
			attribute.String(telemetry.AttributeSDKName, item.IngestionLog.GetSdkName()),
			attribute.String(telemetry.AttributeProducerAppName, item.IngestionLog.GetProducerAppName()),
		}
		metricsCtx := telemetry.ContextWithMetricAttributes(ctx, attrs...)

		if result.Err != nil {
			p.logger.Error("error generating changeset", "error", result.Err, "ingestionLogID", item.ID)
			p.metrics.RecordChangeSetCreate(metricsCtx, false, 0)
			continue
		}
		if result.ChangeSet != nil {
			p.metrics.RecordChangeSetCreate(metricsCtx, true, int64(len(result.ChangeSet.Changes)))
		}
	}
}
