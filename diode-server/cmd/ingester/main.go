package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"

	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-ingester")

	defer s.Recover(sentry.CurrentHub())

	ingesterComponent, err := ingester.New(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate ingester component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingester component", "error", err)
		os.Exit(1)
	}

	//TODO: instantiate prometheus server

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
