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

// BackpressureFunc returns true when the processor should back off to reduce
// Postgres contention (e.g., during heavy Redis stream drain).
type BackpressureFunc func(ctx context.Context) bool

// IngestionLogProcessor polls ingestion_logs for QUEUED rows and generates change sets.
// Applying change sets is handled by a separate ApplyProcessor.
type IngestionLogProcessor struct {
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
	p.mx.Lock()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.mx.Unlock()
	return p.pollLoop(ctx)
}

// Stop stops the ingestion log processor.
func (p *IngestionLogProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.mx.Lock()
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.mx.Unlock()
	return nil
}

func (p *IngestionLogProcessor) pollLoop(ctx context.Context) error {
	concurrency := max(p.config.IngestionLogProcessorConcurrency, 1)

	if concurrency == 1 {
		return p.pollWorker(ctx)
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			_ = p.pollWorker(ctx)
		}()
	}
	wg.Wait()
	return nil
}

func (p *IngestionLogProcessor) pollWorker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("ingestion log processor exiting poll loop on request")
			return nil
		default:
		}

		if p.backpressure != nil && p.backpressure(ctx) {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		batch, err := p.repo.ClaimQueuedIngestionLogs(ctx, p.batchSize)
		if err != nil {
			p.logger.Error("failed to claim queued ingestion logs", "error", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		if len(batch) == 0 {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(defaultIngestionLogIdleInterval):
				continue
			}
		}

		p.processBatch(ctx, batch)

		if len(batch) < int(p.batchSize) {
			select {
			case <-ctx.Done():
				return nil
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

	results := p.ops.BulkGenerateChangeSets(ctx, batch, branchID)

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
