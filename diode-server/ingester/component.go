package ingester

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/sentry"
)

const (
	streamID = "diode.v1.ingest-stream"
)

var (
	errMetadataNotFound = errors.New("no request metadata found")

	// ErrUnauthorized is an error for unauthorized requests
	ErrUnauthorized = errors.New("missing or invalid authorization header")
)

// Component asynchronously ingests data from the distributor
type Component struct {
	diodepb.UnimplementedIngesterServiceServer

	ctx                  context.Context
	config               Config
	logger               *slog.Logger
	hostname             string
	grpcListener         net.Listener
	grpcServer           *grpc.Server
	redisStreamClient    *redis.Client
	reconcilerClient     reconciler.Client
	ingestionDataSources []*reconcilerpb.IngestionDataSource
}

// New creates a new ingester component
func New(ctx context.Context, logger *slog.Logger) (*Component, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

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

	reconcilerClient, err := reconciler.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create reconciler client: %v", err)
	}

	dataSources, err := reconcilerClient.RetrieveIngestionDataSources(ctx, &reconcilerpb.RetrieveIngestionDataSourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ingestion data sources: %v", err)
	}

	ingestionDataSources := dataSources.GetIngestionDataSources()
	//auth := grpc.UnaryServerInterceptor(authUnaryInterceptor)
	auth := newAuthUnaryInterceptor(ingestionDataSources)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(auth))

	component := &Component{
		ctx:                  ctx,
		config:               cfg,
		logger:               logger,
		hostname:             hostname,
		grpcListener:         grpcListener,
		grpcServer:           grpcServer,
		redisStreamClient:    redisStreamClient,
		reconcilerClient:     reconcilerClient,
		ingestionDataSources: ingestionDataSources,
	}

	diodepb.RegisterIngesterServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

func newAuthUnaryInterceptor(dataSources []*reconcilerpb.IngestionDataSource) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMetadataNotFound
		}
		if !authorized(dataSources, md["diode-api-key"]) {
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
	if err := validateRequest(in); err != nil {
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
		c.logger.Error("failed to add element to the stream", "error", err, "streamID", streamID, "value", msg)
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

func authorized(dataSources []*reconcilerpb.IngestionDataSource, authorization []string) bool {
	if len(dataSources) < 1 || len(authorization) != 1 {
		return false
	}

	for _, v := range dataSources {
		if v.GetApiKey() == authorization[0] {
			return true
		}
	}
	return false
}
