package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// ContextWithMetricAttributes adds metric attributes to the context.
func ContextWithMetricAttributes(ctx context.Context, attrs ...attribute.KeyValue) context.Context {
	existing := MetricAttributesFromContext(ctx)
	return context.WithValue(ctx, metricAttributesKey{}, append(existing, attrs...))
}

// MetricAttributesFromContext returns the metric attributes from the context.
func MetricAttributesFromContext(ctx context.Context) []attribute.KeyValue {
	if attrs, ok := ctx.Value(metricAttributesKey{}).([]attribute.KeyValue); ok {
		return attrs
	}
	return []attribute.KeyValue{}
}

type metricAttributesKey struct{}

// GatherOptions collects metric attributes from the context and appends them to the options and attributes given
func GatherOptions(ctx context.Context, attrs []attribute.KeyValue, options ...metric.AddOption) []metric.AddOption {
	return append(options,
		metric.WithAttributes(append(attrs, MetricAttributesFromContext(ctx)...)...))
}
