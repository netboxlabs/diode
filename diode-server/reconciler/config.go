package reconciler

type Config struct {
	GRPCPort int `envconfig:"GRPC_PORT" default:"8081"`
}
