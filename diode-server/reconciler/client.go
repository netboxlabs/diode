package reconciler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	pb "github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	// GRPCTimeoutSecondsEnvVarName is the environment variable name for the reconciler gRPC timeout in seconds
	GRPCTimeoutSecondsEnvVarName = "RECONCILER_GRPC_TIMEOUT_SECONDS"

	defaultGRPCHost = "127.0.0.1"

	defaultGRPCPort = "8082"

	defaultGRPCTimeoutSeconds = 5
)

var (
	// ErrInvalidTimeout is an error for invalid timeout value
	ErrInvalidTimeout = errors.New("invalid timeout value")
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
	return g.client.RetrieveIngestionDataSources(ctx, req, opt...)
}

// NewClient creates a new reconciler client based on gRPC
func NewClient(ctx context.Context) (Client, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithUserAgent(userAgent()),
		grpc.WithTransportCredentials(credentials.NewTLS(new(tls.Config))),
	}

	timeout, err := grpcTimeout()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	target := grpcTarget()

	conn, err := grpc.DialContext(ctx, target, dialOpts...)
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

func grpcTimeout() (time.Duration, error) {
	timeoutSecondsStr, ok := os.LookupEnv(GRPCTimeoutSecondsEnvVarName)
	if !ok || len(timeoutSecondsStr) == 0 {
		return defaultGRPCTimeoutSeconds * time.Second, nil
	}

	timeout, err := strconv.Atoi(timeoutSecondsStr)
	if err != nil || timeout <= 0 {
		return 0, ErrInvalidTimeout
	}
	return time.Duration(timeout) * time.Second, nil
}
