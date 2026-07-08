package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/netboxlabs/diode/diode-server"

// Span names align with diode-pro/docs/observability/tracing.md.
const (
	SpanIngestionHandleStreamMessage = "ingestion.handle_stream_message"
	SpanIngestionCreateIngestionLogs = "ingestion.create_ingestion_logs"
	SpanRateLimiterWait              = "rate_limiter.wait"
)

var tracer = otel.Tracer(tracerName)

// StartSpan begins a custom span with optional attributes.
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// End finishes a span, recording error status when err is non-nil.
func End(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
	span.End()
}
