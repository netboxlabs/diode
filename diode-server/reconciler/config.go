package reconciler

// Config is the configuration for the reconciler service
type Config struct {
	GRPCPort int `envconfig:"GRPC_PORT" default:"8081"`
}
