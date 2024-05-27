package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/run"

	"github.com/netboxlabs/diode/diode-server/version"
)

// A Server is a diode Server
type Server struct {
	ctx         context.Context
	name        string
	environment string
	release     string
	logger      *slog.Logger

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

	logger := newLogger(cfg)

	if cfg.SentryDSN != "" {
		logger.Info("initializing sentry")
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.Environment,
			Debug:            cfg.SentryDebug,
			SampleRate:       cfg.SentrySampleRate,
			EnableTracing:    cfg.SentryEnableTracing,
			TracesSampleRate: cfg.SentryTracesSampleRate,
			AttachStacktrace: cfg.SentryAttachStacktrace,
			ServerName:       name,
			Release:          fmt.Sprintf("v%s", version.GetBuildVersion()),
		}); err != nil {
			logger.Error("failed to initialize sentry", "error", err)
		}
	}

	return &Server{
		ctx:            ctx,
		name:           name,
		environment:    cfg.Environment,
		release:        fmt.Sprintf("v%s-%s", version.GetBuildVersion(), version.GetBuildCommit()),
		logger:         logger,
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

// RegisterComponent registers a Component with the Server.
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
			componentHub := sentry.CurrentHub().Clone()
			componentHub.Scope().SetTag("component", c.Name())

			defer func() {
				if err := recover(); err != nil {
					eventID := componentHub.Recover(err)
					componentHub.Flush(2 * time.Second)
					s.logger.Warn("recovered from panic", "componentName", c.Name(), "eventID", eventID)
				}
			}()

			return c.Start(ctx)
		},
		func(err error) {
			s.logger.Debug("component interrupted", "componentName", c.Name(), "error", err)
			if err2 := c.Stop(); err2 != nil {
				s.logger.Error("failed to stop component", "componentName", c.Name(), "error", err2)
			}
			cancel()
		},
	)
	return nil
}

// Run starts the diode Server
func (s *Server) Run() error {
	s.logger.Info("starting server", "serverName", s.name, "environment", s.environment, "release", s.release)
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
