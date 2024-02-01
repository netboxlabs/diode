package distributor

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/kelseyhightower/envconfig"
	pb "github.com/netboxlabs/diode-internal/diode-sdk-go/diode/v1/diodepb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Component is a gRPC server that handles data ingestion requests
type Component struct {
	pb.UnimplementedDistributorServiceServer

	ctx          context.Context
	config       Config
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
	redisClient  *redis.Client
}

// New creates a new distributor component
func New(ctx context.Context, logger *slog.Logger) (*Component, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisClient.String(), err)
	}

	grpcServer := grpc.NewServer()
	component := &Component{
		ctx:          ctx,
		config:       cfg,
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
		redisClient:  redisClient,
	}
	pb.RegisterDistributorServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

// Name returns the name of the component
func (c *Component) Name() string {
	return "distributor"
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
	if err := c.redisClient.Close(); err != nil {
		return err
	}
	return nil
}

// Push handles a push request
func (c *Component) Push(_ context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	c.logger.Info("diode.v1.DistributorService/Push called", "stream", in.Stream)
	return &pb.PushResponse{}, nil
}
