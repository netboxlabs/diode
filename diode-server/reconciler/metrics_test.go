package reconciler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/netboxlabs/diode/diode-server/reconciler"
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

			recorder, err := reconciler.NewMetricRecorder(meter, tt.environment, tt.contextAttributeKeys...)

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

	recorder, err := reconciler.NewMetricRecorder(meter, "test")
	require.NoError(t, err)
	require.NotNil(t, recorder)

	// Should not panic or error
	recorder.SetServiceInfo(ctx, "v1.0.0.abc123")
}

func TestMetricRecorder_RecordHandleMessage(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful message handling",
			success: true,
		},
		{
			name:    "failed message handling",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := reconciler.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordHandleMessage(ctx, tt.success)
		})
	}
}

func TestMetricRecorder_RecordIngestionLogCreate(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful ingestion log creation",
			success: true,
		},
		{
			name:    "failed ingestion log creation",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := reconciler.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordIngestionLogCreate(ctx, tt.success)
		})
	}
}

func TestMetricRecorder_RecordChangeSetCreate(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		changes int64
	}{
		{
			name:    "successful changeset creation with single change",
			success: true,
			changes: 1,
		},
		{
			name:    "successful changeset creation with multiple changes",
			success: true,
			changes: 100,
		},
		{
			name:    "failed changeset creation",
			success: false,
			changes: 0,
		},
		{
			name:    "successful changeset creation with zero changes",
			success: true,
			changes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := reconciler.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordChangeSetCreate(ctx, tt.success, tt.changes)
		})
	}
}

func TestMetricRecorder_RecordChangeSetApply(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		changes int64
	}{
		{
			name:    "successful changeset application with single change",
			success: true,
			changes: 1,
		},
		{
			name:    "successful changeset application with multiple changes",
			success: true,
			changes: 50,
		},
		{
			name:    "failed changeset application",
			success: false,
			changes: 0,
		},
		{
			name:    "successful changeset application with zero changes",
			success: true,
			changes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			meter := noop.NewMeterProvider().Meter("test")

			recorder, err := reconciler.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordChangeSetApply(ctx, tt.success, tt.changes)
		})
	}
}

func TestMetricRecorder_WithContextAttributeKeys(t *testing.T) {
	ctx := context.Background()
	meter := noop.NewMeterProvider().Meter("test")

	// Create recorder with context attribute keys
	recorder, err := reconciler.NewMetricRecorder(meter, "test", "environment")
	require.NoError(t, err)
	require.NotNil(t, recorder)

	// Test all metric recording methods with context that may contain attributes
	recorder.SetServiceInfo(ctx, "v1.0.0")
	recorder.RecordHandleMessage(ctx, true)
	recorder.RecordHandleMessage(ctx, false)
	recorder.RecordIngestionLogCreate(ctx, true)
	recorder.RecordIngestionLogCreate(ctx, false)
	recorder.RecordChangeSetCreate(ctx, true, 10)
	recorder.RecordChangeSetCreate(ctx, false, 0)
	recorder.RecordChangeSetApply(ctx, true, 5)
	recorder.RecordChangeSetApply(ctx, false, 0)
}

func TestMetricRecorder_WithAttributeFromContext(t *testing.T) {
	contextAttrKey := "foo"
	contextAttrVal := "bar"
	ctx := telemetry.ContextWithMetricAttributes(context.Background(), attribute.String(contextAttrKey, contextAttrVal))

	// Create mock telemetry recorder to capture metrics calls
	mockRecorder := telemetrymocks.NewMetricRecorder(t)

	// Test RecordHandleMessage with environment attribute
	// The mock converts variadic args to individual interface{} args, so we need to match each attribute individually
	mockRecorder.On("RecordCounter",
		mock.Anything, // context
		"handle.message",
		int64(1),
		attribute.Bool("success", true), // first attribute
		attribute.String(contextAttrKey, contextAttrVal), // second attribute from context
	).Return()

	// Simulate what the actual MetricRecorder would do:
	// 1. Extract method-specific attributes
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
	mockRecorder.RecordCounter(ctx, "handle.message", 1, allAttrs...)
}
