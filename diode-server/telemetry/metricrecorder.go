package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// MetricRecorder defines the interface for recording metrics
type MetricRecorder interface {
	// RecordCounter increments a counter metric
	RecordCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue)
	// RecordHistogram records a value in a histogram metric
	RecordHistogram(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue)
	// SetGauge sets the value of a gauge metric
	SetGauge(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue)
	// RecordUpDownCounter modifies an up/down counter metric
	RecordUpDownCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue)
	// SetServiceInfo sets the service information metric (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
}
