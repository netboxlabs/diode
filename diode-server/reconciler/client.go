package reconciler

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
)

const (
	// SDKName is the name of the Diode SDK
	SDKName = "reconciler-sdk-go"

	// SDKVersion is the version of the Diode SDK
	SDKVersion = "0.1.0"

	// GRPCHostEnvVarName is the environment variable name for the reconciler gRPC host
	GRPCHostEnvVarName = "RECONCILER_GRPC_HOST"

	// GRPCPortEnvVarName is the environment variable name for the reconciler gRPC port
	GRPCPortEnvVarName = "RECONCILER_GRPC_PORT"

	defaultGRPCHost = "127.0.0.1"

	defaultGRPCPort = "8081"
)

// Client is an interface that defines the methods available from reconciler API
type Client interface {
	// Close closes the connection to the API service
	Close() error

	// RetrieveIngestionDataSources retrieves ingestion data sources
	RetrieveIngestionDataSources(context.Context, *pb.RetrieveIngestionDataSourcesRequest, ...grpc.CallOption) (*pb.RetrieveIngestionDataSourcesResponse, error)
}

// GRPCClient is a gRPC implementation of the distributor service
type GRPCClient struct {
	// gRPC virtual connection
	conn *grpc.ClientConn

	// The gRPC API client
	client pb.ReconcilerServiceClient
}

// Close closes the connection to the API service
func (g *GRPCClient) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

// RetrieveIngestionDataSources retrieves ingestion data sources
func (g *GRPCClient) RetrieveIngestionDataSources(ctx context.Context, req *pb.RetrieveIngestionDataSourcesRequest, opt ...grpc.CallOption) (*pb.RetrieveIngestionDataSourcesResponse, error) {
	req.SdkName = SDKName
	req.SdkVersion = SDKVersion
	return g.client.RetrieveIngestionDataSources(ctx, req, opt...)
}

// NewClient creates a new reconciler client based on gRPC
func NewClient() (Client, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithUserAgent(userAgent()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	target := grpcTarget()

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	c := &GRPCClient{
		conn:   conn,
		client: pb.NewReconcilerServiceClient(conn),
	}

	return c, nil
}

func userAgent() string {
	return fmt.Sprintf("%s/%s", SDKName, SDKVersion)
}

func grpcTarget() string {
	host, ok := os.LookupEnv(GRPCHostEnvVarName)
	if !ok {
		host = defaultGRPCHost
	}

	port, ok := os.LookupEnv(GRPCPortEnvVarName)
	if !ok {
		port = defaultGRPCPort
	}

	return fmt.Sprintf("%s:%s", host, port)
}
