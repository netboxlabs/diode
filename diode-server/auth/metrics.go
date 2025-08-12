package auth

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/telemetry/otel"
)

// Metrics defines the interface for recording auth-specific metrics
type Metrics interface {
	// SetServiceInfo sets the service information (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
	// RecordServiceStartupAttempt records a service startup attempt
	RecordServiceStartupAttempt(ctx context.Context, success bool)
}

// MetricRecorder is a wrapper around the telemetry.MetricRecorder
type MetricRecorder struct {
	mr                   telemetry.MetricRecorder
	contextAttributeKeys []string
}

// NewMetricRecorder creates a new MetricRecorder with optional context attribute keys
func NewMetricRecorder(meter metric.Meter, environment string, contextAttributeKeys ...string) (*MetricRecorder, error) {
	recorder, err := otel.NewMetricRecorder(meter, environment, "auth", nil)
	if err != nil {
		return nil, err
	}

	return &MetricRecorder{
		mr:                   recorder,
		contextAttributeKeys: contextAttributeKeys,
	}, nil
}

// SetServiceInfo sets the service information (called once at startup)
func (r *MetricRecorder) SetServiceInfo(ctx context.Context, version string) {
	r.mr.SetServiceInfo(ctx, version)
}

// RecordServiceStartupAttempt records a service startup attempt
func (r *MetricRecorder) RecordServiceStartupAttempt(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, "service.startup_attempt", 1, attrs...)
}
