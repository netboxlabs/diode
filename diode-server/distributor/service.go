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

type Server struct {
	pb.UnimplementedDistributorServiceServer

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
	pb.RegisterDistributorServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	return server, nil
}

func (s *Server) Name() string {
	return "distributor"
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting service", "name", s.Name())

	return s.grpcServer.Serve(s.grpcListener)
}

func (s *Server) Stop() error {
	s.logger.Info("stopping service", "name", s.Name())
	s.grpcServer.GracefulStop()
	return nil
}

func (s *Server) Push(_ context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	s.logger.Info("diode.v1.DistributorService/Push called", "stream", in.Stream)
	return &pb.PushResponse{}, nil
}
