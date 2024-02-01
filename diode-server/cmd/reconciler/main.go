package main

import (
	"context"
	"log"

	"github.com/netboxlabs/diode-internal/diode-server/reconciler"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-reconciler")

	reconcilerComponent, err := reconciler.New(s.Logger())
	if err != nil {
		log.Fatalf("failed to instantiate reconciler component: %v", err)
	}

	if err := s.RegisterComponent(reconcilerComponent); err != nil {
		log.Fatalf("failed to register reconciler component: %v", err)
	}

	// instantiate a prom service for /metrics
	// prometheusSvc, err := prometheus.New()

	if err := s.Run(); err != nil {
		log.Fatalf("server %s failure: %v", s.Name(), err)
	}
}
