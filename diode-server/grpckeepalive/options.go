// Package grpckeepalive provides shared gRPC server keepalive settings for Diode
// services. Values mirror typical Diode SDK client keepalive (e.g. 10s ping
// interval, 20s ping timeout) so servers send path traffic through idle timeouts
// and do not enforce the default 5m minimum between client PINGs (which would
// fight those clients).
package grpckeepalive

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	serverPingInterval    = 10 * time.Second
	serverPingTimeout     = 20 * time.Second
	minClientPingInterval = 5 * time.Second
)

// ServerOptions returns grpc.ServerOptions for production Diode gRPC servers.
func ServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    serverPingInterval,
			Timeout: serverPingTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             minClientPingInterval,
			PermitWithoutStream: true,
		}),
	}
}
