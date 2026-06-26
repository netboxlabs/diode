package reconciler

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/graph"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	defaultGraphUpsertPollInterval = 100 * time.Millisecond
	defaultGraphUpsertIdleInterval = time.Second
	defaultGraphUpsertBatchSize    = 100
	defaultGraphUpsertMaxAttempts  = 5
)

// GraphEntityUpserter is the subset of *graph.Service that the
// GraphUpsertProcessor depends on. It exists so the processor can be tested
// with a small mock rather than wiring up a real graph repository.
//
// Implementations are not required to be safe for concurrent use; the
// processor builds a fresh instance per worker via GraphServiceFactory to
// match graph.Service's documented serialization requirement.
type GraphEntityUpserter interface {
	UpsertEntity(ctx context.Context, entity *diodepb.Entity, requestMetadata ...map[string]any) (*graph.Node, error)
}

// GraphServiceFactory returns a GraphEntityUpserter for a single worker
// goroutine. graph.Service holds per-call mutable state on the receiver and
// is documented as not safe for concurrent use, so the processor calls the
// factory once per worker on startup.
type GraphServiceFactory func() GraphEntityUpserter

// GraphUpsertProcessor moves entities from ingestion_logs into the graph
// database. It runs independently of the ingestion state machine — graph
// rows are tracked via the graph_upserted_at/graph_upsert_attempts columns
// rather than the ingestion state column, so plan and apply continue to
// drain the inbox even if the graph DB is unavailable.
//
// On crash mid-batch, rows with a non-null graph_upsert_claimed_at are
// returned to the pool at startup by ResetClaimedGraphUpserts so the next
// worker iteration picks them up.
type GraphUpsertProcessor struct {
	config         Config
	logger         *slog.Logger
	repo           Repository
	metrics        Metrics
	serviceFactory GraphServiceFactory
	backpressure   BackpressureFunc
	cancel         context.CancelFunc
	mx             sync.Mutex
	batchSize      int32
	maxAttempts    int32
}

// NewGraphUpsertProcessor creates a new graph-upsert processor.
//
// serviceFactory must return a fresh GraphEntityUpserter per call — each
// worker goroutine builds its own to avoid contending on the per-call
// mutable state held inside graph.Service.
//
// backpressure may be nil; when supplied, the poll loop yields one idle
// interval while the condition holds, matching the existing processors.
func NewGraphUpsertProcessor(logger *slog.Logger, cfg Config, repo Repository, metrics Metrics, serviceFactory GraphServiceFactory, backpressure BackpressureFunc) *GraphUpsertProcessor {
	batchSize := cfg.GraphUpsertProcessorBatchSize
	if batchSize <= 0 {
		batchSize = defaultGraphUpsertBatchSize
	}
	maxAttempts := cfg.GraphUpsertProcessorMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultGraphUpsertMaxAttempts
	}
	return &GraphUpsertProcessor{
		config:         cfg,
		logger:         logger,
		repo:           repo,
		metrics:        metrics,
		serviceFactory: serviceFactory,
		backpressure:   backpressure,
		batchSize:      batchSize,
		maxAttempts:    maxAttempts,
	}
}

// Name returns the name of the component.
func (p *GraphUpsertProcessor) Name() string {
	return "reconciler-graph-upsert-processor"
}

// Start begins polling for un-upserted ingestion logs and processing them.
func (p *GraphUpsertProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting component", "name", p.Name())
	p.mx.Lock()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.mx.Unlock()
	return p.pollLoop(ctx)
}

// Stop stops the graph-upsert processor.
func (p *GraphUpsertProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.mx.Lock()
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.mx.Unlock()
	return nil
}

func (p *GraphUpsertProcessor) pollLoop(ctx context.Context) error {
	concurrency := max(p.config.GraphUpsertProcessorConcurrency, 1)

	if concurrency == 1 {
		p.pollWorker(ctx, p.serviceFactory())
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			p.pollWorker(ctx, p.serviceFactory())
		}()
	}
	wg.Wait()
	return nil
}

func (p *GraphUpsertProcessor) pollWorker(ctx context.Context, svc GraphEntityUpserter) {
	for {
		if p.backpressure != nil && p.backpressure(ctx) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultGraphUpsertIdleInterval):
				continue
			}
		}

		batch, err := p.repo.ClaimGraphUpsertCandidates(ctx, p.batchSize, p.maxAttempts)
		if err != nil {
			p.logger.Error("failed to claim graph-upsert candidates", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultGraphUpsertIdleInterval):
				continue
			}
		}

		if len(batch) == 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultGraphUpsertIdleInterval):
				continue
			}
		}

		p.processBatch(ctx, svc, batch)

		if len(batch) < int(p.batchSize) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultGraphUpsertPollInterval):
			}
		}
	}
}

func (p *GraphUpsertProcessor) processBatch(ctx context.Context, svc GraphEntityUpserter, batch []ops.QueuedIngestionLog) {
	successIDs := make([]int32, 0, len(batch))
	failureIDs := make([]int32, 0)

	for _, item := range batch {
		entity := item.IngestionLog.GetEntity()
		if entity == nil {
			p.logger.Warn("graph upsert skipped: ingestion log has nil entity", "ingestion_log_id", item.ID)
			successIDs = append(successIDs, item.ID)
			continue
		}

		var reqMeta map[string]any
		if len(item.SourceMetadata) > 0 {
			if err := json.Unmarshal(item.SourceMetadata, &reqMeta); err != nil {
				p.logger.Warn("failed to unmarshal request metadata; continuing without it",
					"error", err,
					"ingestion_log_id", item.ID)
				reqMeta = nil
			}
		}

		nodeType := item.IngestionLog.GetObjectType()
		start := time.Now()

		var err error
		if reqMeta != nil {
			_, err = svc.UpsertEntity(ctx, entity, reqMeta)
		} else {
			_, err = svc.UpsertEntity(ctx, entity)
		}
		duration := time.Since(start).Seconds()

		attrs := []attribute.KeyValue{
			attribute.String(telemetry.AttributeSDKName, item.IngestionLog.GetSdkName()),
			attribute.String(telemetry.AttributeProducerAppName, item.IngestionLog.GetProducerAppName()),
		}
		metricsCtx := telemetry.ContextWithMetricAttributes(ctx, attrs...)

		if err != nil {
			p.logger.Warn("graph upsert failed",
				"error", err,
				"ingestion_log_id", item.ID,
				"object_type", nodeType)
			p.metrics.RecordGraphUpsert(metricsCtx, false, nodeType, duration)
			failureIDs = append(failureIDs, item.ID)
			continue
		}

		p.metrics.RecordGraphUpsert(metricsCtx, true, nodeType, duration)
		successIDs = append(successIDs, item.ID)
	}

	if len(successIDs) > 0 {
		if err := p.repo.MarkGraphUpserted(ctx, successIDs); err != nil {
			p.logger.Error("failed to mark graph-upserted rows; they will be re-claimed", "error", err, "ids", len(successIDs))
		}
	}
	if len(failureIDs) > 0 {
		if err := p.repo.ReleaseGraphUpsertClaims(ctx, failureIDs); err != nil {
			p.logger.Error("failed to release graph-upsert claims; rows will recover on the next ResetClaimedGraphUpserts cycle", "error", err, "ids", len(failureIDs))
		}
	}
}
