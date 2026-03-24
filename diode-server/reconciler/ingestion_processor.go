package reconciler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/cloudflare/backoff"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/gen/netbox"
	"github.com/netboxlabs/diode/diode-server/graph"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
	"github.com/netboxlabs/diode/diode-server/sentry"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	// DefaultRedisStreamID is the default redis stream id for ingestion
	DefaultRedisStreamID = "diode.v1.ingest-stream"

	// DefaultRedisConsumerGroup is the default redis consumer group for ingestion
	DefaultRedisConsumerGroup = "diode-reconciler"

	// RedisIngestEntityIndexName is the name of the redis index for ingest entities
	RedisIngestEntityIndexName = "ingest-entity"

	// RedisConsumerGroupExistsErrMsg is the error message returned by the redis client when the consumer group already exists
	RedisConsumerGroupExistsErrMsg = "BUSYGROUP Consumer Group name already exists"

	// ReclaimMinIdle is the minimum idle time before a pending message can be reclaimed.
	// Prevents stealing messages from live consumers during rolling upgrades.
	ReclaimMinIdle = 2 * time.Minute

	// ReclaimInterval is how often to check for pending messages in the main loop.
	ReclaimInterval = 30 * time.Second

	// ReadBatchSize is the maximum number of messages to read per XReadGroup call.
	ReadBatchSize = 100

	// ReclaimBatchSize is the maximum number of messages to reclaim per cycle.
	ReclaimBatchSize = 100

	// MaxReclaimRetries is the number of times a message can fail reclaim before being discarded.
	MaxReclaimRetries = 3

	// DefaultStreamWorkerCount is the default number of concurrent stream message workers.
	DefaultStreamWorkerCount = 4

	// DefaultStreamHeartbeatInterval is the default interval for the in-flight message heartbeat.
	DefaultStreamHeartbeatInterval = 30 * time.Second
)

// RedisClient is an interface that represents the methods used from redis.Client
type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
	XAutoClaim(ctx context.Context, a *redis.XAutoClaimArgs) *redis.XAutoClaimCmd
	XClaim(ctx context.Context, a *redis.XClaimArgs) *redis.XMessageSliceCmd
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Pipeline() redis.Pipeliner
}

// IngestionProcessor processes ingested data
type IngestionProcessor struct {
	Config             Config
	logger             *slog.Logger
	hostname           string
	redisClient        RedisClient
	redisStreamClient  RedisClient
	redisStreamID      string
	redisConsumerGroup string
	ops                IngestionProcessorOps
	metrics            Metrics
	cancel             context.CancelFunc
	mx                 sync.Mutex
	graphService       *graph.Service // nil when ENABLE_GRAPH_DB is false
	reclaimFailures    map[string]int // tracks per-message reclaim failure counts for poison message detection
	reclaimCursor      string         // XAUTOCLAIM cursor to resume scanning the PEL across cycles
	// inFlightMu protects inFlight
	inFlightMu sync.Mutex
	inFlight   map[string]struct{} // message IDs currently being processed (for heartbeat)
	workerWg   sync.WaitGroup      // tracks in-flight worker goroutines for graceful shutdown
}

// IngestionLogToProcess represents an ingestion log to process
type IngestionLogToProcess struct {
	ingestionLogID int32
	ingestionLog   *reconcilerpb.IngestionLog
	changeSetID    int32
	changeSet      *changeset.ChangeSet
	branchID       string // the branch ID for this ingestion log (empty string means main branch)
}

// IngestionProcessorOps represents the basic operations that the ingestion processor performs
type IngestionProcessorOps interface {
	CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*ops.CreateIngestionLogResult, error)
	BulkCreateIngestionLogs(ctx context.Context, ingestionLogs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) ([]*ops.CreateIngestionLogResult, error)
	GenerateChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, branchID string) (*int32, *changeset.ChangeSet, error)
	ApplyChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, changeSetID int32, changeSet *changeset.ChangeSet) error
	DefaultBranch(ctx context.Context) (*netboxdiodeplugin.Branch, error)
	RefreshDefaultBranch(ctx context.Context) (*netboxdiodeplugin.Branch, error)
}

