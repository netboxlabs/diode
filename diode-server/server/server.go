package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/run"
)

// A Server is a diode Server
type Server struct {
	ctx    context.Context
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
func New(ctx context.Context, name string) *Server {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	return &Server{
		ctx:            ctx,
		name:           name,
		logger:         newLogger(cfg),
		components:     make(map[string]Component),
		componentGroup: run.Group{},
	}
}

// Name returns the name of the Server
func (s *Server) Name() string {
	return s.name
}

// Logger returns the logger of the Server
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// RegisterComponent registers a Component with the Server
func (s *Server) RegisterComponent(c Component) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.components[c.Name()]; ok {
		return fmt.Errorf("Server.RegisterComponent found duplicate component registration for %s", c.Name())
	}

	s.components[c.Name()] = c

	ctx, cancel := context.WithCancel(s.ctx)

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

	s.componentGroup.Add(run.SignalHandler(s.ctx, os.Interrupt, os.Kill))

	return s.componentGroup.Run()
}

func newLogger(cfg Config) *slog.Logger {
	var l slog.Level
	switch strings.ToUpper(cfg.LoggingLevel) {
	case "DEBUG":
		l = slog.LevelDebug
	case "INFO":
		l = slog.LevelInfo
	case "WARN":
		l = slog.LevelWarn
	case "ERROR":
		l = slog.LevelError
	default:
		l = slog.LevelDebug
	}

	var h slog.Handler
	switch strings.ToUpper(cfg.LoggingFormat) {
	case "TEXT":
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l, AddSource: false})
	case "JSON":
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l, AddSource: false})
	default:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l, AddSource: false})
	}

	return slog.New(h)
}
