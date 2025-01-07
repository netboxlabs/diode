package migrator

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigrator_Run(t *testing.T) {
	tests := []struct {
		name      string
		operation Operation
		expectErr bool
	}{
		{
			name:      "operation up",
			operation: OperationUp,
			expectErr: false,
		},
		{
			name:      "operation down",
			operation: OperationDown,
			expectErr: false,
		},
		{
			name:      "unknown operation",
			operation: "unknown",
			expectErr: true,
		},
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			m, err := NewMigrator(
				logger,
				"sqlite3",
				db,
				"testdata/migrations",
				"",
			)
			require.NoError(t, err)

			err = m.Run(context.Background(), tt.operation)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
