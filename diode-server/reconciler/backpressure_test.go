package reconciler_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/reconciler"
)

func TestNewStreamLengthBackpressureDisabledAtZero(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	require.Nil(t, reconciler.NewStreamLengthBackpressure(logger, nil, "stream", 0))
	require.Nil(t, reconciler.NewStreamLengthBackpressure(logger, nil, "stream", -1))
}

func TestStreamLengthBackpressureTracksThresholdAndLogsTransitionsOnce(t *testing.T) {
	ctx := context.Background()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	const stream = "diode.v1.ingest-stream"
	backpressure := reconciler.NewStreamLengthBackpressure(logger, client, stream, 2)
	require.NotNil(t, backpressure)

	require.False(t, backpressure(ctx), "an empty stream must not back off")
	require.Empty(t, logs.String(), "no transition, nothing to log")

	for range 3 {
		require.NoError(t, client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: map[string]any{"request": "x"}}).Err())
	}

	require.True(t, backpressure(ctx))
	require.True(t, backpressure(ctx), "stays engaged while the stream is above the threshold")
	require.Equal(t, 1, strings.Count(logs.String(), "pausing ingestion log processing"),
		"engaging is logged once, not on every one-second poll")

	require.NoError(t, client.Del(ctx, stream).Err())

	require.False(t, backpressure(ctx))
	require.False(t, backpressure(ctx))
	require.Equal(t, 1, strings.Count(logs.String(), "resuming ingestion log processing"))
	require.Equal(t, 1, strings.Count(logs.String(), "pausing ingestion log processing"))
}

func TestStreamLengthBackpressureFailsOpenOnRedisError(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = client.Close() })
	s.Close()

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	backpressure := reconciler.NewStreamLengthBackpressure(logger, client, "stream", 1)

	require.False(t, backpressure(context.Background()), "a Redis error must not stall reconciliation")
}
