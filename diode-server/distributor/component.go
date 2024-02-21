package distributor

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/kelseyhightower/envconfig"
	pb "github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
)

const (
	streamID = "diode.v1.ingest-stream"
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
		DB:       cfg.RedisStreamDB,
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
	return c.redisClient.Close()
}

// Push handles a push request
func (c *Component) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	if err := validatePushRequest(in); err != nil {
		return nil, err
	}

	errs := make([]string, 0)

	encodedRequest, err := proto.Marshal(in)
	if err != nil {
		c.logger.Error("failed to marshal request", "error", err, "request", in)
	}

	for i, v := range in.GetData() {
		if v.GetData() == nil {
			errs = append(errs, fmt.Sprintf("data for index %d is nil", i))
			continue
		}
	}

	msg := map[string]interface{}{
		"request":      encodedRequest,
		"ingestion_ts": time.Now().UnixNano(),
	}

	if err := c.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamID,
		Values: msg,
	}).Err(); err != nil {
		c.logger.Error("failed to add element to the stream", "error", err, "streamID", streamID, "value", msg)
	}

	return &pb.PushResponse{Errors: errs}, nil
}

func validatePushRequest(in *pb.PushRequest) error {
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

	if len(in.GetData()) < 1 {
		return fmt.Errorf("data is empty")
	}

	return nil
}
