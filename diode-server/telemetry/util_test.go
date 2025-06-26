package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func TestContextWithMetricAttributes(t *testing.T) {
	tests := []struct {
		name     string
		existing []attribute.KeyValue
		new      []attribute.KeyValue
		expected []attribute.KeyValue
	}{
		{
			name:     "empty context with new attributes",
			existing: nil,
			new: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.String("key2", "value2"),
			},
			expected: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.String("key2", "value2"),
			},
		},
		{
			name: "existing context with additional attributes",
			existing: []attribute.KeyValue{
				attribute.String("existing", "value"),
			},
			new: []attribute.KeyValue{
				attribute.String("new", "value"),
			},
			expected: []attribute.KeyValue{
				attribute.String("existing", "value"),
				attribute.String("new", "value"),
			},
		},
		{
			name: "duplicate keys",
			existing: []attribute.KeyValue{
				attribute.String("key", "existing"),
			},
			new: []attribute.KeyValue{
				attribute.String("key", "new"),
			},
			expected: []attribute.KeyValue{
				attribute.String("key", "existing"),
				attribute.String("key", "new"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add existing attributes if any
			if tt.existing != nil {
				ctx = ContextWithMetricAttributes(ctx, tt.existing...)
			}

			// Add new attributes
			ctx = ContextWithMetricAttributes(ctx, tt.new...)

			// Verify the result
			attrs := MetricAttributesFromContext(ctx)
			assert.Equal(t, tt.expected, attrs)
		})
	}
}

func TestMetricAttributesFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected []attribute.KeyValue
	}{
		{
			name: "empty context",
			setup: func() context.Context {
				return context.Background()
			},
			expected: []attribute.KeyValue{},
		},
		{
			name: "context with attributes",
			setup: func() context.Context {
				return ContextWithMetricAttributes(context.Background(),
					attribute.String("service", "test"),
					attribute.Int64("count", 42),
				)
			},
			expected: []attribute.KeyValue{
				attribute.String("service", "test"),
				attribute.Int64("count", 42),
			},
		},
		{
			name: "context with wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), metricAttributesKey{}, "not attributes")
			},
			expected: []attribute.KeyValue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			attrs := MetricAttributesFromContext(ctx)
			assert.Equal(t, tt.expected, attrs)
		})
	}
}

func TestGatherOptions(t *testing.T) {
	ctx := ContextWithMetricAttributes(context.Background(),
		attribute.String("ctx_key", "ctx_value"),
	)

	attrs := []attribute.KeyValue{
		attribute.String("param_key", "param_value"),
	}

	existingOption := metric.WithAttributes(attribute.String("existing", "value"))
	options := GatherOptions(ctx, attrs, existingOption)

	require.Len(t, options, 2) // existing option + WithAttributes option

	// Verify the existing option is preserved
	assert.Equal(t, existingOption, options[0])

	// The second option should be a WithAttributes option combining both sets of attributes
	// We can't directly inspect the WithAttributes option, but we can verify it exists
	assert.NotNil(t, options[1])
}

func TestGetAttributesFromContext(t *testing.T) {
	ctx := ContextWithMetricAttributes(context.Background(),
		attribute.String("service", "test-service"),
		attribute.String("version", "1.0.0"),
		attribute.String("environment", "test"),
		attribute.Int64("port", 8080),
	)

	tests := []struct {
		name     string
		keys     []string
		expected []attribute.KeyValue
	}{
		{
			name: "no keys - return all attributes",
			keys: nil,
			expected: []attribute.KeyValue{
				attribute.String("service", "test-service"),
				attribute.String("version", "1.0.0"),
				attribute.String("environment", "test"),
				attribute.Int64("port", 8080),
			},
		},
		{
			name: "empty keys - return all attributes",
			keys: []string{},
			expected: []attribute.KeyValue{
				attribute.String("service", "test-service"),
				attribute.String("version", "1.0.0"),
				attribute.String("environment", "test"),
				attribute.Int64("port", 8080),
			},
		},
		{
			name: "specific keys - return matching attributes",
			keys: []string{"service", "version"},
			expected: []attribute.KeyValue{
				attribute.String("service", "test-service"),
				attribute.String("version", "1.0.0"),
			},
		},
		{
			name:     "non-existent keys - return empty",
			keys:     []string{"nonexistent"},
			expected: []attribute.KeyValue{},
		},
		{
			name: "mix of existing and non-existent keys",
			keys: []string{"service", "nonexistent", "environment"},
			expected: []attribute.KeyValue{
				attribute.String("service", "test-service"),
				attribute.String("environment", "test"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := GetAttributesFromContext(ctx, tt.keys...)
			assert.Equal(t, tt.expected, attrs)
		})
	}
}

func TestGetAttributesFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()

	// Test with no keys
	attrs := GetAttributesFromContext(ctx)
	assert.Equal(t, []attribute.KeyValue{}, attrs)

	// Test with specific keys
	attrs = GetAttributesFromContext(ctx, "service", "version")
	assert.Equal(t, []attribute.KeyValue{}, attrs)
}
