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

// ApplyProcessor polls ingestion_logs for OPEN rows (with changesets loaded from the DB)
// and applies them to NetBox independently of the plan processor.
type ApplyProcessor struct {
	config    Config
	logger    *slog.Logger
	ops       IngestionProcessorOps
	repo      Repository
	metrics   Metrics
	cancel    context.CancelFunc
	mx        sync.Mutex
	batchSize int32
}

// NewApplyProcessor creates a new apply processor.
func NewApplyProcessor(logger *slog.Logger, cfg Config, repo Repository, ops IngestionProcessorOps, metrics Metrics) *ApplyProcessor {
	batchSize := cfg.ApplyProcessorBatchSize
	if batchSize <= 0 {
		batchSize = defaultIngestionLogBatchSize
	}
	return &ApplyProcessor{
		config:    cfg,
		logger:    logger,
		ops:       ops,
		repo:      repo,
		metrics:   metrics,
		batchSize: batchSize,
	}
}

// Name returns the name of the component.
func (p *ApplyProcessor) Name() string {
	return "reconciler-apply-processor"
}

// Start starts polling for OPEN ingestion logs and applying them.
func (p *ApplyProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting component", "name", p.Name())
	p.mx.Lock()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.mx.Unlock()
	return p.pollLoop(ctx)
}

// Stop stops the apply processor.
func (p *ApplyProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.mx.Lock()
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.mx.Unlock()
	return nil
}

func (p *ApplyProcessor) pollLoop(ctx context.Context) error {
	concurrency := max(p.config.ApplyProcessorConcurrency, 1)

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

func (p *ApplyProcessor) pollWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("apply processor exiting poll loop on request")
			return
		default:
		}

		batch, err := p.repo.ClaimOpenIngestionLogs(ctx, p.batchSize)
		if err != nil {
			p.logger.Error("failed to claim open ingestion logs", "error", err)
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

func (p *ApplyProcessor) processBatch(ctx context.Context, batch []ops.OpenIngestionLog) {
	branchID := ""
	if branch, err := p.ops.DefaultBranch(ctx); err == nil && branch != nil {
		branchID = branch.ID
	}

	applyItems := make([]ops.BulkApplyItem, len(batch))
	for i, item := range batch {
		applyItems[i] = ops.BulkApplyItem{
			IngestionLogID: item.ID,
			IngestionLog:   item.IngestionLog,
			ChangeSetID:    item.ChangeSetID,
			ChangeSet:      item.ChangeSet,
		}
	}

	applyResults := p.ops.BulkApplyChangeSets(ctx, applyItems, branchID)

	for j, ar := range applyResults {
		item := batch[j]

		attrs := []attribute.KeyValue{
			attribute.String(telemetry.AttributeSDKName, item.IngestionLog.GetSdkName()),
			attribute.String(telemetry.AttributeProducerAppName, item.IngestionLog.GetProducerAppName()),
		}
		metricsCtx := telemetry.ContextWithMetricAttributes(ctx, attrs...)

		if ar.Err != nil {
			p.logger.Error("error applying changeset", "error", ar.Err, "ingestionLogID", item.ID)
			p.metrics.RecordChangeSetApply(metricsCtx, false, 0)
			continue
		}
		p.metrics.RecordChangeSetApply(metricsCtx, true, int64(len(item.ChangeSet.Changes)))
	}
}
