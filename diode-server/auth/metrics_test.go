package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/telemetry"
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

			recorder, err := auth.NewMetricRecorder(meter, tt.environment, tt.contextAttributeKeys...)

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

	recorder, err := auth.NewMetricRecorder(meter, "test")
	require.NoError(t, err)
	require.NotNil(t, recorder)

	// Should not panic or error
	recorder.SetServiceInfo(ctx, "v1.0.0.abc123")
}

func TestMetricRecorder_ContextAttributeHandling(t *testing.T) {
	contextAttrKey := "foo"
	contextAttrVal := "bar"
	ctx := telemetry.ContextWithMetricAttributes(context.Background(), attribute.String(contextAttrKey, contextAttrVal))

	meter := noop.NewMeterProvider().Meter("test")
	recorder, err := auth.NewMetricRecorder(meter, "test", contextAttrKey)
	require.NoError(t, err)
	require.NotNil(t, recorder)

	// Test that context attributes are properly extracted
	contextAttrs := telemetry.GetAttributesFromContext(ctx, contextAttrKey)
	assert.Len(t, contextAttrs, 1)
	assert.Equal(t, contextAttrKey, string(contextAttrs[0].Key))
	assert.Equal(t, contextAttrVal, contextAttrs[0].Value.AsString())

	// Should not panic with context containing attributes
	recorder.SetServiceInfo(ctx, "v1.0.0.abc123")
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

			recorder, err := auth.NewMetricRecorder(meter, "test")
			require.NoError(t, err)
			require.NotNil(t, recorder)

			// Should not panic or error
			recorder.RecordServiceStartupAttempt(ctx, tt.success)
		})
	}
}
