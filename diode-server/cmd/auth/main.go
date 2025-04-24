package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	applicationName = "diode-auth" // used by sentry

	// used by open telemetry metrics
	telemetryServiceName = "netboxlabs/diode/auth"
	metricStartup        = "netboxlabs/diode/auth/startup_count"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, applicationName)

	defer s.Recover(sentry.CurrentHub())

	var cfg auth.Config
	envconfig.MustProcess("", &cfg)

	// Set default telemetry configuration if not provided
	if cfg.Telemetry.ServiceName == "" {
		cfg.Telemetry.ServiceName = telemetryServiceName
	}

	// Initialize telemetry
	shutdown, err := telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		s.Logger().Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			s.Logger().Error("failed to shutdown telemetry", "error", err)
		}
	}()

	appMeter := otel.GetMeterProvider().Meter(telemetryServiceName)
	startupCounter, err := appMeter.Int64Counter(metricStartup,
		metric.WithDescription("Number of times the auth service has started"))
	if err != nil {
		s.Logger().Error("failed to create startup metric", "error", err)
		os.Exit(1)
	}
	startupCounter.Add(ctx, 1)

	httpServer, err := auth.NewServer(ctx, s.Logger(), auth.JWTParser{})
	if err != nil {
		s.Logger().Error("failed to instantiate HTTP server", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(httpServer); err != nil {
		s.Logger().Error("failed to register HTTP server", "error", err)
		os.Exit(1)
	}

	telemetry.ServePrometheusMetricsIfNecessary(cfg.Telemetry, s.Logger())

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
