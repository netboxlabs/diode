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
// not deadlock with an INSERT that spans the same cells. Both triggers have
// to lock counter cells in (state, object_type) order.
//
// Three transactions make a wrong order show:
//  1. holder inserts a dcim.site row and keeps its transaction open, so it
//     holds the middle cell (3, 'dcim.site').
//  2. writer inserts one row of every object type in one statement. Its
//     insert trigger locks the cells before (3, 'dcim.site') in its own
//     order, then waits for holder.
//  3. deleter deletes the seed rows. Its delete trigger waits on a cell that
//     writer holds.
//
// When holder commits, writer and deleter finish only if neither holds a
// cell the other still needs. The test catches an unordered delete trigger
// and a reversed insert trigger. It does not catch every reordering: an
// insert trigger that takes (3, 'dcim.device') first and the rest in another
// order still passes.
func TestStateCountsDeleteTriggerLocksCellsInOrder(t *testing.T) {
	ctx := context.Background()
	pool := startMigratedPostgres(t)

	// In (state, object_type) order the cells are dcim.device, dcim.interface,
	// dcim.site, ipam.ipaddress, ipam.prefix, ipam.vlan.
	objectTypes := []string{"ipam.vlan", "dcim.site", "ipam.prefix", "dcim.interface", "ipam.ipaddress", "dcim.device"}
	for _, objectType := range objectTypes {
		_, err := pool.Exec(ctx, `INSERT INTO ingestion_logs (external_id, object_type, state) VALUES ($1, $1, 3)`, objectType)
		require.NoError(t, err)
	}

	waitForBlocked := func(n int, msg string) {
		require.Eventually(t, func() bool {
			var waiting int
			err := pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND datname = current_database()`).Scan(&waiting)
			return err == nil && waiting >= n
		}, 10*time.Second, 20*time.Millisecond, msg)
	}

	holder, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = holder.Rollback(ctx) }()
	_, err = holder.Exec(ctx, `INSERT INTO ingestion_logs (external_id, object_type, state) VALUES ('held', 'dcim.site', 3)`)
	require.NoError(t, err)

	writerDone := make(chan error, 1)
	go func() {
		_, err := pool.Exec(ctx, `INSERT INTO ingestion_logs (external_id, object_type, state)
			SELECT 'next-' || ot, ot, 3 FROM unnest($1::text[]) ot`, objectTypes)
		writerDone <- err
	}()
	waitForBlocked(1, "the writer never blocked on the held counter cell")

	type result struct {
		deleted int64
		err     error
	}
	deleterDone := make(chan result, 1)
	go func() {
		tag, err := pool.Exec(ctx, `DELETE FROM ingestion_logs WHERE external_id = object_type`)
		deleterDone <- result{tag.RowsAffected(), err}
	}()
	waitForBlocked(2, "the deleter never blocked on a counter cell")

	require.NoError(t, holder.Commit(ctx))

	for range 2 {
		select {
		case err := <-writerDone:
			require.NoError(t, err, "writer")
		case r := <-deleterDone:
			require.NoError(t, r.err, "deleter")
			require.Equal(t, int64(len(objectTypes)), r.deleted)
		case <-time.After(10 * time.Second):
			t.Fatal("the writer or the deleter did not finish")
		}
	}

	// The seed rows are gone. The held row and the writer's rows remain.
	var total int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT COALESCE(SUM(n), 0) FROM ingestion_log_state_object_type_counts`).Scan(&total))
	require.Equal(t, int64(len(objectTypes)+1), total)
}
