package reconciler

// Config is the configuration for the reconciler service
type Config struct {
	GRPCPort      int    `envconfig:"GRPC_PORT" default:"8081"`
	RedisHost     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
	RedisStreamDB int    `envconfig:"REDIS_STREAM_DB" default:"1"`

	// API keys
	DiodeToNetBoxAPIKey string `envconfig:"DIODE_TO_NETBOX_API_KEY" required:"true"`
	NetBoxToDiodeAPIKey string `envconfig:"NETBOX_TO_DIODE_API_KEY" required:"true"`
	DiodeAPIKey         string `envconfig:"DIODE_API_KEY" required:"true"`
}
