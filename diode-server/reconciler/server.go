package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// Server is a reconciler Server
type Server struct {
	reconcilerpb.UnimplementedReconcilerServiceServer

	config       Config
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
	redisClient  RedisClient
	repository   Repository
}

// NewServer creates a new reconciler server
func NewServer(ctx context.Context, logger *slog.Logger, repository Repository, serverInterceptors ...grpc.UnaryServerInterceptor) (*Server, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	redisTLSConfig, err := cfg.RedisTLS.ToTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config for Redis: %v", err)
	}

	redisOptions := redis.Options{
		Addr:      fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:  cfg.RedisPassword,
		DB:        cfg.RedisDB,
		TLSConfig: redisTLSConfig,
	}
	if cfg.RedisUsername != "" {
		redisOptions.Username = cfg.RedisUsername
	}
	redisClient := redis.NewClient(&redisOptions)

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisClient.String(), err)
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(serverInterceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	component := &Server{
		config:       cfg,
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
		redisClient:  redisClient,
		repository:   repository,
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

// GRPCServer returns the grpc.Server managed by this Server
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

// RetrieveIngestionLogs retrieves logs
func (s *Server) RetrieveIngestionLogs(ctx context.Context, in *reconcilerpb.RetrieveIngestionLogsRequest) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	return retrieveIngestionLogs(ctx, s.logger, s.repository, in)
}

// RetrieveDeviations retrieves deviations
func (s *Server) RetrieveDeviations(ctx context.Context, req *reconcilerpb.RetrieveDeviationsRequest) (*reconcilerpb.RetrieveDeviationsResponse, error) {
	return retrieveDeviations(ctx, s.logger, s.repository, req)
}

// RetrieveDeviationByID retrieves a deviation by ID
func (s *Server) RetrieveDeviationByID(ctx context.Context, req *reconcilerpb.RetrieveDeviationByIDRequest) (*reconcilerpb.RetrieveDeviationByIDResponse, error) {
	return retrieveDeviationByID(ctx, s.logger, s.repository, req)
}
