package main

import (
	"context"
	"os"

	"github.com/netboxlabs/diode-internal/diode-server/ingester"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-ingester")

	ingesterComponent := ingester.New(s.Logger())

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingerster component: %v", err)
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		s.Logger().Error("server %s failure: %v", s.Name(), err)
		os.Exit(1)
	}
}
