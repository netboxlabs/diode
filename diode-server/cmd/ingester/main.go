package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	applicationName = "diode-ingester"
	metricStartup   = "diode.ingester.startup_count"
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

	meter := otel.GetMeterProvider().Meter(applicationName)
	startupCounter, err := meter.Int64Counter(metricStartup,
		metric.WithDescription("Number of times the ingester service has started"))
	if err != nil {
		s.Logger().Error("failed to create startup metric", "error", err)
		os.Exit(1)
	}
	startupCounter.Add(ctx, 1)

	ingesterComponent, err := ingester.New(ctx, s.Logger(), cfg, meter)
	if err != nil {
		s.Logger().Error("failed to instantiate ingester component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingester component", "error", err)
		os.Exit(1)
	}

	telemetry.ServePrometheusMetricsIfNecessary(cfg.Telemetry, s.Logger())

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
