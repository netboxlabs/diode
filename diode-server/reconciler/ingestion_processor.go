package reconciler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/backoff"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/gen/netbox"
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
)

// RedisClient is an interface that represents the methods used from redis.Client
type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
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
}

// IngestionLogToProcess represents an ingestion log to process
type IngestionLogToProcess struct {
	ingestionLogID int32
	ingestionLog   *reconcilerpb.IngestionLog
	changeSetID    int32
	changeSet      *changeset.ChangeSet
}

// IngestionProcessorOps represents the basic operations that the ingestion processor performs
type IngestionProcessorOps interface {
	CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*ops.CreateIngestionLogResult, error)
	GenerateChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, branchID string) (*int32, *changeset.ChangeSet, error)
	ApplyChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, changeSetID int32, changeSet *changeset.ChangeSet) error
}

// NewIngestionProcessor creates a new ingestion processor
func NewIngestionProcessor(_ context.Context, logger *slog.Logger, cfg Config, redisClient, redisStreamClient RedisClient, redisStreamID string, redisConsumerGroup string, ops IngestionProcessorOps, metrics Metrics) (*IngestionProcessor, error) {
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

func (p *IngestionProcessor) consumeIngestionStream(ctx context.Context, redisStreamID string, redisConsumerGroup, redisConsumer string) error {
	err := p.redisStreamClient.XGroupCreateMkStream(ctx, redisStreamID, redisConsumerGroup, "$").Err()
	if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
		return err
	}

	b := backoff.New(10*time.Second, time.Second)
	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("ingestion processor exiting consumer loop on request")
			return nil
		default:
		}
		streams, err := p.redisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    redisConsumerGroup,
			Consumer: redisConsumer,
			Streams:  []string{redisStreamID, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			if strings.Contains(err.Error(), "NOGROUP") {
				err := p.redisStreamClient.XGroupCreateMkStream(ctx, redisStreamID, redisConsumerGroup, "$").Err()
				if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
					p.logger.Debug("Failed to recreate Redis consumer group.")
				}
			}
			select {
			case <-ctx.Done():
				p.logger.Debug("ingestion processor exiting consumer loop on request")
				return nil
			case <-time.After(b.Duration()):
				continue
			}
		}
		b.Reset()

		for _, msg := range streams[0].Messages {
			_, err := p.handleStreamMessage(ctx, msg)
			if err != nil {
				p.logger.Error("failed to handle stream message", "error", err, "message", msg)

				contextMap := map[string]any{
					"redis_stream_msg_id": msg.ID,
					"consumer":            redisConsumer,
					"hostname":            p.hostname,
				}
				sentry.CaptureError(fmt.Errorf("failed to handle stream message: %v", err), nil, "Ingestion stream", contextMap)

				return err
			}
		}
	}
}

func (p *IngestionProcessor) handleStreamMessage(ctx context.Context, msg redis.XMessage) (chan struct{}, error) {
	doneChan := make(chan struct{})
	defer close(doneChan)

	// Create attributes for metrics
	attrs := []attribute.KeyValue{
		attribute.String(telemetry.AttributeHostname, p.hostname),
	}
	ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

	ingestReq := &diodepb.IngestRequest{}
	if err := proto.Unmarshal([]byte(msg.Values["request"].(string)), ingestReq); err != nil {
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
	var applyChangeSetDoneChan chan struct{}

	if p.Config.AutoApplyChangesets {
		applyChangeSetChan = make(chan IngestionLogToProcess, bufCapacity)
		applyChangeSetDoneChan = make(chan struct{})
	}

	p.GenerateChangeSet(ctx, generateIngestionLogChan, applyChangeSetChan, generateIngestionLogDoneChan)

	if p.Config.AutoApplyChangesets {
		p.ApplyChangeSet(ctx, applyChangeSetChan, applyChangeSetDoneChan)
	} else {
		// Only close the channel if it's not nil to avoid panic
		if applyChangeSetDoneChan != nil {
			close(applyChangeSetDoneChan)
		}
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

	p.redisStreamClient.XAck(ctx, p.redisStreamID, p.redisConsumerGroup, msg.ID)

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
		p.redisStreamClient.XDel(ctx, p.redisStreamID, msg.ID)
		p.metrics.RecordHandleMessage(ctx, true)
	}

	return allDone, nil
}

// GenerateChangeSet generates a change set for an ingestion log
func (p *IngestionProcessor) GenerateChangeSet(ctx context.Context, generateChangeSetChan <-chan IngestionLogToProcess, applyChangeSetChan chan<- IngestionLogToProcess, doneChan chan<- struct{}) {
	limiter := rate.NewLimiter(rate.Limit(p.Config.ReconcilerRateLimiterRPS), p.Config.ReconcilerRateLimiterBurst)

	go func() {
		defer func() {
			if applyChangeSetChan != nil {
				close(applyChangeSetChan)
			}
			if doneChan != nil {
				doneChan <- struct{}{}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				p.logger.Debug("context cancelled", "error", ctx.Err())
				return
			case msg, ok := <-generateChangeSetChan:
				if !ok {
					return
				}
				if err := limiter.Wait(ctx); err != nil {
					p.logger.Debug("rate limiter wait", "error", err)
					return
				}

				id, changeSet, err := p.ops.GenerateChangeSet(ctx, msg.ingestionLogID, msg.ingestionLog, "")
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
						}
					}
				}
			}
		}
	}()
}

