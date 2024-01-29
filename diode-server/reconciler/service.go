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

type Server struct {
	logger       *slog.Logger
	grpcListener net.Listener
	grpcServer   *grpc.Server
}

func New(logger *slog.Logger) (*Server, error) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", DefaultGRPCPort))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to listen on port %s: %v", DefaultGRPCPort, err))
	}

	grpcServer := grpc.NewServer()
	server := &Server{
		logger:       logger,
		grpcListener: grpcListener,
		grpcServer:   grpcServer,
	}
	reflection.Register(grpcServer)

	return server, nil
}

func (s *Server) Name() string {
	return "reconciler"
}

func (s *Server) Start(_ context.Context) error {
	s.logger.Info("starting service", "name", s.Name())

	return s.grpcServer.Serve(s.grpcListener)
}

func (s *Server) Stop() error {
	s.logger.Info("stopping service", "name", s.Name())
	s.grpcServer.GracefulStop()
	return nil
}
