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

// GetAttributesFromContext extracts attributes from context. If no keys are provided,
// returns all attributes. If keys are provided, returns only attributes matching those keys.
func GetAttributesFromContext(ctx context.Context, keys ...string) []attribute.KeyValue {
	allAttrs := MetricAttributesFromContext(ctx)

	// If no keys specified, return all attributes
	if len(keys) == 0 {
		return allAttrs
	}

	// If keys specified, filter attributes
	ctxAttrs := make([]attribute.KeyValue, 0)
	keySet := make(map[string]bool)
	for _, key := range keys {
		keySet[key] = true
	}

	for _, attr := range allAttrs {
		if keySet[string(attr.Key)] {
			ctxAttrs = append(ctxAttrs, attr)
		}
	}

	return ctxAttrs
}
