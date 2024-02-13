package ingester

// Config is the configuration for the ingester service
type Config struct {
	RedisHost     string `envconfig:"REDIS_HOST" default:"127.0.0.1"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisStreamDB int    `envconfig:"REDIS_STREAM_DB" default:"0"`
	RedisDB       int    `envconfig:"REDIS_DB" default:"1"`
}
