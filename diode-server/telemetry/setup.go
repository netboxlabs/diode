package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Setup initializes OpenTelemetry with the provided configuration
// Returns a shutdown function to clean up resources that should be called
// when the application is shutting down.
// The OTLP exporters can be further configured using the environment variables
// specified in the OTLP documentation.
func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	enableTraces := true
	var traceExporter sdktrace.SpanExporter
	switch cfg.TracesExporter {
	case "console":
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create console trace exporter: %w", err)
		}
	case "otlp":
		traceExporter, err = otlptracegrpc.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	case "none":
		enableTraces = false
	default:
		return nil, fmt.Errorf("unsupported traces exporter type: %s", cfg.TracesExporter)
	}

	enableMetrics := true
	var metricReader sdkmetric.Reader
	switch cfg.MetricsExporter {
	case "console":
		metricExporter, err := stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create console metric exporter: %w", err)
		}
		metricReader = sdkmetric.NewPeriodicReader(metricExporter)
	case "otlp":
		metricExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}
		metricReader = sdkmetric.NewPeriodicReader(metricExporter)
	case "prometheus":
		metricReader, err = prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create prometheus metric exporter: %w", err)
		}
	case "none":
		enableMetrics = false
	default:
		return nil, fmt.Errorf("unsupported metrics exporter type: %s", cfg.MetricsExporter)
	}

	var tracerProvider *sdktrace.TracerProvider
	var meterProvider *sdkmetric.MeterProvider

	if enableTraces {
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExporter),
		)
		otel.SetTracerProvider(tracerProvider)
	}

	if enableMetrics {
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(metricReader),
		)
		otel.SetMeterProvider(meterProvider)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown = func(ctx context.Context) error {
		var shutdownErr error
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("failed to shutdown meter provider: %w", err)
			}
		}
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("failed to shutdown tracer provider: %w", err)
			}
		}
		return shutdownErr
	}

	return shutdown, nil
}
