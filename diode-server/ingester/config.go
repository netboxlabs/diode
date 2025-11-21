package ingester

import (
	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/tls"
)

// Config is the configuration for the ingester service
type Config struct {
	GRPCPort      int    `envconfig:"GRPC_PORT" default:"8081"`
	RedisHost     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisUsername string `envconfig:"REDIS_USERNAME" default:""`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisStreamDB int    `envconfig:"REDIS_STREAM_DB" default:"1"`

	RedisTLS  tls.Config       `envconfig:"REDIS_TLS"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
