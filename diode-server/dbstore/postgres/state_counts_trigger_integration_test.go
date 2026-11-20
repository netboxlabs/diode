//go:build integration_test

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx to database/sql compatibility, for goose
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/netboxlabs/diode/diode-server/migrator"
)

// startMigratedPostgres starts a Postgres container, applies the diode
// migrations and returns a pool connected to it.
func startMigratedPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env:          map[string]string{"POSTGRES_USER": "diode", "POSTGRES_PASSWORD": "diode", "POSTGRES_DB": "diode"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(ctr) })

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("postgres://diode:diode@%s:%s/diode?sslmode=disable", host, port.Port())

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	m, err := migrator.NewMigrator(slog.Default(), "postgres", db, "migrations", "")
	require.NoError(t, err)
	require.NoError(t, m.Run(ctx, migrator.OperationUp))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

// A DELETE whose rows span several (state, object_type) counter cells must
// not deadlock with a state transition that locks counter cells in
// (state, object_type) order, the order the insert and update triggers use.
func TestStateCountsDeleteTriggerLocksCellsInOrder(t *testing.T) {
	ctx := context.Background()
	pool := startMigratedPostgres(t)

	objectTypes := []string{"ipam.vlan", "dcim.site", "ipam.prefix", "dcim.interface", "ipam.ipaddress", "dcim.device"}
	for _, objectType := range objectTypes {
		_, err := pool.Exec(ctx, `INSERT INTO ingestion_logs (external_id, object_type, state) VALUES ($1, $1, 3)`, objectType)
		require.NoError(t, err)
	}

	// The state transition holds the first cell in (state, object_type)
	// order and is about to take the rest.
	transition, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = transition.Rollback(ctx) }()
	_, err = transition.Exec(ctx, `SELECT 1 FROM ingestion_log_state_object_type_counts WHERE state = 3 AND object_type = 'dcim.device' FOR UPDATE`)
	require.NoError(t, err)

	type result struct {
		deleted int64
		err     error
	}
	done := make(chan result, 1)
	go func() {
		tag, err := pool.Exec(ctx, `DELETE FROM ingestion_logs`)
		done <- result{tag.RowsAffected(), err}
	}()
	require.Eventually(t, func() bool {
		var waiting int
		err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND datname = current_database()`).Scan(&waiting)
		return err == nil && waiting > 0
	}, 10*time.Second, 20*time.Millisecond, "the delete never blocked on the held counter cell")

	_, err = transition.Exec(ctx, `SELECT 1 FROM ingestion_log_state_object_type_counts WHERE state = 3 ORDER BY state, object_type FOR UPDATE`)
	require.NoError(t, err)
	require.NoError(t, transition.Commit(ctx))

	select {
	case r := <-done:
		require.NoError(t, r.err)
		require.Equal(t, int64(len(objectTypes)), r.deleted)
	case <-time.After(10 * time.Second):
		t.Fatal("the delete did not finish")
	}

	var total int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT COALESCE(SUM(n), 0) FROM ingestion_log_state_object_type_counts`).Scan(&total))
	require.Equal(t, int64(0), total)
}
