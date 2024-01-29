package distributor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	pb "github.com/netboxlabs/diode-internal/diode-sdk-go/diode/v1/diodepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	DefaultGRPCPort = "8081"
)

type Component struct {
	pb.UnimplementedDistributorServiceServer

	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
}

func New(logger *slog.Logger) (*Component, error) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", DefaultGRPCPort))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to listen on port %s: %v", DefaultGRPCPort, err))
	}

	grpcServer := grpc.NewServer()
	component := &Component{
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
	}
	pb.RegisterDistributorServiceServer(grpcServer, component)
	reflection.Register(grpcServer)

	return component, nil
}

func (c *Component) Name() string {
	return "distributor"
}

func (c *Component) Start(_ context.Context) error {
	c.logger.Info("starting component", "name", c.Name())

	return c.grpcServer.Serve(c.grpcListener)
}

func (c *Component) Stop() error {
	c.logger.Info("stopping component", "name", c.Name())
	c.grpcServer.GracefulStop()
	return nil
}

func (c *Component) Push(_ context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	c.logger.Info("diode.v1.DistributorService/Push called", "stream", in.Stream)
	return &pb.PushResponse{}, nil
}
