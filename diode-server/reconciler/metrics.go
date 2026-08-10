package reconciler

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/telemetry/otel"
)

// ReconcilerMetricDefinitions defines the metrics specific to the reconciler service
var ReconcilerMetricDefinitions = map[string]otel.MetricDefinition{
	"handle.message": {
		Name:        "handle_message_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of messages handled",
		Attributes:  []string{"success"},
	},
	"ingestionlog.create": {
		Name:        "ingestion_log_create_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of ingestion logs created",
		Attributes:  []string{"success"},
	},
	"ingestionlog.requeue": {
		Name:        "ingestion_log_requeue_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of duplicate ingestion logs requeued for re-plan due to possible NetBox state drift",
		Attributes:  []string{},
	},
	"changeset.create": {
		Name:        "change_set_create_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of change sets created",
		Attributes:  []string{"success"},
	},
	"changeset.apply": {
		Name:        "change_set_apply_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of change sets applied",
		Attributes:  []string{"success"},
	},
	"change.create": {
		Name:        "change_create_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of changes created",
		Attributes:  []string{},
	},
	"change.apply": {
		Name:        "change_apply_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Total number of changes applied",
		Attributes:  []string{},
	},
	"graph.upsert.duration": {
		Name:        "graph_upsert_duration_seconds",
		Type:        otel.Histogram,
		Unit:        otel.Seconds,
		Description: "Duration of graph entity upsert operations",
		Attributes:  []string{"success", "node_type"},
		Boundaries:  []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	},
}

// Metrics defines the interface for recording reconciler-specific metrics
type Metrics interface {
	// SetServiceInfo sets the service information (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
	// RecordServiceStartupAttempt records a service startup attempt
	RecordServiceStartupAttempt(ctx context.Context, success bool)
	// RecordHandleMessage records the handling of a message
	RecordHandleMessage(ctx context.Context, success bool)
	// RecordIngestionLogCreate records the creation of an ingestion log
	RecordIngestionLogCreate(ctx context.Context, success bool)
	// RecordIngestionLogRequeue records the requeue of a duplicate ingestion log for re-plan
	RecordIngestionLogRequeue(ctx context.Context)
	// RecordChangeSetCreate records the creation of a change set
	RecordChangeSetCreate(ctx context.Context, success bool, changes int64)
	// RecordChangeSetApply records the application of a change set
	RecordChangeSetApply(ctx context.Context, success bool, changes int64)
	// RecordGraphUpsert records the duration of a graph entity upsert
	RecordGraphUpsert(ctx context.Context, success bool, nodeType string, duration float64)
}

// MetricRecorder is a wrapper around the telemetry.MetricRecorder
type MetricRecorder struct {
	mr                   telemetry.MetricRecorder
	contextAttributeKeys []string
}

// NewMetricRecorder creates a new MetricRecorder with optional context attribute keys
func NewMetricRecorder(meter metric.Meter, environment string, contextAttributeKeys ...string) (*MetricRecorder, error) {
	recorder, err := otel.NewMetricRecorder(meter, environment, "reconciler", ReconcilerMetricDefinitions)
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

// RecordHandleMessage records the handling of a message
func (r *MetricRecorder) RecordHandleMessage(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, "handle.message", 1, attrs...)
}

// RecordIngestionLogCreate records the creation of an ingestion log
func (r *MetricRecorder) RecordIngestionLogCreate(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, "ingestionlog.create", 1, attrs...)
}

// RecordIngestionLogRequeue records the requeue of a duplicate ingestion log for re-plan
func (r *MetricRecorder) RecordIngestionLogRequeue(ctx context.Context) {
	attrs := telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)

	r.mr.RecordCounter(ctx, "ingestionlog.requeue", 1, attrs...)
}

// RecordChangeSetCreate records the creation of a change set
func (r *MetricRecorder) RecordChangeSetCreate(ctx context.Context, success bool, changes int64) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	// Record the change set creation
	r.mr.RecordCounter(ctx, "changeset.create", 1, attrs...)

	// If successful, also record the number of changes created
	if success {
		changeAttrs := telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)
		r.mr.RecordCounter(ctx, "change.create", changes, changeAttrs...)
	}
}

// RecordChangeSetApply records the application of a change set
func (r *MetricRecorder) RecordChangeSetApply(ctx context.Context, success bool, changes int64) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	// Record the change set application
	r.mr.RecordCounter(ctx, "changeset.apply", 1, attrs...)

	// If successful, also record the number of changes applied
	if success {
		changeAttrs := telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)
		r.mr.RecordCounter(ctx, "change.apply", changes, changeAttrs...)
	}
}

// RecordGraphUpsert records the duration of a graph entity upsert
func (r *MetricRecorder) RecordGraphUpsert(ctx context.Context, success bool, nodeType string, duration float64) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
		attribute.String("node_type", nodeType),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordHistogram(ctx, "graph.upsert.duration", duration, attrs...)
}
