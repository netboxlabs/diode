package distributor

// Config is the configuration for the distributor service
type Config struct {
	GRPCPort      int    `envconfig:"GRPC_PORT" default:"8081"`
	RedisHost     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6378"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
}
