package reconciler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	DefaultGRPCPort = "8081"
)

type Component struct {
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
	reflection.Register(grpcServer)

	return component, nil
}

func (c *Component) Name() string {
	return "reconciler"
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
