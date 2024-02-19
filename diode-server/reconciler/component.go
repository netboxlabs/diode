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
)

// Component reconciles ingested data
type Component struct {
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
	return nil
}