// ProcessorOption is a functional option for configuring IngestionProcessor
type ProcessorOption func(*IngestionProcessor)

// WithGraphService sets the graph.Service for graph-based entity extraction.
// When set, entities are also stored in the graph database for relationship tracking.
// Pass nil to disable graph extraction.
func WithGraphService(svc *graph.Service) ProcessorOption {
	return func(p *IngestionProcessor) {
		p.graphService = svc
	}
}

// NewIngestionProcessor creates a new ingestion processor
func NewIngestionProcessor(_ context.Context, logger *slog.Logger, cfg Config, redisClient, redisStreamClient RedisClient, redisStreamID string, redisConsumerGroup string, ops IngestionProcessorOps, metrics Metrics, opts ...ProcessorOption) (*IngestionProcessor, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	component := &IngestionProcessor{
		Config:             cfg,
		logger:             logger,
		hostname:           hostname,
		redisClient:        redisClient,
		redisStreamClient:  redisStreamClient,
		redisStreamID:      redisStreamID,
		redisConsumerGroup: redisConsumerGroup,
		ops:                ops,
		metrics:            metrics,
		reclaimFailures:    make(map[string]int),
		reclaimCursor:      "0-0",
		inFlight:           make(map[string]struct{}),
	}

	// Apply functional options
	for _, opt := range opts {
		opt(component)
	}

	return component, nil
}

// Name returns the name of the component
func (p *IngestionProcessor) Name() string {
	return "reconciler-ingestion-processor"
}

// Start starts the component
func (p *IngestionProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting component", "name", p.Name())
	p.mx.Lock()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.mx.Unlock()
	return p.consumeIngestionStream(ctx, p.redisStreamID, p.redisConsumerGroup, fmt.Sprintf("%s-%s", p.redisConsumerGroup, p.hostname))
}

// Stop stops the component
func (p *IngestionProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	p.mx.Lock()
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.mx.Unlock()
	redisClientErr := p.redisClient.Close()
	redisStreamErr := p.redisStreamClient.Close()

	return errors.Join(redisStreamErr, redisClientErr)
}

func (p *IngestionProcessor) trackInFlight(msgID string) {
	p.inFlightMu.Lock()
	defer p.inFlightMu.Unlock()
	p.inFlight[msgID] = struct{}{}
}

func (p *IngestionProcessor) untrackInFlight(msgID string) {
	p.inFlightMu.Lock()
	defer p.inFlightMu.Unlock()
	delete(p.inFlight, msgID)
}

func (p *IngestionProcessor) getInFlightIDs() []string {
	p.inFlightMu.Lock()
	defer p.inFlightMu.Unlock()
	ids := make([]string, 0, len(p.inFlight))
	for id := range p.inFlight {
		ids = append(ids, id)
	}
	return ids
}

// runStreamHeartbeat periodically XCLAIMs all in-flight message IDs to reset their idle time,
// preventing XAUTOCLAIM from other consumers from stealing messages being actively processed.
func (p *IngestionProcessor) runStreamHeartbeat(ctx context.Context, stream, group, consumer string) {
	interval := time.Duration(p.Config.StreamHeartbeatInterval) * time.Second
	if interval <= 0 {
		interval = DefaultStreamHeartbeatInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ids := p.getInFlightIDs()
			if len(ids) == 0 {
				continue
			}
			p.logger.Debug("heartbeat: refreshing in-flight message idle times", "count", len(ids))
			if err := p.redisStreamClient.XClaim(ctx, &redis.XClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: consumer,
				MinIdle:  0,
				Messages: ids,
			}).Err(); err != nil {
				p.logger.Warn("heartbeat: failed to refresh in-flight message idle times", "error", err)
			}
		}
	}
}

