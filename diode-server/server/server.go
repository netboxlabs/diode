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

	mu         sync.Mutex
	components map[string]Component

	componentGroup run.Group
}

// Component is used for registering components managed by the diode Server
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
}

// New returns a new Server
func New(ctx context.Context, name string, logger *slog.Logger) *Server {
	return &Server{
		cxt:            ctx,
		name:           name,
		logger:         logger,
		components:     make(map[string]Component),
		componentGroup: run.Group{},
	}
}

// Name returns the name of the Server
func (s *Server) Name() string {
	return s.name
}

// RegisterComponent registers a Component with the Server
func (s *Server) RegisterComponent(c Component) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.components[c.Name()]; ok {
		return errors.New(fmt.Sprintf("Server.RegisterComponent found duplicate component registration for %s", c.Name()))
	}

	s.components[c.Name()] = c

	ctx, cancel := context.WithCancel(s.cxt)

	s.componentGroup.Add(
		func() error {
			return c.Start(ctx)
		},
		func(err error) {
			if err := c.Stop(); err != nil {
				s.logger.Error("failed to stop component", "componentName", c.Name(), "error", err)
			}
			cancel()
		},
	)
	return nil
}

// Run starts the diode Server
func (s *Server) Run() error {
	s.logger.Info("starting server", "serverName", s.name)

	s.componentGroup.Add(run.SignalHandler(s.cxt, os.Interrupt, os.Kill))

	return s.componentGroup.Run()
}
