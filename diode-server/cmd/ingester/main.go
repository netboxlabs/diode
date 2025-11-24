package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/netboxlabs/diode/diode-server/authutil"
	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/version"
)

const (
	applicationName = "diode-ingester" // used by sentry

	// used by open telemetry metrics
	telemetryServiceName = "netboxlabs/diode/ingester"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, applicationName, version.Release())

	defer s.Recover(sentry.CurrentHub())

	// Load configuration
	var cfg ingester.Config
	envconfig.MustProcess("", &cfg)

	// Set default telemetry configuration if not provided
	if cfg.Telemetry.ServiceName == "" {
		cfg.Telemetry.ServiceName = telemetryServiceName
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

	meter := otel.GetMeterProvider().Meter(telemetryServiceName)
	metricRecorder, err := ingester.NewMetricRecorder(meter, cfg.Telemetry.Environment)
	if err != nil {
		s.Logger().Error("failed to create ingester metrics", "error", err)
		os.Exit(1)
	}

	metricRecorder.SetServiceInfo(ctx, fmt.Sprintf("%s.%s", version.GetBuildVersion(), version.GetBuildCommit()))

	redisTLSConfig, err := cfg.RedisTLS.ToTLSConfig()
	if err != nil {
		s.Logger().Error("failed to create TLS config for Redis", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	redisOptions := redis.Options{
		Addr:      fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:  cfg.RedisPassword,
		DB:        cfg.RedisStreamDB,
		TLSConfig: redisTLSConfig,
	}
	if cfg.RedisUsername != "" {
		redisOptions.Username = cfg.RedisUsername
	}
	redisStreamClient := redis.NewClient(&redisOptions)

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		s.Logger().Error("failed to connect to redis stream", "redisStream", redisStreamClient.String(), "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	streamRouter := &ingester.DefaultStreamRouter{}
	authorizer := authutil.NewContextAuthorizer(s.Logger())
	ingesterComponent, err := ingester.New(ctx, s.Logger(), cfg, redisStreamClient, metricRecorder, streamRouter, serverInterceptors(authorizer, s.Logger())...)
	if err != nil {
		s.Logger().Error("failed to instantiate ingester component", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingester component", "error", err)
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

func serverInterceptors(authorizer authutil.Authorizer, logger *slog.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		authutil.NewUnverifiedJWTInterceptor(logger),
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if err := authorizer.RequireScopes(ctx, []string{authutil.ScopeDiodeIngest}); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		},
	}
}
