package reconciler

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

// StreamLengthReader is the subset of redis.Client needed to size the ingest stream.
type StreamLengthReader interface {
	XLen(ctx context.Context, stream string) *redis.IntCmd
}

// NewStreamLengthBackpressure returns a BackpressureFunc that reports true while
// the ingest stream holds more than threshold entries, so the database
// processors yield to the consume loop until it has drained the burst.
//
// A threshold of zero or less disables the gate and returns nil, which the
// processors treat as "never back off". Both processors poll the function
// every second while blocked, so a transition is logged once rather than on
// every poll: without that, a stalled gate is invisible in the logs.
func NewStreamLengthBackpressure(logger *slog.Logger, client StreamLengthReader, streamID string, threshold int64) BackpressureFunc {
	if threshold <= 0 {
		logger.Info("ingest stream backpressure disabled", "stream", streamID)
		return nil
	}

	var engaged atomic.Bool

	return func(ctx context.Context) bool {
		xlen, err := client.XLen(ctx, streamID).Result()
		if err != nil {
			// Fail open: a Redis hiccup must not stall reconciliation.
			return false
		}

		above := xlen > threshold
		switch {
		case above && engaged.CompareAndSwap(false, true):
			logger.Warn("ingest stream above backpressure threshold, pausing ingestion log processing until it drains",
				"stream", streamID, "xlen", xlen, "threshold", threshold)
		case !above && engaged.CompareAndSwap(true, false):
			logger.Info("ingest stream back below backpressure threshold, resuming ingestion log processing",
				"stream", streamID, "xlen", xlen, "threshold", threshold)
		}
		return above
	}
}
