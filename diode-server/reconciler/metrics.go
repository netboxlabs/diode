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
}

// Metrics defines the interface for recording ingester-specific metrics
type Metrics interface {
	// SetServiceInfo sets the service information (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
	// RecordHandleMessage records the handling of a message
	RecordHandleMessage(ctx context.Context, success bool)
	// RecordIngestionLogCreate records the creation of an ingestion log
	RecordIngestionLogCreate(ctx context.Context, success bool)
	// RecordChangeSetCreate records the creation of a change set
	RecordChangeSetCreate(ctx context.Context, success bool, changes int64)
	// RecordChangeSetApply records the application of a change set
	RecordChangeSetApply(ctx context.Context, success bool, changes int64)
}

// MetricRecorder is a wrapper around the telemetry.MetricRecorder
type MetricRecorder struct {
	mr telemetry.MetricRecorder
}

// NewMetricRecorder creates a new MetricRecorder
func NewMetricRecorder(meter metric.Meter, environment string) (*MetricRecorder, error) {
	recorder, err := otel.NewMetricRecorder(meter, environment, "reconciler", ReconcilerMetricDefinitions)
	if err != nil {
		return nil, err
	}

	return &MetricRecorder{recorder}, nil
}

// SetServiceInfo sets the service information (called once at startup)
func (r *MetricRecorder) SetServiceInfo(ctx context.Context, version string) {
	r.mr.SetServiceInfo(ctx, version)
}

// RecordHandleMessage records the handling of a message
func (r *MetricRecorder) RecordHandleMessage(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx)...)

	r.mr.RecordCounter(ctx, "handle.message", 1, attrs...)
}

// RecordIngestionLogCreate records the creation of an ingestion log
func (r *MetricRecorder) RecordIngestionLogCreate(ctx context.Context, success bool) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx)...)

	r.mr.RecordCounter(ctx, "ingestionlog.create", 1, attrs...)
}

// RecordChangeSetCreate records the creation of a change set
func (r *MetricRecorder) RecordChangeSetCreate(ctx context.Context, success bool, changes int64) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx)...)

	// Record the change set creation
	r.mr.RecordCounter(ctx, "changeset.create", 1, attrs...)

	// If successful, also record the number of changes created
	if success {
		changeAttrs := telemetry.GetAttributesFromContext(ctx)
		r.mr.RecordCounter(ctx, "change.create", changes, changeAttrs...)
	}
}

// RecordChangeSetApply records the application of a change set
func (r *MetricRecorder) RecordChangeSetApply(ctx context.Context, success bool, changes int64) {
	attrs := append([]attribute.KeyValue{
		attribute.Bool("success", success),
	}, telemetry.GetAttributesFromContext(ctx)...)

	// Record the change set application
	r.mr.RecordCounter(ctx, "changeset.apply", 1, attrs...)

	// If successful, also record the number of changes applied
	if success {
		changeAttrs := telemetry.GetAttributesFromContext(ctx)
		r.mr.RecordCounter(ctx, "change.apply", changes, changeAttrs...)
	}
}