// drainPEL reads and processes own pending messages from a prior crash/restart
// using XREADGROUP with "0" before switching to new message consumption.
func (p *IngestionProcessor) drainPEL(ctx context.Context, stream, group, consumer string, sem chan struct{}) error {
	p.logger.Info("draining own PEL before reading new messages")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		streams, err := p.redisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, "0"},
			Count:    ReadBatchSize,
		}).Result()
		if err != nil {
			if strings.Contains(err.Error(), "NOGROUP") {
				return nil
			}
			return fmt.Errorf("draining PEL: %w", err)
		}

		if len(streams) == 0 || len(streams[0].Messages) == 0 {
			p.logger.Info("PEL drained, switching to new messages")
			return nil
		}

		p.logger.Info("draining pending messages from PEL", "count", len(streams[0].Messages))
		for _, msg := range streams[0].Messages {
			p.processStreamMessageAsync(ctx, msg, stream, group, consumer, sem)
		}
	}
}

// processStreamMessageAsync sends a message to the worker pool for concurrent processing.
// It acquires a semaphore token, tracks the message as in-flight, and spawns a goroutine
// that processes the message and ACKs it on success.
func (p *IngestionProcessor) processStreamMessageAsync(ctx context.Context, msg redis.XMessage, stream, group, consumer string, sem chan struct{}) {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	p.trackInFlight(msg.ID)
	p.workerWg.Add(1)

	go func() {
		defer func() {
			<-sem
			p.untrackInFlight(msg.ID)
			p.workerWg.Done()
		}()

		allDone, err := p.handleStreamMessage(ctx, msg)
		if err != nil {
			p.logger.Error("failed to handle stream message", "error", err, "message_id", msg.ID)
			contextMap := map[string]any{
				"redis_stream_msg_id": msg.ID,
				"consumer":            consumer,
				"hostname":            p.hostname,
			}
			sentry.CaptureError(fmt.Errorf("failed to handle stream message: %w", err), nil, "Ingestion stream", contextMap)
			// Leave un-ACK'd in PEL for reclaimer to pick up
			return
		}

		// Wait for async pipelines (GenerateChangeSet/ApplyChangeSet) to complete
		select {
		case <-allDone:
		case <-ctx.Done():
			return
		}

		if err := p.redisStreamClient.XAck(ctx, stream, group, msg.ID).Err(); err != nil {
			p.logger.Error("failed to ACK stream message", "error", err, "message_id", msg.ID)
			return
		}
		p.redisStreamClient.XDel(ctx, stream, msg.ID)
	}()
}

func (p *IngestionProcessor) consumeIngestionStream(ctx context.Context, redisStreamID string, redisConsumerGroup, redisConsumer string) error {
	err := p.redisStreamClient.XGroupCreateMkStream(ctx, redisStreamID, redisConsumerGroup, "$").Err()
	if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
		return err
	}

	// Local cancel ensures heartbeat/reclaimer goroutines are stopped on all exit paths
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerCount := p.Config.StreamWorkerCount
	if workerCount <= 0 {
		workerCount = DefaultStreamWorkerCount
	}
	sem := make(chan struct{}, workerCount)

	// Start heartbeat goroutine — keeps in-flight messages' idle time low
	go p.runStreamHeartbeat(ctx, redisStreamID, redisConsumerGroup, redisConsumer)

	// Start reclaimer goroutine — periodically reclaims messages from dead consumers
	go p.runPELReclaimer(ctx, redisStreamID, redisConsumerGroup, redisConsumer, sem)

	// Phase 1: Drain own PEL from prior crash/restart
	if err := p.drainPEL(ctx, redisStreamID, redisConsumerGroup, redisConsumer, sem); err != nil {
		return err
	}

	// Phase 2: Read new messages
	b := backoff.New(10*time.Second, time.Second)
	drainWorkers := func() {
		p.logger.Debug("ingestion processor: waiting for in-flight workers to finish")
		p.workerWg.Wait()
	}
	for {
		select {
		case <-ctx.Done():
			drainWorkers()
			return nil
		default:
		}

		streams, err := p.redisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    redisConsumerGroup,
			Consumer: redisConsumer,
			Streams:  []string{redisStreamID, ">"},
			Count:    ReadBatchSize,
			Block:    ReclaimInterval,
		}).Result()
		if err != nil || len(streams) == 0 {
			if err != nil && strings.Contains(err.Error(), "NOGROUP") {
				err := p.redisStreamClient.XGroupCreateMkStream(ctx, redisStreamID, redisConsumerGroup, "$").Err()
				if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
					p.logger.Debug("Failed to recreate Redis consumer group.")
				}
			}
			select {
			case <-ctx.Done():
				drainWorkers()
				return nil
			case <-time.After(b.Duration()):
				continue
			}
		}
		b.Reset()

		for _, msg := range streams[0].Messages {
			p.processStreamMessageAsync(ctx, msg, redisStreamID, redisConsumerGroup, redisConsumer, sem)
		}
	}
}

