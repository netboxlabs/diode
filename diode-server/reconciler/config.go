package reconciler

import (
	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/tls"
)

// Config is the configuration for the reconciler service
type Config struct {
	GRPCPort                      int    `envconfig:"GRPC_PORT" default:"8081"`
	RedisHost                     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort                     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisUsername                 string `envconfig:"REDIS_USERNAME" default:""`
	RedisPassword                 string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisDB                       int    `envconfig:"REDIS_DB" default:"0"`
	RedisStreamDB                 int    `envconfig:"REDIS_STREAM_DB" default:"1"`
	MigrationEnabled              bool   `envconfig:"MIGRATION_ENABLED" default:"true"`
	AutoApplyChangesets           bool   `envconfig:"AUTO_APPLY_CHANGESETS" default:"true"`
	ReconcilerRateLimiterRPS      int    `envconfig:"RECONCILER_RATE_LIMITER_RPS" default:"20"`
	ReconcilerRateLimiterBurst    int    `envconfig:"RECONCILER_RATE_LIMITER_BURST" default:"1"`
	DiodeToNetBoxRateLimiterRPS   int    `envconfig:"DIODE_TO_NETBOX_RATE_LIMITER_RPS" default:"20"`
	DiodeToNetBoxRateLimiterBurst int    `envconfig:"DIODE_TO_NETBOX_RATE_LIMITER_BURST" default:"1"`
	PostgresHost                  string `envconfig:"POSTGRES_HOST"`
	PostgresPort                  int    `envconfig:"POSTGRES_PORT"`
	PostgresDBName                string `envconfig:"POSTGRES_DB_NAME"`
	PostgresUser                  string `envconfig:"POSTGRES_USER"`
	PostgresPassword              string `envconfig:"POSTGRES_PASSWORD"`
	PostgresSSLMode               string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`

	NetBoxDiodePluginAPIBaseURL    string `envconfig:"NETBOX_DIODE_PLUGIN_API_BASE_URL" required:"true"`
	NetBoxDiodePluginSkipTLSVerify bool   `envconfig:"NETBOX_DIODE_PLUGIN_SKIP_TLS_VERIFY" default:"false"`
	DiodeAuthTokenURL              string `envconfig:"DIODE_AUTH_TOKEN_URL" required:"true"`
	DiodeToNetBoxClientID          string `envconfig:"DIODE_TO_NETBOX_CLIENT_ID" required:"true"`
	DiodeToNetBoxClientSecret      string `envconfig:"DIODE_TO_NETBOX_CLIENT_SECRET" required:"true"`

	RedisTLS  tls.Config       `envconfig:"REDIS_TLS"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
