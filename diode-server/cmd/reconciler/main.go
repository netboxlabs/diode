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

	"github.com/netboxlabs/diode/diode-server/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/migrator"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-reconciler")

	defer s.Recover(sentry.CurrentHub())

	var cfg reconciler.Config
	envconfig.MustProcess("", &cfg)

	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDBName)

	if err := runDBMigrations(ctx, s.Logger(), dbURL); err != nil {
		s.Logger().Error("failed to run db migrations", "error", err)
		os.Exit(1)
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		s.Logger().Error("failed to connect to postgres database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	ingestionLogRepo := postgres.NewIngestionLogRepository(dbPool)
	changeSetRepo := postgres.NewChangeSetRepository(dbPool)

	ingestionProcessor, err := reconciler.NewIngestionProcessor(ctx, s.Logger(), ingestionLogRepo, changeSetRepo)
	if err != nil {
		s.Logger().Error("failed to instantiate ingestion processor", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingestionProcessor); err != nil {
		s.Logger().Error("failed to register ingestion processor", "error", err)
		os.Exit(1)
	}

	gRPCServer, err := reconciler.NewServer(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate gRPC server", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(gRPCServer); err != nil {
		s.Logger().Error("failed to register gRPC server", "error", err)
		os.Exit(1)
	}

	// TODO: instantiate prometheus server

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
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

	m, err := migrator.NewMigrator(logger, "postgres", db, "/etc/diode/migrations")
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}
	if err := m.Run(ctx, migrator.OperationUp); err != nil {
		return err
	}

	return nil
}
