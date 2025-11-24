package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx to database/sql compatibility
	"github.com/kelseyhightower/envconfig"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/netboxlabs/diode/diode-server/authutil"
	"github.com/netboxlabs/diode/diode-server/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/migrator"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/version"
)

const (
	applicationName = "diode-reconciler" // used by sentry

	// used by open telemetry metrics
	telemetryServiceName = "netboxlabs/diode/reconciler"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, applicationName, version.Release())

	defer s.Recover(sentry.CurrentHub())

	var cfg reconciler.Config
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
	metricRecorder, err := reconciler.NewMetricRecorder(meter, cfg.Telemetry.Environment)
	if err != nil {
		s.Logger().Error("failed to create ingester metrics", "error", err)
		os.Exit(1)
	}

	metricRecorder.SetServiceInfo(ctx, fmt.Sprintf("%s.%s", version.GetBuildVersion(), version.GetBuildCommit()))

	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDBName, cfg.PostgresSSLMode)

	if cfg.MigrationEnabled {
		if err := runDBMigrations(ctx, s.Logger(), dbURL); err != nil {
			s.Logger().Error("failed to run db migrations", "error", err)
			metricRecorder.RecordServiceStartupAttempt(ctx, false)
			os.Exit(1)
		}
	}

	redisTLSConfig, err := cfg.RedisTLS.ToTLSConfig()
	if err != nil {
		s.Logger().Error("failed to create TLS config for Redis", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	redisOptions := redis.Options{
		Addr:      fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:  cfg.RedisPassword,
		DB:        cfg.RedisDB,
		TLSConfig: redisTLSConfig,
	}
	if cfg.RedisUsername != "" {
		redisOptions.Username = cfg.RedisUsername
	}
	redisClient := redis.NewClient(&redisOptions)

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		s.Logger().Error("failed to connect to redis", "redis", redisClient.String(), "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	redisStreamOptions := redis.Options{
		Addr:      fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:  cfg.RedisPassword,
		DB:        cfg.RedisStreamDB,
		TLSConfig: redisTLSConfig,
	}
	if cfg.RedisUsername != "" {
		redisStreamOptions.Username = cfg.RedisUsername
	}
	redisStreamClient := redis.NewClient(&redisStreamOptions)

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		s.Logger().Error("failed to connect to redis stream", "redisStream", redisStreamClient.String(), "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		s.Logger().Error("failed to connect to postgres database", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}
	defer dbPool.Close()

	repository := postgres.NewRepository(dbPool)

	diodeToNetBoxMaxRetries := 3

	nbClient, err := netboxdiodeplugin.NewClient(
		netboxdiodeplugin.ClientOptions{
			Logger:            s.Logger(),
			BaseURL:           cfg.NetBoxDiodePluginAPIBaseURL,
			ClientID:          cfg.DiodeToNetBoxClientID,
			ClientSecret:      cfg.DiodeToNetBoxClientSecret,
			TokenURL:          cfg.DiodeAuthTokenURL,
			RateLimitRPS:      cfg.DiodeToNetBoxRateLimiterRPS,
			RateLimitBurstRPS: cfg.DiodeToNetBoxRateLimiterBurst,
			MaxRetries:        diodeToNetBoxMaxRetries,
		})
	if err != nil {
		s.Logger().Error("failed to create netbox diode plugin client", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	ops := reconciler.NewOps(repository, nbClient, s.Logger(), nil)

	ingestionProcessor, err := reconciler.NewIngestionProcessor(ctx, s.Logger(), cfg, redisClient, redisStreamClient, reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup, ops, metricRecorder)
	if err != nil {
		s.Logger().Error("failed to instantiate ingestion processor", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingestionProcessor); err != nil {
		s.Logger().Error("failed to register ingestion processor", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	authorizer := authutil.NewContextAuthorizer(s.Logger())
	gRPCServer, err := reconciler.NewServer(ctx, s.Logger(), repository, serverInterceptors(authorizer, s.Logger())...)
	if err != nil {
		s.Logger().Error("failed to instantiate gRPC server", "error", err)
		metricRecorder.RecordServiceStartupAttempt(ctx, false)
		os.Exit(1)
	}

	if err := s.RegisterComponent(gRPCServer); err != nil {
		s.Logger().Error("failed to register gRPC server", "error", err)
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

func runDBMigrations(ctx context.Context, logger *slog.Logger, dbURL string) error {
	dbDialect := "postgres"
	db, err := goose.OpenDBWithDriver(dbDialect, dbURL)
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close connection to database", "error", err)
		}
	}()

	m, err := migrator.NewMigrator(logger, "postgres", db, "/etc/diode/migrations", "")
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}
	if err := m.Run(ctx, migrator.OperationUp); err != nil {
		return err
	}

	return nil
}

func serverInterceptors(authorizer authutil.Authorizer, logger *slog.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		authutil.NewUnverifiedJWTInterceptor(logger),
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			// TODO: this is applied to all rpcs but could be checked per rpc
			// if the permissions differ (all are reads currently)
			if err := authorizer.RequireScopes(ctx, []string{authutil.ScopeDiodeRead}); err != nil {
				return nil, err
			}

			return handler(ctx, req)
		},
	}
}
