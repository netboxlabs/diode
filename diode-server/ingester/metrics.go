package ingester

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	metricIngestRequest = "netboxlabs.com/diode/ingester/ingest_request_count"
	metricIngestEntity  = "netboxlabs.com/diode/ingester/ingest_entity_count"
)

// Metrics is a struct that contains the metrics for the ingester.
type Metrics struct {
	ingestRequest metric.Int64Counter
	ingestEntity  metric.Int64Counter
}

// NewMetrics creates a new Metrics instance.
func NewMetrics(meter metric.Meter) (*Metrics, error) {
	ingestRequest, err := meter.Int64Counter(metricIngestRequest,
		metric.WithDescription("Total number of ingest requests handled"))
	if err != nil {
		return nil, fmt.Errorf("failed to create total counter: %v", err)
	}

	ingestEntity, err := meter.Int64Counter(metricIngestEntity,
		metric.WithDescription("Total number of entities ingested"))
	if err != nil {
		return nil, fmt.Errorf("failed to create total counter: %v", err)
	}

	return &Metrics{
		ingestRequest: ingestRequest,
		ingestEntity:  ingestEntity,
	}, nil
}

// RecordIngestRequest records an ingest request.
func (m *Metrics) RecordIngestRequest(ctx context.Context, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool(telemetry.AttributeSuccess, success),
	}
	m.ingestRequest.Add(ctx, 1, telemetry.GatherOptions(ctx, attrs)...)
}

// RecordIngestEntities records the number of entities ingested.
func (m *Metrics) RecordIngestEntities(ctx context.Context, count int64) {
	attrs := []attribute.KeyValue{}
	m.ingestEntity.Add(ctx, count, telemetry.GatherOptions(ctx, attrs)...)
}
