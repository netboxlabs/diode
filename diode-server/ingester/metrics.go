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
	"ingest.redis_rejections": {
		Name:        "ingest_redis_rejections_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Ingest requests rejected due to Redis memory pressure",
		Attributes:  []string{"reason"},
	},
	"redis.memory_used_ratio_bps": {
		Name:        "redis_memory_used_ratio_bps",
		Type:        otel.Gauge,
		Unit:        otel.Dimensionless,
		Description: "Redis used_memory as basis points (0..10000) of maxmemory; refresh cadence is bounded by REDIS_MEMORY_CHECK_INTERVAL",
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
	// RecordRedisRejection increments the rejection counter with a reason
	// label. Reason is one of "watermark" (soft-shed before Redis is full)
	// or "redis_oom" (Redis itself returned OOM on XAdd).
	RecordRedisRejection(ctx context.Context, reason string)
	// SetRedisMemoryRatioBPS reports Redis used_memory as basis points of
	// maxmemory (e.g. 7050 = 70.50%). Called on each successful poll.
	SetRedisMemoryRatioBPS(ctx context.Context, ratioBPS int64)
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

// RecordRedisRejection records a rejection due to Redis memory pressure.
func (r *MetricRecorder) RecordRedisRejection(ctx context.Context, reason string) {
	attrs := append([]attribute.KeyValue{
		attribute.String("reason", reason),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, "ingest.redis_rejections", 1, attrs...)
}

// SetRedisMemoryRatioBPS records the current Redis used/max memory ratio in
// basis points (0..10000).
func (r *MetricRecorder) SetRedisMemoryRatioBPS(ctx context.Context, ratioBPS int64) {
	attrs := telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)
	r.mr.SetGauge(ctx, "redis.memory_used_ratio_bps", ratioBPS, attrs...)
}
