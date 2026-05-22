package ingester

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/grpckeepalive"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/sentry"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

// defaultRedisMemoryCheckInterval is used when Config.RedisMemoryCheckInterval
// is zero or negative. The watermark check is only as fresh as this interval;
// bursts within it are admitted on the previously-cached value.
const defaultRedisMemoryCheckInterval = 500 * time.Millisecond

// defaultRedisInfoTimeout is used when Config.RedisMemoryCheckTimeout is
// zero or negative. The mutex is held across the INFO I/O, so a hung Redis
// would otherwise stall every concurrent Ingest for the duration of the
// redis-client ReadTimeout. INFO is a tiny command — if it hasn't returned
// within the configured timeout, prefer to fail-open on this tick and try
// again next interval rather than convoy the request path.
const defaultRedisInfoTimeout = 250 * time.Millisecond

// Component asynchronously ingests data from the distributor
type Component struct {
	diodepb.UnimplementedIngesterServiceServer

	ctx               context.Context
	config            Config
	logger            *slog.Logger
	hostname          string
	grpcListener      net.Listener
	grpcServer        *grpc.Server
	redisStreamClient *redis.Client
	metrics           Metrics
	streamRouter      StreamRouter

	memCheckMu   sync.Mutex
	memUsedBytes int64
	memMaxBytes  int64
	memCheckedAt time.Time
}

// StreamRouter is an interface for determining the stream ID to add ingested data into
type StreamRouter interface {
	// GetIngestStreamID returns the redis stream ID to add ingested data into
	GetIngestStreamID(ctx context.Context, in *diodepb.IngestRequest) (string, error)
}

// DefaultStreamRouter is the default implementation of the StreamRouter interface
type DefaultStreamRouter struct{}

// GetIngestStreamID returns the default redis stream ID
func (s *DefaultStreamRouter) GetIngestStreamID(_ context.Context, _ *diodepb.IngestRequest) (string, error) {
	return reconciler.DefaultRedisStreamID, nil
}

