package main

import (
	"context"
	"os"

	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-reconciler")

	reconcilerComponent, err := reconciler.New(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate reconciler component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(reconcilerComponent); err != nil {
		s.Logger().Error("failed to register reconciler component", "error", err)
		os.Exit(1)
	}

	// instantiate a prom service for /metrics
	// prometheusSvc, err := prometheus.New()

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
