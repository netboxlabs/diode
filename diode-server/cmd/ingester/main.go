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
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"

	"github.com/netboxlabs/diode/diode-server/authutil"
	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	applicationName = "diode-ingester" // used by sentry

	// used by open telemetry metrics
	telemetryServiceName = "netboxlabs/diode/ingester"
	metricStartup        = "netboxlabs/diode/ingester/startup_count"
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
	startupCounter, err := meter.Int64Counter(metricStartup,
		metric.WithDescription("Number of times the ingester service has started"))
	if err != nil {
		s.Logger().Error("failed to create startup metric", "error", err)
		os.Exit(1)
	}
	startupCounter.Add(ctx, 1)

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisStreamDB,
	})

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		s.Logger().Error("failed to connect to redis stream", "redisStream", redisStreamClient.String(), "error", err)
		os.Exit(1)
	}

	streamRouter := &ingester.DefaultStreamRouter{}
	authorizer := authutil.NewContextAuthorizer(s.Logger())
	ingesterComponent, err := ingester.New(ctx, s.Logger(), cfg, redisStreamClient, meter, streamRouter, serverInterceptors(authorizer, s.Logger())...)
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
