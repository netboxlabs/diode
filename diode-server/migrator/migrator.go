package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

// Operation is the type of migration operation
type Operation string

const (
	// OperationUp applies all pending migrations
	OperationUp Operation = "up"
	// OperationDown rolls back the most recently applied migration
	OperationDown Operation = "down"
)

// A Migrator runs migrations against a database
type Migrator struct {
	logger         *slog.Logger
	migrationsPath string
	provider       *goose.Provider
}

// NewMigrator creates a new migrator
func NewMigrator(logger *slog.Logger, dialect string, db *sql.DB, migrationsPath string) (*Migrator, error) {
	migrationsFS := os.DirFS(migrationsPath)
	provider, err := goose.NewProvider(goose.Dialect(dialect), db, migrationsFS)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration provider: %w", err)
	}

	return &Migrator{
		logger:         logger,
		migrationsPath: migrationsPath,
		provider:       provider,
	}, nil
}

// Run runs the migrations
func (m *Migrator) Run(ctx context.Context, op Operation) error {
	switch op {
	case OperationUp:
		results, err := m.provider.Up(ctx)
		if err != nil && !errors.Is(err, goose.ErrAlreadyApplied) {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		m.logger.Debug("applied migrations", "results", results)
	case OperationDown:
		results, err := m.provider.Down(ctx)
		if err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
			return fmt.Errorf("failed to rollback migrations: %w", err)
		}
		m.logger.Debug("rolled back migrations", "results", results)
	default:
		return fmt.Errorf("unsupported operation: %s", op)
	}

	return nil
}
