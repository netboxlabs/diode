package distributor

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/kelseyhightower/envconfig"
	pb "github.com/netboxlabs/diode-internal/diode-sdk-go/diode/v1/diodepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Component is a gRPC server that handles data ingestion requests
type Component struct {
	pb.UnimplementedDistributorServiceServer

	config       Config
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
}

// New creates a new distributor component
func New(logger *slog.Logger) (*Component, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	component := &Component{
		config:       cfg,
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
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
	return nil
}

func (c *Component) Push(_ context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	c.logger.Info("diode.v1.DistributorService/Push called", "stream", in.Stream)
	return &pb.PushResponse{}, nil
}
