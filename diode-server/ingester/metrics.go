package ingester

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/telemetry/otel"
)

// IngesterMetricDefinitions defines the metrics specific to the ingester service
var IngesterMetricDefinitions = map[string]otel.MetricDefinition{
	"ingest.requests": {
		Name:        "ingest_requests_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of ingest requests handled",
		Attributes:  []string{"success"},
	},
	"ingest.entities": {
		Name:        "ingest_entities_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of entities ingested",
		Attributes:  []string{},
	},
}

// Metrics defines the interface for recording ingester-specific metrics
type Metrics interface {
	// SetServiceInfo sets the service information (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
	// RecordServiceStartupAttempt records a service startup attempt
	RecordServiceStartupAttempt(ctx context.Context, success bool)
	// RecordIngestRequest records an ingest request
	RecordIngestRequest(ctx context.Context, success bool)
	// RecordIngestEntities records the number of entities ingested
	RecordIngestEntities(ctx context.Context, count int64)
}

// MetricRecorder is a wrapper around the telemetry.MetricRecorder
type MetricRecorder struct {
	mr                   telemetry.MetricRecorder
	contextAttributeKeys []string
}

// NewMetricRecorder creates a new MetricRecorder with optional context attribute keys
func NewMetricRecorder(meter metric.Meter, environment string, contextAttributeKeys ...string) (*MetricRecorder, error) {
	recorder, err := otel.NewMetricRecorder(meter, environment, "ingester", IngesterMetricDefinitions)
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
	r.mr.RecordServiceStartupAttempt(ctx, success)
}

// RecordIngestRequest records an ingest request.
func (r *MetricRecorder) RecordIngestRequest(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, "ingest.requests", 1, attrs...)
}

// RecordIngestEntities records the number of entities ingested.
func (r *MetricRecorder) RecordIngestEntities(ctx context.Context, count int64) {
	attrs := telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)
	r.mr.RecordCounter(ctx, "ingest.entities", count, attrs...)
}