// ApplyChangeSet applies a change set for an ingestion log
func (p *IngestionProcessor) ApplyChangeSet(ctx context.Context, applyChan <-chan IngestionLogToProcess, doneChan chan<- struct{}) {
	limiter := rate.NewLimiter(rate.Limit(p.Config.ReconcilerRateLimiterRPS), p.Config.ReconcilerRateLimiterBurst)

	go func() {
		defer func() {
			if doneChan != nil {
				doneChan <- struct{}{}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				p.logger.Debug("context cancelled", "error", ctx.Err())
				return
			case msg, ok := <-applyChan:
				if !ok {
					return
				}
				if err := limiter.Wait(ctx); err != nil {
					p.logger.Debug("rate limiter wait", "error", err)
					return
				}

				if err := p.ops.ApplyChangeSet(ctx, msg.ingestionLogID, msg.ingestionLog, msg.changeSetID, msg.changeSet); err != nil {
					p.logger.Error("error applying changeset", "error", err)
					p.metrics.RecordChangeSetApply(ctx, false, 0)
				} else {
					p.metrics.RecordChangeSetApply(ctx, true, int64(len(msg.changeSet.Changes)))
				}
			}
		}
	}()
}

// CreateIngestionLogs creates ingestion logs for an ingest request
func (p *IngestionProcessor) CreateIngestionLogs(ctx context.Context, ingestReq *diodepb.IngestRequest, ingestionTs int, generateIngestionLogChan chan<- IngestionLogToProcess) []error {
	defer close(generateIngestionLogChan)

	errs := make([]error, 0)

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

		result, err := p.ops.CreateIngestionLog(ctx, ingestionLog, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to process ingestion log: %v", err))
			p.metrics.RecordIngestionLogCreate(ctx, false)
			continue
		}
		ingestionLog = result.IngestionLog
		id := result.ID

		if !result.WasDuplicate {
			p.logger.Debug("created new ingestion log", "id", id, "externalID", ingestionLog.GetId())
		} else {
			p.logger.Debug("ingested duplicate ingestion log", "id", id, "externalID", ingestionLog.GetId())
		}

		attrs := []attribute.KeyValue{
			attribute.Bool(telemetry.AttributeDuplicate, result.WasDuplicate),
		}
		ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

		p.metrics.RecordIngestionLogCreate(ctx, true)

		if result.WasDuplicate && result.IngestionLog.State == reconcilerpb.State_IGNORED {
			p.logger.Debug("skipping ingestion log because it is a duplicate of an ignored ingestion log", "id", id, "externalID", ingestionLog.GetId())
			continue
		}

		// otherwise, even if it was a duplicate, reprocess to see if it has been updated
		generateIngestionLogChan <- IngestionLogToProcess{
			ingestionLogID: id,
			ingestionLog:   ingestionLog,
		}
	}

	return errs
}
