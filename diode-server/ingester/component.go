package ingester

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
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
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/sentry"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

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

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(serverInterceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

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
