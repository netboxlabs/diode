package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/kelseyhightower/envconfig"
	pb "github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Component reconciles ingested data
type Component struct {
	pb.UnimplementedReconcilerServiceServer

	config       Config
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
	redisClient  *redis.Client
	apiKeys      map[string]string
}

// New creates a new reconciler component
func New(ctx context.Context, logger *slog.Logger) (*Component, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisClient.String(), err)
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	apiKeys, err := loadAPIKeys(ctx, cfg, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to configure data sources: %v", err)
	}

	grpcServer := grpc.NewServer()
	component := &Component{
		config:       cfg,
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
		redisClient:  redisClient,
		apiKeys:      apiKeys,
	}
	pb.RegisterReconcilerServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

// Name returns the name of the component
func (c *Component) Name() string {
	return "reconciler"
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
	return c.redisClient.Close()
}

// RetrieveIngestionDataSources retrieves ingestion data sources
func (c *Component) RetrieveIngestionDataSources(_ context.Context, in *pb.RetrieveIngestionDataSourcesRequest) (*pb.RetrieveIngestionDataSourcesResponse, error) {
	if err := validateRetrieveIngestionDataSourcesRequest(in); err != nil {
		return nil, err
	}

	dataSources := make([]*pb.IngestionDataSource, 0)
	filterByName := in.Name != ""

	if filterByName {
		if _, ok := c.apiKeys[in.Name]; !ok || !strings.HasPrefix(in.Name, "INGESTION") {
			return nil, fmt.Errorf("data source %s not found", in.Name)
		}
		dataSources = append(dataSources, &pb.IngestionDataSource{Name: in.Name, ApiKey: c.apiKeys[in.Name]})
		return &pb.RetrieveIngestionDataSourcesResponse{IngestionDataSources: dataSources}, nil
	}

	for name, key := range c.apiKeys {
		if strings.HasPrefix(name, "INGESTION") {
			dataSources = append(dataSources, &pb.IngestionDataSource{Name: name, ApiKey: key})
		}
	}
	return &pb.RetrieveIngestionDataSourcesResponse{IngestionDataSources: dataSources}, nil
}

func validateRetrieveIngestionDataSourcesRequest(in *pb.RetrieveIngestionDataSourcesRequest) error {
	if in.GetSdkName() == "" {
		return fmt.Errorf("sdk name is empty")
	}
	if in.GetSdkVersion() == "" {
		return fmt.Errorf("sdk version is empty")
	}
	return nil
}
