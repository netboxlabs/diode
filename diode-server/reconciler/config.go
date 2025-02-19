package reconciler

import "github.com/netboxlabs/diode/diode-server/telemetry"

// Config is the configuration for the reconciler service
type Config struct {
	GRPCPort                   int    `envconfig:"GRPC_PORT" default:"8081"`
	RedisHost                  string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort                  string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword              string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisDB                    int    `envconfig:"REDIS_DB" default:"0"`
	RedisStreamDB              int    `envconfig:"REDIS_STREAM_DB" default:"1"`
	MigrationEnabled           bool   `envconfig:"MIGRATION_ENABLED" default:"true"`
	AutoApplyChangesets        bool   `envconfig:"AUTO_APPLY_CHANGESETS" default:"true"`
	ReconcilerRateLimiterRPS   int    `envconfig:"RECONCILER_RATE_LIMITER_RPS" default:"20"`
	ReconcilerRateLimiterBurst int    `envconfig:"RECONCILER_RATE_LIMITER_BURST" default:"1"`
	PostgresHost               string `envconfig:"POSTGRES_HOST"`
	PostgresPort               int    `envconfig:"POSTGRES_PORT"`
	PostgresDBName             string `envconfig:"POSTGRES_DB_NAME"`
	PostgresUser               string `envconfig:"POSTGRES_USER"`
	PostgresPassword           string `envconfig:"POSTGRES_PASSWORD"`

	// API keys
	DiodeToNetBoxAPIKey string `envconfig:"DIODE_TO_NETBOX_API_KEY" required:"true"`
	NetBoxToDiodeAPIKey string `envconfig:"NETBOX_TO_DIODE_API_KEY" required:"true"`

	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
