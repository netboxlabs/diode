package main

import (
	"context"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-ingester")

	defer func() {
		if err := recover(); err != nil {
			if sentry.CurrentHub().Client() != nil {
				eventID := sentry.CurrentHub().Recover(err)
				sentry.Flush(2 * time.Second)
				s.Logger().Error("recovered from panic", "error", err, "eventID", eventID)
			} else {
				s.Logger().Error("recovered from panic", "error", err)
			}
		}
	}()

	ingesterComponent, err := ingester.New(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate ingester component", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		s.Logger().Error("failed to register ingester component", "error", err)
		os.Exit(1)
	}

	//TODO: instantiate a prom service for /metrics

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
