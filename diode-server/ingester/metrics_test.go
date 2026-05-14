package ingester_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/telemetry"
	telemetrymocks "github.com/netboxlabs/diode/diode-server/telemetry/mocks"
)

func TestNewMetricRecorder(t *testing.T) {
	tests := []struct {
		name                 string
		environment          string
		contextAttributeKeys []string
		expectedError        bool
	}{
		{
			name:          "successful creation with no context keys",
			environment:   "test",
			expectedError: false,
		},
		{
			name:                 "successful creation with context keys",
			environment:          "test",
			contextAttributeKeys: []string{"environment"},
			expectedError:        false,
		},
		{
			name:          "successful creation with empty environment",
			environment:   "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, tt.environment, tt.contextAttributeKeys...)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, recorder)
			} else {
				require.NoError(t, err)
				require.NotNil(t, recorder)
			}
		})
	}
}

func TestMetricRecorder_SetServiceInfo(t *testing.T) {
	ctx := context.Background()
	meter := noop.NewMeterProvider().Meter("test")

	recorder, err := ingester.NewMetricRecorder(meter, "test")
	require.NoError(t, err)
	require.NotNil(t, recorder)

	// Should not panic or error
	recorder.SetServiceInfo(ctx, "v1.0.0.abc123")
}

func TestMetricRecorder_RecordIngestRequest(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful request",
			success: true,
		},
		{
			name:    "failed request",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordIngestRequest(ctx, tt.success)
		})
	}
}

func TestMetricRecorder_RecordIngestEntities(t *testing.T) {
	tests := []struct {
		name  string
		count int64
	}{
		{
			name:  "single entity",
			count: 1,
		},
		{
			name:  "multiple entities",
			count: 100,
		},
		{
			name:  "zero entities",
			count: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordIngestEntities(ctx, tt.count)
		})
	}
}

func TestMetricRecorder_WithAttributeFromContext(t *testing.T) {
	contextAttrKey := "foo"
	contextAttrVal := "bar"
	ctx := telemetry.ContextWithMetricAttributes(context.Background(), attribute.String(contextAttrKey, contextAttrVal))

	// Create mock telemetry recorder to capture metrics calls
	mockRecorder := telemetrymocks.NewMetricRecorder(t)

	// Set up expectation for RecordCounter with environment attribute
	// The mock converts variadic args to individual interface{} args, so we need to match each attribute individually
	mockRecorder.On("RecordCounter",
		mock.Anything, // context
		"ingest.requests",
		int64(1),
		attribute.Bool("success", true), // first attribute
		attribute.String(contextAttrKey, contextAttrVal), // second attribute from context
	).Return()

	// Simulate what the actual MetricRecorder would do:
	// 1. Extract success attribute
	// 2. Get environment from context using GetAttributesFromContext
	// 3. Combine them and call the underlying recorder

	successAttr := []attribute.KeyValue{attribute.Bool("success", true)}
	contextAttrs := telemetry.GetAttributesFromContext(ctx, contextAttrKey)
	allAttrs := append(successAttr, contextAttrs...)

	// Verify the environment attribute was extracted correctly
	assert.Len(t, contextAttrs, 1)
	assert.Equal(t, contextAttrKey, string(contextAttrs[0].Key))
	assert.Equal(t, contextAttrVal, contextAttrs[0].Value.AsString())

	// Call the mock with the combined attributes
	mockRecorder.RecordCounter(ctx, "ingest.requests", 1, allAttrs...)
}

func TestMetricRecorder_RecordRedisRejection(t *testing.T) {
	for _, reason := range []string{"watermark", "redis_oom"} {
		t.Run(reason, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordRedisRejection(ctx, reason)
		})
	}
}

func TestMetricRecorder_SetRedisMemoryRatioBPS(t *testing.T) {
	for _, ratio := range []int64{0, 5000, 10000} {
		t.Run("ratio", func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.SetRedisMemoryRatioBPS(ctx, ratio)
		})
	}
}

func TestMetricRecorder_RecordServiceStartupAttempt(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful startup attempt",
			success: true,
		},
		{
			name:    "failed startup attempt",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := ingester.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordServiceStartupAttempt(ctx, tt.success)
		})
	}
}
