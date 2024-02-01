package main

import (
	"context"
	"os"

	"github.com/netboxlabs/diode-internal/diode-server/distributor"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-distributor")

	distributorComponent, err := distributor.New(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate distributor component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(distributorComponent); err != nil {
		s.Logger().Error("failed to register distributor component: %v", err)
		os.Exit(1)
	}

	// instantiate a prom service for /metrics
	// prometheusSvc, err := prometheus.New()

	if err := s.Run(); err != nil {
		s.Logger().Error("server %s failure: %v", s.Name(), err)
		os.Exit(1)
	}
}
