package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"

	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	applicationName = "diode-ingester"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, applicationName)

	defer s.Recover(sentry.CurrentHub())

	// Load configuration
	var cfg ingester.Config
	envconfig.MustProcess("", &cfg)

	// Set default telemetry configuration if not provided
	if cfg.Telemetry.ServiceName == "" {
		cfg.Telemetry.ServiceName = applicationName
	}

	// Initialize telemetry
	shutdown, err := telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		s.Logger().Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer shutdown(ctx)

	ingesterComponent, err := ingester.New(ctx, s.Logger(), cfg)
	if err != nil {
		s.Logger().Error("failed to instantiate ingester component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingester component", "error", err)
		os.Exit(1)
	}

	// TODO: instantiate prometheus server

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