// runPELReclaimer periodically reclaims messages from dead consumers using XAUTOCLAIM.
func (p *IngestionProcessor) runPELReclaimer(ctx context.Context, stream, group, consumer string, sem chan struct{}) {
	ticker := time.NewTicker(ReclaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.reclaimPendingMessages(ctx, stream, group, consumer, sem)
		}
	}
}

// reclaimPendingMessages uses XAUTOCLAIM to claim and process a single batch
// of messages stuck in the consumer group's PEL (Pending Entries List).
// It is called periodically by runPELReclaimer and is non-blocking:
// errors are logged but do not propagate to the caller.
// Poison messages that fail MaxReclaimRetries times are ACK'd and discarded.
func (p *IngestionProcessor) reclaimPendingMessages(ctx context.Context, stream, group, consumer string, sem chan struct{}) {
	p.logger.Debug("checking for pending messages to reclaim", "stream", stream, "group", group)

	cmd := p.redisStreamClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  ReclaimMinIdle,
		Start:    p.reclaimCursor,
		Count:    ReclaimBatchSize,
	})
	messages, nextCursor, err := cmd.Result()
	if err != nil {
		if strings.Contains(err.Error(), "NOGROUP") {
			return
		}
		p.logger.Error("failed to autoclaim pending messages", "error", err)
		return
	}

	// Advance cursor for next cycle; "0-0" means we've scanned the entire PEL, so wrap around
	if nextCursor == "0-0" {
		p.reclaimCursor = "0-0"
	} else {
		p.reclaimCursor = nextCursor
	}

	if len(messages) == 0 {
		return
	}

	p.logger.Info("reclaiming pending messages", "count", len(messages))
	for _, msg := range messages {
		// Acquire semaphore to respect worker pool concurrency limit
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		}

		if !p.processReclaimedMessage(ctx, msg, stream, group, consumer, sem) {
			return
		}
	}
}

// processReclaimedMessage processes a single reclaimed message within the semaphore.
// Returns false if the caller should stop processing (context cancelled).
func (p *IngestionProcessor) processReclaimedMessage(ctx context.Context, msg redis.XMessage, stream, group, consumer string, sem chan struct{}) bool {
	p.trackInFlight(msg.ID)
	defer func() {
		p.untrackInFlight(msg.ID)
		<-sem
	}()

	allDone, err := p.handleStreamMessage(ctx, msg)
	if err != nil {
		p.reclaimFailures[msg.ID]++
		failures := p.reclaimFailures[msg.ID]

		if failures >= MaxReclaimRetries {
			p.logger.Error("discarding poison message after max reclaim retries",
				"message_id", msg.ID, "retries", failures, "error", err)

			contextMap := map[string]any{
				"redis_stream_msg_id": msg.ID,
				"consumer":            consumer,
				"hostname":            p.hostname,
				"retries":             failures,
			}
			sentry.CaptureError(fmt.Errorf("poison message discarded after %d retries: %w", failures, err), nil, "Reclaim poison message", contextMap)

			p.redisStreamClient.XAck(ctx, stream, group, msg.ID)
			delete(p.reclaimFailures, msg.ID)
		} else {
			p.logger.Warn("failed to handle reclaimed message, will retry",
				"message_id", msg.ID, "retries", failures, "error", err)
		}
		return true
	}

	// Wait for async pipelines to complete
	select {
	case <-allDone:
	case <-ctx.Done():
		return false
	}

	// Success — ACK, delete from stream, and clean up tracking
	if err := p.redisStreamClient.XAck(ctx, stream, group, msg.ID).Err(); err != nil {
		p.logger.Error("failed to ACK reclaimed message", "error", err, "message_id", msg.ID)
		return true
	}
	p.redisStreamClient.XDel(ctx, stream, msg.ID)
	delete(p.reclaimFailures, msg.ID)
	return true
}