// New creates a new ingester component
func New(ctx context.Context, logger *slog.Logger, cfg Config, redisStreamClient *redis.Client, metrics Metrics, streamRouter StreamRouter, serverInterceptors ...grpc.UnaryServerInterceptor) (*Component, error) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	grpcServer := grpc.NewServer(append(grpckeepalive.ServerOptions(),
		grpc.ChainUnaryInterceptor(serverInterceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)...)

	component := &Component{
		ctx:               ctx,
		config:            cfg,
		logger:            logger,
		hostname:          hostname,
		grpcListener:      grpcListener,
		grpcServer:        grpcServer,
		redisStreamClient: redisStreamClient,
		metrics:           metrics,
		streamRouter:      streamRouter,
	}

	diodepb.RegisterIngesterServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

// Name returns the name of the component
func (c *Component) Name() string {
	return "ingester"
}

// Start starts the component
func (c *Component) Start(_ context.Context) error {
	c.logger.Info("starting component", "name", c.Name(), "port", c.config.GRPCPort)
	return c.grpcServer.Serve(c.grpcListener)
}

// Stop stops the component
func (c *Component) Stop() error {
	c.logger.Info("stopping component", "name", c.Name())
	c.grpcServer.GracefulStop()
	return c.redisStreamClient.Close()
}

// Ingest handles the ingest request
func (c *Component) Ingest(ctx context.Context, in *diodepb.IngestRequest) (*diodepb.IngestResponse, error) {
	// Create attributes for metrics
	attrs := []attribute.KeyValue{
		attribute.String(telemetry.AttributeSDKName, in.SdkName),
		attribute.String(telemetry.AttributeSDKVersion, in.SdkVersion),
		attribute.String(telemetry.AttributeHostname, c.hostname),
		attribute.String(telemetry.AttributeProducerAppName, in.ProducerAppName),
		attribute.String(telemetry.AttributeProducerAppVersion, in.ProducerAppVersion),
	}
	ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

	if err := validateRequest(in); err != nil {
		c.metrics.RecordIngestRequest(ctx, false)

		tags := map[string]string{
			"hostname":    c.hostname,
			"sdk_name":    in.SdkName,
			"sdk_version": in.SdkVersion,
		}
		contextMap := map[string]any{
			"request_id":           in.Id,
			"producer_app_name":    in.ProducerAppName,
			"producer_app_version": in.ProducerAppVersion,
			"sdk_name":             in.SdkName,
			"sdk_version":          in.SdkVersion,
		}
		sentry.CaptureError(err, tags, "Ingest Request", contextMap)
		return nil, err
	}

	if err := c.checkRedisMemoryWatermark(ctx); err != nil {
		c.metrics.RecordIngestRequest(ctx, false)
		return nil, err
	}

	errs := make([]string, 0)

	encodedRequest, err := proto.Marshal(in)
	if err != nil {
		c.metrics.RecordIngestRequest(ctx, false)
		c.logger.Error("failed to marshal request", "error", err, "request", in)
	}

	for i, v := range in.GetEntities() {
		if v.GetEntity() == nil {
			errs = append(errs, fmt.Sprintf("entity at index %d is nil", i))
			continue
		}
	}

	msg := map[string]any{
		"ingestion_ts": time.Now().UnixNano(),
	}

	if c.config.CompressStreamMessages {
		compressed, cErr := compressBrotli(encodedRequest)
		if cErr != nil {
			c.metrics.RecordIngestRequest(ctx, false)
			c.logger.Error("failed to compress request", "error", cErr)
			return nil, status.Error(codes.Internal, "")
		}
		msg["request"] = compressed
		msg["encoding"] = "br"
	} else {
		msg["request"] = encodedRequest
	}

	streamID, err := c.streamRouter.GetIngestStreamID(ctx, in)
	if err != nil {
		c.metrics.RecordIngestRequest(ctx, false)
		c.logger.Error("failed to get stream ID", "error", err, "request", in)
		return nil, err
	}

	attrs = []attribute.KeyValue{
		attribute.String(telemetry.AttributeStream, streamID),
	}
	ctx = telemetry.ContextWithMetricAttributes(ctx, attrs...)

	err = c.redisStreamClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamID,
		Values: msg,
	}).Err()
	if err != nil {
		c.metrics.RecordIngestRequest(ctx, false)
		// Defense in depth: if a burst crossed the threshold between
		// watermark polls, or fail-open admitted under INFO error, Redis
		// itself may reject with OOM (only when configured with
		// maxmemory + maxmemory-policy noeviction). Surface that as
		// ResourceExhausted so SDK callers see a consistent backoff
		// signal rather than Internal.
		if isRedisOOMErr(err) {
			c.metrics.RecordRedisRejection(ctx, "redis_oom")
			c.logger.Warn("rejecting ingest: redis returned OOM on XAdd", "error", err, "streamID", streamID)
			return nil, status.Error(codes.ResourceExhausted, "redis at maxmemory; retry with backoff")
		}
		c.logger.Error("failed to add element to the stream", "error", err, "streamID", streamID, "value", msg)
		return nil, status.Error(codes.Internal, "")
	}

	entityCount := int64(len(in.GetEntities()))
	c.metrics.RecordIngestRequest(ctx, true)
	c.metrics.RecordIngestEntities(ctx, entityCount)

	return &diodepb.IngestResponse{Errors: errs}, nil
}

func validateRequest(in *diodepb.IngestRequest) error {
	if in.GetId() == "" {
		return fmt.Errorf("id is empty")
	}

	if in.GetProducerAppName() == "" {
		return fmt.Errorf("producer app name is empty")
	}

	if in.GetProducerAppVersion() == "" {
		return fmt.Errorf("producer app version is empty")
	}

	if in.GetSdkName() == "" {
		return fmt.Errorf("sdk name is empty")
	}

	if in.GetSdkVersion() == "" {
		return fmt.Errorf("sdk version is empty")
	}

	if len(in.GetEntities()) < 1 {
		return fmt.Errorf("entities is empty")
	}

	return nil
}

