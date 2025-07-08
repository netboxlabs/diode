package main

import (
	"context"
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/version"
)

const (
	applicationName = "diode-auth" // used by sentry

	// used by open telemetry metrics
	telemetryServiceName = "netboxlabs/diode/auth"
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

	meter := otel.GetMeterProvider().Meter(telemetryServiceName)
	metricRecorder, err := auth.NewMetricRecorder(meter, cfg.Telemetry.Environment)
	if err != nil {
		s.Logger().Error("failed to create auth metrics", "error", err)
		os.Exit(1)
	}

	metricRecorder.SetServiceInfo(ctx, fmt.Sprintf("%s.%s", version.GetBuildVersion(), version.GetBuildCommit()))

	clientManager := auth.NewHydraClientManager(cfg.OAuth2.AdminServerURL, s.Logger())
	tokenOwner := &auth.DefaultTokenOwner{}
	httpServer, err := auth.NewServer(ctx, s.Logger(), auth.JWTParser{}, clientManager, tokenOwner)
	if err != nil {
		s.Logger().Error("failed to instantiate HTTP server", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	if err := s.RegisterComponent(httpServer); err != nil {
		s.Logger().Error("failed to register HTTP server", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	telemetry.ServePrometheusMetricsIfNecessary(cfg.Telemetry, s.Logger())

	metricRecorder.RecordServiceStartupAttempt(ctx, true)

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}
}
