package telemetry

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServePrometheusMetricsIfNecessary serves the prometheus metrics if the metrics exporter is set to prometheus
func ServePrometheusMetricsIfNecessary(cfg Config, log *slog.Logger) {
	if cfg.MetricsExporter == "prometheus" {
		go ServePrometheusMetrics(cfg.MetricsPort, log)
	}
}

// ServePrometheusMetrics serves the prometheus metrics on the given port
func ServePrometheusMetrics(port int, log *slog.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Error("failed to serve prometheus metrics", "error", err)
	}
}