// checkRedisMemoryWatermark returns codes.ResourceExhausted when Redis
// used_memory is at or above the configured percentage of maxmemory.
// Disabled when the threshold is <= 0.
//
// The check is global on used_memory (not per-tenant/per-stream). The
// goal is to leave headroom on a Redis that may be shared with other
// workloads, so Diode's total share against the Redis pool is the
// relevant budget.
//
// INFO is called at most once per RedisMemoryCheckInterval and bounded
// by RedisMemoryCheckTimeout. On INFO failure or timeout the previous
// reading is retained, so a struggling Redis does not suddenly admit a
// flood. With no successful reading yet (cold start) or when Redis
// maxmemory is 0 (unlimited), the request is admitted.
func (c *Component) checkRedisMemoryWatermark(ctx context.Context) error {
	pct := c.config.RedisMemoryHighWatermarkPct
	if pct <= 0 {
		return nil
	}

	c.memCheckMu.Lock()
	defer c.memCheckMu.Unlock()

	interval := c.config.RedisMemoryCheckInterval
	if interval <= 0 {
		interval = defaultRedisMemoryCheckInterval
	}

	if time.Since(c.memCheckedAt) >= interval {
		timeout := c.config.RedisMemoryCheckTimeout
		if timeout <= 0 {
			timeout = defaultRedisInfoTimeout
		}
		infoCtx, cancel := context.WithTimeout(ctx, timeout)
		usedBytes, maxBytes, err := readRedisMemory(infoCtx, c.redisStreamClient)
		cancel()
		if err != nil {
			// Retain the previous reading rather than zeroing it: if Redis
			// is struggling, INFO is exactly when we want to keep enforcing
			// the last-known state, not suddenly admit a flood. memCheckedAt
			// is still bumped so we don't hammer a struggling Redis with
			// retries every Ingest call.
			c.logger.Warn("failed to read Redis memory; retaining previous reading",
				"error", err,
				"used_bytes", c.memUsedBytes,
				"max_bytes", c.memMaxBytes)
			c.memCheckedAt = time.Now()
		} else {
			c.memUsedBytes = usedBytes
			c.memMaxBytes = maxBytes
			c.memCheckedAt = time.Now()
			if maxBytes > 0 {
				c.metrics.SetRedisMemoryRatioBPS(ctx, usedBytes*10000/maxBytes)
			} else {
				// Logged once per refresh tick (not per Ingest request).
				c.logger.Warn("Redis maxmemory is 0 (unlimited); high-watermark check has no effect")
			}
		}
	}

	if c.memMaxBytes <= 0 {
		return nil
	}

	// Integer compare: used/max >= pct/100  <=>  used*100 >= max*pct
	if c.memUsedBytes*100 >= c.memMaxBytes*int64(pct) {
		c.metrics.RecordRedisRejection(ctx, "watermark")
		c.logger.Warn("rejecting ingest: Redis memory above high-watermark",
			"used_bytes", c.memUsedBytes, "max_bytes", c.memMaxBytes, "threshold_pct", pct)
		return status.Errorf(codes.ResourceExhausted,
			"redis memory above high-watermark (%d/%d bytes >= %d%%)",
			c.memUsedBytes, c.memMaxBytes, pct)
	}
	return nil
}

// isRedisOOMErr reports whether err is the Redis "OOM command not allowed
// when used memory > 'maxmemory'" error, which is only returned when Redis
// is configured with maxmemory and a noeviction policy. Redis sends this
// as a RESP error reply with the literal "OOM " prefix; go-redis surfaces
// the message verbatim.
func isRedisOOMErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "OOM ")
}

// readRedisMemory issues INFO memory and parses the used_memory and maxmemory lines.
func readRedisMemory(ctx context.Context, client *redis.Client) (usedBytes, maxBytes int64, err error) {
	out, err := client.Info(ctx, "memory").Result()
	if err != nil {
		return 0, 0, err
	}
	return parseMemory(out)
}

// parseMemory extracts used_memory and maxmemory from Redis INFO output.
// Returns an error if used_memory is missing; maxmemory defaults to 0 if
// absent (Redis omits it when no limit is configured).
func parseMemory(info string) (usedBytes, maxBytes int64, err error) {
	usedFound := false
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "used_memory:"); ok {
			v, parseErr := strconv.ParseInt(strings.TrimSpace(rest), 10, 64)
			if parseErr != nil {
				return 0, 0, parseErr
			}
			usedBytes = v
			usedFound = true
			continue
		}
		if rest, ok := strings.CutPrefix(line, "maxmemory:"); ok {
			v, parseErr := strconv.ParseInt(strings.TrimSpace(rest), 10, 64)
			if parseErr != nil {
				return 0, 0, parseErr
			}
			maxBytes = v
		}
	}
	if !usedFound {
		return 0, 0, fmt.Errorf("used_memory not found in INFO output")
	}
	return usedBytes, maxBytes, nil
}

func compressBrotli(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := brotli.NewWriterLevel(&buf, 1)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("brotli write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("brotli close: %w", err)
	}
	return buf.Bytes(), nil
}
