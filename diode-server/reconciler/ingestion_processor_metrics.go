package reconciler

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	// Metric names
	metricHandleMessage      = "diode.reconciler.ingestion_processor.handle_message"
	metricIngestionLogCreate = "diode.reconciler.ingestion_processor.ingestion_log_create"
	metricChangeSetCreate    = "diode.reconciler.ingestion_processor.change_set_create"
	metricChangeSetApply     = "diode.reconciler.ingestion_processor.change_set_apply"
	metricChangeCreate       = "diode.reconciler.ingestion_processor.change_create"
	metricChangeApply        = "diode.reconciler.ingestion_processor.change_apply"
)

// IngestionProcessorMetrics is a struct that contains the metrics for the ingestion processor.
type IngestionProcessorMetrics struct {
	// Metrics
	handleMessage      metric.Int64Counter
	ingestionLogCreate metric.Int64Counter
	changeSetCreate    metric.Int64Counter
	changeCreate       metric.Int64Counter
	changeSetApply     metric.Int64Counter
	changeApply        metric.Int64Counter
}

// NewIngestionProcessorMetrics creates a new IngestionProcessorMetrics instance.
func NewIngestionProcessorMetrics(meter metric.Meter) (*IngestionProcessorMetrics, error) {
	handleMessage, err := meter.Int64Counter(metricHandleMessage, metric.WithDescription("Number of messages handled"))
	if err != nil {
		return nil, fmt.Errorf("failed to create handle message counter: %v", err)
	}

	ingestionLogCreate, err := meter.Int64Counter(metricIngestionLogCreate, metric.WithDescription("Number of ingestion logs created"))
	if err != nil {
		return nil, fmt.Errorf("failed to create ingestion log create counter: %v", err)
	}

	changeSetCreate, err := meter.Int64Counter(metricChangeSetCreate, metric.WithDescription("Number of change sets created"))
	if err != nil {
		return nil, fmt.Errorf("failed to create change set create counter: %v", err)
	}

	changeSetApply, err := meter.Int64Counter(metricChangeSetApply, metric.WithDescription("Number of change sets applied"))
	if err != nil {
		return nil, fmt.Errorf("failed to create change set apply counter: %v", err)
	}

	changeCreate, err := meter.Int64Counter(metricChangeCreate, metric.WithDescription("Number of changes created"))
	if err != nil {
		return nil, fmt.Errorf("failed to create change create counter: %v", err)
	}

	changeApply, err := meter.Int64Counter(metricChangeApply, metric.WithDescription("Number of changes applied"))
	if err != nil {
		return nil, fmt.Errorf("failed to create change apply counter: %v", err)
	}

	return &IngestionProcessorMetrics{
		handleMessage:      handleMessage,
		ingestionLogCreate: ingestionLogCreate,
		changeSetCreate:    changeSetCreate,
		changeCreate:       changeCreate,
		changeSetApply:     changeSetApply,
		changeApply:        changeApply,
	}, nil
}

// RecordHandleMessage records a message being handled.
func (m *IngestionProcessorMetrics) RecordHandleMessage(ctx context.Context, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool(telemetry.AttributeSuccess, success),
	}
	m.handleMessage.Add(ctx, 1, telemetry.GatherOptions(ctx, attrs)...)
}

// RecordIngestionLogCreate records an ingestion log being created.
func (m *IngestionProcessorMetrics) RecordIngestionLogCreate(ctx context.Context, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool(telemetry.AttributeSuccess, success),
	}
	m.ingestionLogCreate.Add(ctx, 1, telemetry.GatherOptions(ctx, attrs)...)
}

// RecordChangeSetCreate records a change set being created.
func (m *IngestionProcessorMetrics) RecordChangeSetCreate(ctx context.Context, success bool, changes int64) {
	attrs := []attribute.KeyValue{
		attribute.Bool(telemetry.AttributeSuccess, success),
	}
	options := telemetry.GatherOptions(ctx, attrs)
	m.changeSetCreate.Add(ctx, 1, options...)
	if success {
		m.changeCreate.Add(ctx, changes, options...)
	}
}

// RecordChangeSetApply records a change set being applied.
func (m *IngestionProcessorMetrics) RecordChangeSetApply(ctx context.Context, success bool, changes int64) {
	attrs := []attribute.KeyValue{
		attribute.Bool(telemetry.AttributeSuccess, success),
	}
	options := telemetry.GatherOptions(ctx, attrs)
	m.changeSetApply.Add(ctx, 1, options...)
	if success {
		m.changeApply.Add(ctx, changes, options...)
	}
}
