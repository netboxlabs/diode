package ingester

import (
	"time"

	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/tls"
)

// Config is the configuration for the ingester service
type Config struct {
	GRPCPort      int    `envconfig:"GRPC_PORT" default:"8081"`
	PProfAddr     string `envconfig:"PPROF_ADDR"`
	RedisHost     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisUsername string `envconfig:"REDIS_USERNAME" default:""`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisStreamDB int    `envconfig:"REDIS_STREAM_DB" default:"1"`

	// CompressStreamMessages enables Brotli compression for Redis stream messages.
	// When enabled, protobuf payloads are compressed before XADD and an "encoding":"br"
	// field is added to the message. The reconciler auto-detects compressed vs raw messages.
	CompressStreamMessages bool `envconfig:"COMPRESS_STREAM_MESSAGES" default:"true"`

	// RedisMemoryHighWatermarkPct, when > 0, causes the ingester to reject new
	// Ingest requests with codes.ResourceExhausted once Redis used_memory reaches
	// this percentage of maxmemory. Applied globally regardless of tenant/stream.
	// Zero disables the check (current behavior). The value is checked at most
	// once per RedisMemoryCheckInterval. Has no effect when Redis maxmemory is
	// 0 (unlimited).
	RedisMemoryHighWatermarkPct int `envconfig:"REDIS_MEMORY_HIGH_WATERMARK_PCT" default:"0"`

	// RedisMemoryCheckInterval bounds how often INFO memory is issued when
	// the watermark check is enabled. Smaller values reduce the staleness
	// window (and thus the size of an admit-burst that can slip past the
	// threshold) at the cost of more INFO calls. Values <= 0 fall back to
	// the default. The redis_memory_used_ratio_bps gauge is updated on
	// each successful poll, so its freshness tracks this interval.
	RedisMemoryCheckInterval time.Duration `envconfig:"REDIS_MEMORY_CHECK_INTERVAL" default:"500ms"`

	// RedisMemoryCheckTimeout bounds the INFO memory call itself. The
	// watermark mutex is held across this I/O, so an unbounded INFO would
	// stall every concurrent Ingest until the redis-client ReadTimeout
	// fires. Tune this down on latency-sensitive deployments; tune up if
	// running against a Redis that occasionally needs longer than the
	// default. Values <= 0 fall back to the default.
	RedisMemoryCheckTimeout time.Duration `envconfig:"REDIS_MEMORY_CHECK_TIMEOUT" default:"250ms"`

	RedisTLS  tls.Config       `envconfig:"REDIS_TLS"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
