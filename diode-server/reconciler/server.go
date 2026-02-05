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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

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

// ListEntities lists observed entities with filtering
func (s *Server) ListEntities(ctx context.Context, req *reconcilerpb.ListEntitiesRequest) (*reconcilerpb.ListEntitiesResponse, error) {
	// Set default page size if not specified
	pageSize := int32(100)
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = *req.PageSize
		// Cap at 1000 per proto definition
		if pageSize > 1000 {
			pageSize = 1000
		}
	}

	// Parse page token for offset (simple implementation using offset as token)
	offset := int32(0)
	if req.PageToken != "" {
		// Simple implementation: page token is just the offset as string
		// In production, you might want to use a more sophisticated token
		if _, err := fmt.Sscanf(req.PageToken, "%d", &offset); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %v", err)
		}
	}

	// Query entities from the repository
	entities, err := s.repository.ListGraphEntities(ctx, req, pageSize, offset)
	if err != nil {
		s.logger.Error("failed to list graph entities", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list entities: %v", err)
	}

	// Generate next page token if we got a full page
	nextPageToken := ""
	if len(entities) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", offset+pageSize)
	}

	s.logger.Debug("listed entities",
		"count", len(entities),
		"page_size", pageSize,
		"offset", offset,
		"has_next_page", nextPageToken != "")

	return &reconcilerpb.ListEntitiesResponse{
		Entities:      entities,
		NextPageToken: nextPageToken,
	}, nil
}

// CreateEntity creates an entity synchronously in the graph database (idempotent - returns existing ID if entity already exists)
func (s *Server) CreateEntity(ctx context.Context, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	if req.Entity == nil {
		return nil, status.Errorf(codes.InvalidArgument, "entity is required")
	}

	// Create the entity in the graph using the repository
	externalID, nodeType, err := s.repository.CreateEntityInGraph(ctx, req.Entity)
	if err != nil {
		s.logger.Error("failed to create entity in graph", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create entity: %v", err)
	}

	s.logger.Info("created entity in graph",
		"external_id", externalID,
		"node_type", nodeType)

	return &reconcilerpb.CreateEntityResponse{
		Id:         externalID,
		ObjectType: nodeType,
	}, nil
}
