package pprof

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // Register pprof handlers
	"time"
)

// Listen starts an HTTP server for pprof profiling endpoints on the specified address.
// It gracefully shuts down when the provided context is canceled.
//
// The server exposes standard pprof endpoints at:
//   - /debug/pprof/
//   - /debug/pprof/cmdline
//   - /debug/pprof/profile
//   - /debug/pprof/symbol
//   - /debug/pprof/trace
//
// NOTE: it is a blocking call; run it in a separate goroutine.
func Listen(ctx context.Context, logger *slog.Logger, addr string) {
	if addr == "" {
		logger.Error("pprof address cannot be empty")
		return
	}

	mux := http.NewServeMux()
	// The pprof handlers are automatically registered at /debug/pprof/ when we import _ "net/http/pprof"
	// We use DefaultServeMux which has the pprof handlers registered
	mux.Handle("/debug/", http.DefaultServeMux)

	server := &http.Server{
		Addr:              addr,
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Channel to signal server startup errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		logger.Info("starting pprof server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("pprof server error", "error", err)
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		logger.Info("shutting down pprof server", "reason", ctx.Err())

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("pprof server shutdown error", "error", err)
		}

		logger.Info("pprof server stopped")

	case err := <-errChan:
		logger.Error("pprof server encountered an error", "error", err)
	}
}
