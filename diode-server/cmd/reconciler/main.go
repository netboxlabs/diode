package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/netboxlabs/diode-internal/diode-server/reconciler"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	// TODO(mfiedorowicz): make logger configurable (handler, level, etc.)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	ctx := context.Background()
	s := server.New(ctx, "diode-distributor", logger)

	// instantiate a reconciler service
	// with config instantiated here and passed to New()
	reconcilerSvc, err := reconciler.New(logger)
	if err != nil {
		log.Fatalf("failed to instantiate reconciler service: %v", err)
	}

	if err := s.RegisterService(reconcilerSvc); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// instantiate a prom service for /metrics
	// prometheusSvc, err := prometheus.New()

	if err := s.Run(); err != nil {
		log.Fatalf("server %s failure: %v", s.Name(), err)
	}
}
