package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
)

// Server is a reconciler Server
type Server struct {
	reconcilerpb.UnimplementedReconcilerServiceServer

	config       Config
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
	redisClient  *redis.Client
	apiKeys      map[string]string
}

// NewServer creates a new reconciler server
func NewServer(ctx context.Context, logger *slog.Logger) (*Server, error) {
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
	component := &Server{
		config:       cfg,
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
		redisClient:  redisClient,
		apiKeys:      apiKeys,
	}
	reconcilerpb.RegisterReconcilerServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

// Name returns the name of the server
func (s *Server) Name() string {
	return "reconciler-grpc-server"
}

// Start starts the server
func (s *Server) Start(_ context.Context) error {
	s.logger.Info("starting component", "name", s.Name(), "port", s.config.GRPCPort)
	return s.grpcServer.Serve(s.grpcListener)
}

// Stop stops the server
func (s *Server) Stop() error {
	s.logger.Info("stopping component", "name", s.Name())
	s.grpcServer.GracefulStop()
	return s.redisClient.Close()
}

// RetrieveIngestionDataSources retrieves ingestion data sources
func (s *Server) RetrieveIngestionDataSources(_ context.Context, in *reconcilerpb.RetrieveIngestionDataSourcesRequest) (*reconcilerpb.RetrieveIngestionDataSourcesResponse, error) {
	if err := validateRetrieveIngestionDataSourcesRequest(in); err != nil {
		return nil, err
	}

	dataSources := make([]*reconcilerpb.IngestionDataSource, 0)
	filterByName := in.Name != ""

	if filterByName {
		if _, ok := s.apiKeys[in.Name]; !ok || in.Name != "DIODE" {
			return nil, fmt.Errorf("data source %s not found", in.Name)
		}
		dataSources = append(dataSources, &reconcilerpb.IngestionDataSource{Name: in.Name, ApiKey: s.apiKeys[in.Name]})
		return &reconcilerpb.RetrieveIngestionDataSourcesResponse{IngestionDataSources: dataSources}, nil
	}

	for name, key := range s.apiKeys {
		if name == "DIODE" {
			dataSources = append(dataSources, &reconcilerpb.IngestionDataSource{Name: name, ApiKey: key})
		}
	}
	return &reconcilerpb.RetrieveIngestionDataSourcesResponse{IngestionDataSources: dataSources}, nil
}

func validateRetrieveIngestionDataSourcesRequest(in *reconcilerpb.RetrieveIngestionDataSourcesRequest) error {
	if in.GetSdkName() == "" {
		return fmt.Errorf("sdk name is empty")
	}
	if in.GetSdkVersion() == "" {
		return fmt.Errorf("sdk version is empty")
	}
	return nil
}