func (p *IngestionProcessor) handleStreamMessage(ctx context.Context, msg redis.XMessage) (chan struct{}, error) {
	doneChan := make(chan struct{})
	defer close(doneChan)

	// Create attributes for metrics
	attrs := []attribute.KeyValue{
		attribute.String(telemetry.AttributeHostname, p.hostname),
	}
	ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

	reqStr, ok := msg.Values["request"].(string)
	if !ok {
		p.metrics.RecordHandleMessage(ctx, false)
		return doneChan, fmt.Errorf("message %s has missing or invalid request field", msg.ID)
	}
	reqBytes := []byte(reqStr)
	if enc, ok := msg.Values["encoding"].(string); ok && enc == "br" {
		var err error
		reqBytes, err = decompressBrotli(reqBytes)
		if err != nil {
			p.metrics.RecordHandleMessage(ctx, false)
			return doneChan, fmt.Errorf("decompressing request: %w", err)
		}
	}

	ingestReq := &diodepb.IngestRequest{}
	if err := proto.Unmarshal(reqBytes, ingestReq); err != nil {
		p.metrics.RecordHandleMessage(ctx, false)
		return doneChan, err
	}

	// Add request-specific attributes
	attrs = append(attrs,
		attribute.String(telemetry.AttributeSDKName, ingestReq.SdkName),
		attribute.String(telemetry.AttributeSDKVersion, ingestReq.SdkVersion),
		attribute.String(telemetry.AttributeProducerAppName, ingestReq.ProducerAppName),
		attribute.String(telemetry.AttributeProducerAppVersion, ingestReq.ProducerAppVersion),
	)
	ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

	errs := make([]error, 0)

	ingestionTs, err := strconv.Atoi(msg.Values["ingestion_ts"].(string))
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to convert ingestion timestamp: %v", err))
	}

	p.logger.Debug("handling ingest request", "request", ingestReq)

	bufCapacity := 100

	generateIngestionLogChan := make(chan IngestionLogToProcess, bufCapacity)
	generateIngestionLogDoneChan := make(chan struct{})
	var applyChangeSetChan chan IngestionLogToProcess
	applyChangeSetDoneChan := make(chan struct{})

	if p.Config.AutoApplyChangesets {
		applyChangeSetChan = make(chan IngestionLogToProcess, bufCapacity)
	}

	p.GenerateChangeSet(ctx, generateIngestionLogChan, applyChangeSetChan, generateIngestionLogDoneChan)

	if p.Config.AutoApplyChangesets {
		p.ApplyChangeSet(ctx, applyChangeSetChan, applyChangeSetDoneChan)
	} else {
		close(applyChangeSetDoneChan)
	}

	allDone := make(chan struct{})
	go func() {
		<-doneChan
		<-generateIngestionLogDoneChan
		<-applyChangeSetDoneChan
		close(allDone)
	}()

	createIngestionLogsErrs := p.CreateIngestionLogs(ctx, ingestReq, ingestionTs, generateIngestionLogChan)
	if len(createIngestionLogsErrs) > 0 {
		errs = append(errs, createIngestionLogsErrs...)
	}

	if len(errs) > 0 {
		errsStr := make([]string, 0)
		for _, err := range errs {
			errsStr = append(errsStr, err.Error())
		}
		p.logger.Warn("failed to handle ingest request", slog.String("request_id", ingestReq.Id), slog.Any("errors", errsStr))

		contextMap := map[string]any{
			"redis_stream_msg_id": msg.ID,
			"consumer":            fmt.Sprintf("%s-%s", p.redisConsumerGroup, p.hostname),
			"hostname":            p.hostname,
		}
		sentry.CaptureError(fmt.Errorf("failed to handle ingest request: %v", errs), nil, "Ingestion request", contextMap)
		p.metrics.RecordHandleMessage(ctx, false)
	} else {
		p.metrics.RecordHandleMessage(ctx, true)
	}

	return allDone, nil
}

