package main

import (
	"context"
	"log"

	"github.com/netboxlabs/diode-internal/diode-server/distributor"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-distributor")

	distributorComponent, err := distributor.New(s.Logger())
	if err != nil {
		log.Fatalf("failed to instantiate distributor component: %v", err)
	}

	if err := s.RegisterComponent(distributorComponent); err != nil {
		log.Fatalf("failed to register distributor component: %v", err)
	}

	// instantiate a prom service for /metrics
	// prometheusSvc, err := prometheus.New()

	if err := s.Run(); err != nil {
		log.Fatalf("server %s failure: %v", s.Name(), err)
	}
}
