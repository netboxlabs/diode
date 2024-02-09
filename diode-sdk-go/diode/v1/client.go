package diode

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	// SDKName is the name of the Diode SDK
	SDKName = "diode-sdk-go"

	// SDKVersion is the version of the Diode SDK
	SDKVersion = "0.1.0"

	// DiodeAPIKeyEnvVarName is the environment variable name for the Diode API key
	DiodeAPIKeyEnvVarName = "DIODE_API_KEY"

	// DiodeGRPCInsecureEnvVarName is the environment variable name for the Diode gRPC disabling transport security
	DiodeGRPCInsecureEnvVarName = "DIODE_GRPC_INSECURE"

	// DiodeGRPCHostEnvVarName is the environment variable name for the Diode gRPC host
	DiodeGRPCHostEnvVarName = "DIODE_GRPC_HOST"

	// DiodeGRPCPortEnvVarName is the environment variable name for the Diode gRPC port
	DiodeGRPCPortEnvVarName = "DIODE_GRPC_PORT"

	// DiodeGRPCTimeoutSecondsEnvVarName is the environment variable name for the Diode gRPC timeout in seconds
	DiodeGRPCTimeoutSecondsEnvVarName = "DIODE_GRPC_TIMEOUT_SECONDS"

	authAPIKeyName = "diode-api-key"

	defaultGRPCHost = "127.0.0.1"

	defaultGRPCPort = "8081"

	defaultGRPCTimeoutSeconds = 5
)

var (
	// ErrInvalidTimeout is an error for invalid timeout value
	ErrInvalidTimeout = errors.New("invalid timeout value")
)

// Client is an interface that defines the methods available from Diode API
type Client interface {
	// Close closes the connection to the API service
	Close() error

	// Push sends a push request to the distributor service
	Push(context.Context, *diodepb.PushRequest, ...grpc.CallOption) (*diodepb.PushResponse, error)
}

// GRPCClient is a gRPC implementation of the distributor service
type GRPCClient struct {
	// gRPC virtual connection
	conn *grpc.ClientConn

	// The gRPC API client
	client diodepb.DistributorServiceClient

	// An API key for the Diode API
	apiKey *string
}

// Close closes the connection to the API service
func (g *GRPCClient) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

// Push sends a push request to the distributor service
func (g *GRPCClient) Push(ctx context.Context, req *diodepb.PushRequest, opt ...grpc.CallOption) (*diodepb.PushResponse, error) {
	return g.client.Push(ctx, req, opt...)
}

func authUnaryInterceptor(apiKey string) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if apiKey != "" {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(authAPIKeyName, apiKey))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

// NewClient creates a new distributor client based on gRPC
func NewClient(ctx context.Context) (Client, error) {
	apiKey, ok := os.LookupEnv(DiodeAPIKeyEnvVarName)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not found", DiodeAPIKeyEnvVarName)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithUserAgent(userAgent()),
		authUnaryInterceptor(apiKey),
	}

	if grpcInsecure() {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(new(tls.Config))))
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
		client: diodepb.NewDistributorServiceClient(conn),
		apiKey: &apiKey,
	}

	return c, nil
}

func userAgent() string {
	return fmt.Sprintf("%s/%s", SDKName, SDKVersion)
}

func grpcTarget() string {
	host, ok := os.LookupEnv(DiodeGRPCHostEnvVarName)
	if !ok {
		host = defaultGRPCHost
	}

	port, ok := os.LookupEnv(DiodeGRPCPortEnvVarName)
	if !ok {
		port = defaultGRPCPort
	}

	return fmt.Sprintf("%s:%s", host, port)
}

func grpcTimeout() (time.Duration, error) {
	timeoutSecondsStr, ok := os.LookupEnv(DiodeGRPCTimeoutSecondsEnvVarName)
	if !ok || len(timeoutSecondsStr) == 0 {
		return defaultGRPCTimeoutSeconds * time.Second, nil
	}

	timeout, err := strconv.Atoi(timeoutSecondsStr)
	if err != nil || timeout <= 0 {
		return 0, ErrInvalidTimeout
	}
	return time.Duration(timeout) * time.Second, nil
}

func grpcInsecure() bool {
	insecureStr, ok := os.LookupEnv(DiodeGRPCInsecureEnvVarName)
	if !ok {
		return false
	}

	insecureVal, err := strconv.ParseBool(insecureStr)
	if err != nil {
		return false
	}

	return insecureVal
}
