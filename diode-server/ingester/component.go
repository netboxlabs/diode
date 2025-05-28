package ingester

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	metrics           *Metrics
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
func (s *DefaultStreamRouter) GetIngestStreamID(ctx context.Context, in *diodepb.IngestRequest) (string, error) {
	return reconciler.DefaultRedisStreamID, nil
}

// New creates a new ingester component
func New(ctx context.Context, logger *slog.Logger, cfg Config, redisStreamClient *redis.Client, meter metric.Meter, streamRouter StreamRouter, serverInterceptors ...grpc.UnaryServerInterceptor) (*Component, error) {
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

	metrics, err := NewMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create ingester metrics: %v", err)
	}

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
		"request":      encodedRequest,
		"ingestion_ts": time.Now().UnixNano(),
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

	if err := c.redisStreamClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamID,
		Values: msg,
	}).Err(); err != nil {
		c.metrics.RecordIngestRequest(ctx, false)
		c.logger.Error("failed to add element to the stream", "error", err, "streamID", streamID, "value", msg)
	} else {
		c.metrics.RecordIngestRequest(ctx, true)
		c.metrics.RecordIngestEntities(ctx, int64(len(in.GetEntities())))
	}

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
