package auth

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/netboxlabs/diode/diode-server/telemetry"
	"github.com/netboxlabs/diode/diode-server/telemetry/otel"
)

// Metric keys. The recorder silently drops anything not declared in
// AuthMetricDefinitions, so record calls and definitions share these constants rather
// than repeating a string literal in both places.
const (
	metricTokenCacheOutcome   = "token.cache_outcome"
	metricUpstreamTokenReqDur = "token.upstream_request_duration"
)

// AuthMetricDefinitions defines the metrics specific to the auth service.
//
// A metric that is not declared here is silently dropped by the recorder, so anything
// recorded below must have an entry.
var AuthMetricDefinitions = map[string]otel.MetricDefinition{
	metricTokenCacheOutcome: {
		Name:        "token_cache_outcome_total",
		Type:        otel.Counter,
		Unit:        otel.Dimensionless,
		Description: "Token requests by how they interacted with the token cache",
		Attributes:  []string{"outcome"},
	},
	metricUpstreamTokenReqDur: {
		Name:        "token_upstream_request_duration_seconds",
		Type:        otel.Histogram,
		Unit:        otel.Seconds,
		Description: "Duration of token requests forwarded to the oauth2 server",
		Attributes:  []string{"outcome"},
	},
}

// Metrics defines the interface for recording auth-specific metrics
type Metrics interface {
	// SetServiceInfo sets the service information (called once at startup)
	SetServiceInfo(ctx context.Context, version string)
	// RecordServiceStartupAttempt records a service startup attempt
	RecordServiceStartupAttempt(ctx context.Context, success bool)
	// RecordTokenCacheOutcome records how a token request interacted with the token cache
	RecordTokenCacheOutcome(ctx context.Context, outcome string)
	// RecordUpstreamTokenRequest records the duration and outcome of an upstream token request
	RecordUpstreamTokenRequest(ctx context.Context, duration time.Duration, outcome string)
}

// MetricRecorder is a wrapper around the telemetry.MetricRecorder
type MetricRecorder struct {
	mr                   telemetry.MetricRecorder
	contextAttributeKeys []string
}

// NewMetricRecorder creates a new MetricRecorder with optional context attribute keys
func NewMetricRecorder(meter metric.Meter, environment string, contextAttributeKeys ...string) (*MetricRecorder, error) {
	recorder, err := otel.NewMetricRecorder(meter, environment, "auth", AuthMetricDefinitions)
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

// RecordTokenCacheOutcome records how a token request interacted with the token cache
func (r *MetricRecorder) RecordTokenCacheOutcome(ctx context.Context, outcome string) {
	attrs := append([]attribute.KeyValue{
		attribute.String("outcome", outcome),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordCounter(ctx, metricTokenCacheOutcome, 1, attrs...)
}

// RecordUpstreamTokenRequest records the duration and outcome of an upstream token request
func (r *MetricRecorder) RecordUpstreamTokenRequest(ctx context.Context, duration time.Duration, outcome string) {
	attrs := append([]attribute.KeyValue{
		attribute.String("outcome", outcome),
	}, telemetry.GetAttributesFromContext(ctx, r.contextAttributeKeys...)...)

	r.mr.RecordHistogram(ctx, metricUpstreamTokenReqDur, duration.Seconds(), attrs...)
}
