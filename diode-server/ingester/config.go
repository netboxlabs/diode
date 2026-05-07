package ingester

import (
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

	// RedisMemoryHighWatermarkBytes, when > 0, causes the ingester to reject new
	// Ingest requests with codes.ResourceExhausted once Redis used_memory crosses
	// this threshold. Applied globally regardless of tenant/stream. Zero disables
	// the check (current behavior). The value is checked at most once per second.
	RedisMemoryHighWatermarkBytes int64 `envconfig:"REDIS_MEMORY_HIGH_WATERMARK_BYTES" default:"0"`

	RedisTLS  tls.Config       `envconfig:"REDIS_TLS"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
