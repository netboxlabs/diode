package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/oklog/run"
)

// A Server is a diode Server
type Server struct {
	cxt    context.Context
	name   string
	logger *slog.Logger

	mu       sync.Mutex
	services map[string]Service

	serviceGroup run.Group
}

// Service is used for registering services managed by the diode Server
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
}

// New returns a new Server
func New(ctx context.Context, name string, logger *slog.Logger) *Server {
	return &Server{
		cxt:          ctx,
		name:         name,
		logger:       logger,
		services:     make(map[string]Service),
		serviceGroup: run.Group{},
	}
}

// Name returns the name of the Server
func (s *Server) Name() string {
	return s.name
}

// RegisterService registers a Service with the Server
func (s *Server) RegisterService(service Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.services[service.Name()]; ok {
		return errors.New(fmt.Sprintf("Server.RegisterService found duplicate service registration for %s", service.Name()))
	}

	s.services[service.Name()] = service

	ctx, cancel := context.WithCancel(s.cxt)

	s.serviceGroup.Add(
		func() error {
			return service.Start(ctx)
		},
		func(err error) {
			if err := service.Stop(); err != nil {
				s.logger.Error("failed to stop service", "serviceName", service.Name(), "error", err)
			}
			cancel()
		},
	)
	return nil
}

// Run starts the diode Server
func (s *Server) Run() error {
	s.logger.Info("starting server", "serverName", s.name)

	s.serviceGroup.Add(run.SignalHandler(s.cxt, os.Interrupt, os.Kill))

	return s.serviceGroup.Run()
}