// GenerateChangeSet generates a change set for an ingestion log
func (p *IngestionProcessor) GenerateChangeSet(ctx context.Context, generateChangeSetChan <-chan IngestionLogToProcess, applyChangeSetChan chan<- IngestionLogToProcess, doneChan chan<- struct{}) {
	concurrency := max(p.Config.GenerateChangeSetConcurrency, 1)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for msg := range generateChangeSetChan {
				id, changeSet, err := p.ops.GenerateChangeSet(ctx, msg.ingestionLogID, msg.ingestionLog, msg.branchID)
				if err != nil {
					p.logger.Error("error generating changeset", "error", err)
					p.metrics.RecordChangeSetCreate(ctx, false, 0)
				} else {
					p.metrics.RecordChangeSetCreate(ctx, true, int64(len(changeSet.Changes)))
				}

				if changeSet != nil && len(changeSet.Changes) > 0 {
					if applyChangeSetChan != nil {
						applyChangeSetChan <- IngestionLogToProcess{
							ingestionLogID: msg.ingestionLogID,
							ingestionLog:   msg.ingestionLog,
							changeSetID:    *id,
							changeSet:      changeSet,
							branchID:       msg.branchID,
						}
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		if applyChangeSetChan != nil {
			close(applyChangeSetChan)
		}
		if doneChan != nil {
			doneChan <- struct{}{}
		}
	}()
}

// ApplyChangeSet applies a change set for an ingestion log
func (p *IngestionProcessor) ApplyChangeSet(ctx context.Context, applyChan <-chan IngestionLogToProcess, doneChan chan<- struct{}) {
	concurrency := max(p.Config.ApplyChangeSetConcurrency, 1)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for msg := range applyChan {
				if err := p.ops.ApplyChangeSet(ctx, msg.ingestionLogID, msg.ingestionLog, msg.changeSetID, msg.changeSet); err != nil {
					p.logger.Error("error applying changeset", "error", err)
					p.metrics.RecordChangeSetApply(ctx, false, 0)
				} else {
					p.metrics.RecordChangeSetApply(ctx, true, int64(len(msg.changeSet.Changes)))
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		if doneChan != nil {
			doneChan <- struct{}{}
		}
	}()
}

// CreateIngestionLogs creates ingestion logs for an ingest request using bulk operations
func (p *IngestionProcessor) CreateIngestionLogs(ctx context.Context, ingestReq *diodepb.IngestRequest, ingestionTs int, generateIngestionLogChan chan<- IngestionLogToProcess) []error {
	defer close(generateIngestionLogChan)

	errs := make([]error, 0)

	// Ensure the current default branch is retrieved
	_, _ = p.ops.RefreshDefaultBranch(ctx)

	// Phase 1: Pre-validate entities, build ingestion log protos, and generate entity hashes
	fingerprinter := entityhash.NewEntityFingerprinter()

	type validEntity struct {
		index        int
		ingestionLog *reconcilerpb.IngestionLog
		entityHash   string
		entity       *diodepb.Entity
		objectType   string
	}

	var valid []validEntity

	for i, v := range ingestReq.GetEntities() {
		if v.GetEntity() == nil {
			errs = append(errs, fmt.Errorf("entity at index %d is nil", i))
			continue
		}

		objectType, err := netbox.GetObjectType(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to extract object type for index %d: %v", i, err))
			continue
		}

		hash, err := fingerprinter.GenerateEntityHash(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to generate entity hash for index %d: %v", i, err))
			continue
		}

		ingestionLog := &reconcilerpb.IngestionLog{
			Id:                 uuid.NewString(),
			RequestId:          ingestReq.GetId(),
			ProducerAppName:    ingestReq.GetProducerAppName(),
			ProducerAppVersion: ingestReq.GetProducerAppVersion(),
			SdkName:            ingestReq.GetSdkName(),
			SdkVersion:         ingestReq.GetSdkVersion(),
			DataType:           objectType, // backwards compatibility
			ObjectType:         objectType,
			Entity:             v,
			IngestionTs:        int64(ingestionTs),
			State:              reconcilerpb.State_QUEUED,
			SourceTs:           v.GetTimestamp().AsTime().UnixNano(),
		}

		valid = append(valid, validEntity{
			index:        i,
			ingestionLog: ingestionLog,
			entityHash:   hash,
			entity:       v,
			objectType:   objectType,
		})
	}

	if len(valid) == 0 {
		return errs
	}

	// Phase 2-4: Bulk create via Ops
	logs := make([]*reconcilerpb.IngestionLog, len(valid))
	sourceMetadata := make([][]byte, len(valid))
	entityHashes := make([]string, len(valid))
	for i, v := range valid {
		logs[i] = v.ingestionLog
		sourceMetadata[i] = nil
		entityHashes[i] = v.entityHash
	}

	results, err := p.ops.BulkCreateIngestionLogs(ctx, logs, sourceMetadata, entityHashes)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to bulk create ingestion logs: %v", err))
		p.metrics.RecordIngestionLogCreate(ctx, false)
		return errs
	}

	// Phase 5: Post-processing — metrics, graph upserts, send to channel
	for i, result := range results {
		if result == nil {
			continue
		}

		v := valid[i]
		ingestionLog := result.IngestionLog
		id := result.ID

		if !result.WasDuplicate {
			p.logger.Debug("created new ingestion log", "id", id, "externalID", ingestionLog.GetId())
		} else {
			p.logger.Debug("ingested duplicate ingestion log", "id", id, "externalID", ingestionLog.GetId())
		}

		attrs := []attribute.KeyValue{
			attribute.Bool(telemetry.AttributeDuplicate, result.WasDuplicate),
		}
		metricsCtx := telemetry.ContextWithMetricAttributes(ctx, attrs...)
		p.metrics.RecordIngestionLogCreate(metricsCtx, true)

		// Upsert entity into graph if graph DB is enabled (non-blocking, errors logged but not fatal)
		if p.graphService != nil {
			start := time.Now()
			_, graphErr := p.graphService.UpsertEntity(ctx, v.entity)
			duration := time.Since(start).Seconds()
			if graphErr != nil {
				p.logger.Warn("graph upsert entity failed",
					"error", graphErr,
					"ingestion_log_id", id,
					"entity_type", v.objectType)
				p.metrics.RecordGraphUpsert(ctx, false, v.objectType, duration)
			} else {
				p.metrics.RecordGraphUpsert(ctx, true, v.objectType, duration)
			}
		}

		if result.WasDuplicate && result.IngestionLog.State == reconcilerpb.State_IGNORED {
			p.logger.Debug("skipping ingestion log because it is a duplicate of an ignored ingestion log", "id", id, "externalID", ingestionLog.GetId())
			continue
		}

		// otherwise, even if it was a duplicate, reprocess to see if it has been updated
		generateIngestionLogChan <- IngestionLogToProcess{
			ingestionLogID: id,
			ingestionLog:   ingestionLog,
			branchID:       result.BranchID,
		}
	}

	return errs
}

func decompressBrotli(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("brotli decompress: %w", err)
	}
	return out, nil
}
