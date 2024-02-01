package distributor

// Config is the configuration for the distributor service
type Config struct {
	GRPCPort int `envconfig:"GRPC_PORT" default:"8081"`
}
