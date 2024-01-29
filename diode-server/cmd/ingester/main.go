package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/netboxlabs/diode-internal/diode-server/ingester"
	"github.com/netboxlabs/diode-internal/diode-server/server"
)

func main() {
	// TODO(mfiedorowicz): make logger configurable (handler, level, etc.)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	ctx := context.Background()
	s := server.New(ctx, "diode-ingester", logger)

	ingesterComponent := ingester.New(logger)

	if err := s.RegisterComponent(ingesterComponent); err != nil {
		log.Fatalf("failed to register ingerster component: %v", err)
	}

	if err := s.Run(); err != nil {
		log.Fatalf("server %s failure: %v", s.Name(), err)
	}
}
