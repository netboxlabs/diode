package main

import (
	"context"
	"log"

	"github.com/netboxlabs/diode-internal/diode-server/ingester"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-ingester")

	ingesterComponent := ingester.New(s.Logger())

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		log.Fatalf("failed to register ingerster component: %v", err)
	}

	if err := s.Run(); err != nil {
		log.Fatalf("server %s failure: %v", s.Name(), err)
	}
}
