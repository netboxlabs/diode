package telemetry

import (
	"context"
	"strings"

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

// ExtractPathFromPattern removes the HTTP method from a pattern like "GET /foo" -> "/foo"
func ExtractPathFromPattern(pattern string) string {
	// Split by space and take the second part (the path)
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	// If no space found, return as-is (shouldn't happen with ServeMux patterns)
	return pattern
}
