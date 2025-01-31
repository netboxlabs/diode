package ingester

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/sentry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	streamID = "diode.v1.ingest-stream"

	// Metric names
	metricIngestTotal   = "diode.ingester.ingest.total"
	metricIngestSuccess = "diode.ingester.ingest.success"
	metricIngestFailure = "diode.ingester.ingest.failure"
)

var (
	errMetadataNotFound = errors.New("no request metadata found")

	// ErrUnauthorized is an error for unauthorized requests
	ErrUnauthorized = errors.New("missing or invalid authorization header")
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

	// Metrics
	ingestTotalCounter   metric.Int64Counter
	ingestSuccessCounter metric.Int64Counter
	ingestFailureCounter metric.Int64Counter
}

// New creates a new ingester component
func New(ctx context.Context, logger *slog.Logger, cfg Config) (*Component, error) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisStreamDB,
	})

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisStreamClient.String(), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	apiKeys := loadAPIKeys(cfg)
	auth := newAuthUnaryInterceptor(apiKeys)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(auth))

	meter := otel.GetMeterProvider().Meter("diode/ingester")

	ingestTotal, err := meter.Int64Counter(metricIngestTotal,
		metric.WithDescription("Total number of ingest requests received"))
	if err != nil {
		return nil, fmt.Errorf("failed to create total counter: %v", err)
	}

	ingestSuccess, err := meter.Int64Counter(metricIngestSuccess,
		metric.WithDescription("Number of successful ingest requests"))
	if err != nil {
		return nil, fmt.Errorf("failed to create success counter: %v", err)
	}

	ingestFailure, err := meter.Int64Counter(metricIngestFailure,
		metric.WithDescription("Number of failed ingest requests"))
	if err != nil {
		return nil, fmt.Errorf("failed to create failure counter: %v", err)
	}

	component := &Component{
		ctx:                  ctx,
		config:               cfg,
		logger:               logger,
		hostname:             hostname,
		grpcListener:         grpcListener,
		grpcServer:           grpcServer,
		redisStreamClient:    redisStreamClient,
		ingestTotalCounter:   ingestTotal,
		ingestSuccessCounter: ingestSuccess,
		ingestFailureCounter: ingestFailure,
	}

	diodepb.RegisterIngesterServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

func newAuthUnaryInterceptor(apiKeys []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMetadataNotFound
		}
		if !isAuthenticated(apiKeys, md["diode-api-key"]) {
			return nil, ErrUnauthorized
		}
		return handler(ctx, req)
	}
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
		attribute.String("sdk_name", in.SdkName),
		attribute.String("sdk_version", in.SdkVersion),
		attribute.String("hostname", c.hostname),
		attribute.String("producer_app_name", in.ProducerAppName),
		attribute.String("producer_app_version", in.ProducerAppVersion),
		attribute.String("stream", in.Stream),
	}

	// Record total ingest attempt
	c.ingestTotalCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	if err := validateRequest(in); err != nil {
		// Record failure
		c.ingestFailureCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

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
			"stream":               in.Stream,
		}
		sentry.CaptureError(err, tags, "Ingest Request", contextMap)
		return nil, err
	}

	errs := make([]string, 0)

	encodedRequest, err := proto.Marshal(in)
	if err != nil {
		c.logger.Error("failed to marshal request", "error", err, "request", in)
	}

	for i, v := range in.GetEntities() {
		if v.GetEntity() == nil {
			errs = append(errs, fmt.Sprintf("entity at index %d is nil", i))
			continue
		}
	}

	msg := map[string]interface{}{
		"request":      encodedRequest,
		"ingestion_ts": time.Now().UnixNano(),
	}

	if err := c.redisStreamClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamID,
		Values: msg,
	}).Err(); err != nil {
		c.ingestFailureCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		c.logger.Error("failed to add element to the stream", "error", err, "streamID", streamID, "value", msg)
	} else {
		c.ingestSuccessCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
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

func loadAPIKeys(cfg Config) []string {
	return []string{
		cfg.DiodeAPIKey,
	}
}

func isAuthenticated(apiKeys []string, authorization []string) bool {
	if len(apiKeys) < 1 || len(authorization) != 1 {
		return false
	}

	return slices.Contains(apiKeys, authorization[0])
}
