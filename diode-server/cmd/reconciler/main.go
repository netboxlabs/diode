package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"

	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/server"
)

func main() {
	ctx := context.Background()
	s := server.New(ctx, "diode-reconciler")

	defer s.Recover(sentry.CurrentHub())

	ingestionProcessor, err := reconciler.NewIngestionProcessor(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate ingestion processor", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(ingestionProcessor); err != nil {
		s.Logger().Error("failed to register ingestion processor", "error", err)
		os.Exit(1)
	}

	gRPCServer, err := reconciler.NewServer(ctx, s.Logger())
	if err != nil {
		s.Logger().Error("failed to instantiate gRPC server", "error", err)
		os.Exit(1)
	}

	if err := s.RegisterComponent(gRPCServer); err != nil {
		s.Logger().Error("failed to register gRPC server", "error", err)
		os.Exit(1)
	}

	//TODO: instantiate prometheus server

	if err := s.Run(); err != nil {
		s.Logger().Error("server failure", "serverName", s.Name(), "error", err)
		os.Exit(1)
	}
}
