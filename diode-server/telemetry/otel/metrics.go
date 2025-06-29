package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	// MetricsNamespace is the namespace for all Diode metrics
	MetricsNamespace = "diode"

	// AttrServiceName is the attribute key for a service name
	AttrServiceName = string(semconv.ServiceNameKey)

	// AttrServiceVersion is the attribute key for a service version
	AttrServiceVersion = string(semconv.ServiceVersionKey)

	// AttrDeploymentEnvironment is the attribute key for the deployment environment (e.g., production, staging)
	AttrDeploymentEnvironment = "deployment.environment"
)

// MetricRecorder provides a standardized way to create and use OpenTelemetry metrics
type MetricRecorder struct {
	environment     string
	metricNamespace string
	meter           metric.Meter
	serviceName     string
	counters        map[string]metric.Int64Counter
	histograms      map[string]metric.Float64Histogram
	gauges          map[string]metric.Int64Gauge
	updowns         map[string]metric.Int64UpDownCounter
}

// NewMetricRecorder creates a new OpenTelemetry metrics recorder that combines core service metrics
// with service-specific metrics. It initializes and holds all metric instruments (counters,
// histograms, gauges, updown counters) for efficient recording of metric values.
func NewMetricRecorder(meter metric.Meter, environment, serviceName string, serviceMetrics map[string]MetricDefinition) (*MetricRecorder, error) {
	m := &MetricRecorder{
		environment:     environment,
		metricNamespace: MetricsNamespace,
		meter:           meter,
		serviceName:     serviceName,
		counters:        make(map[string]metric.Int64Counter),
		histograms:      make(map[string]metric.Float64Histogram),
		gauges:          make(map[string]metric.Int64Gauge),
		updowns:         make(map[string]metric.Int64UpDownCounter),
	}

	// Merge core metrics with service-specific metrics
	allMetrics := make(map[string]MetricDefinition)
	for k, v := range CoreServiceMetrics {
		allMetrics[k] = v
	}
	for k, v := range serviceMetrics {
		allMetrics[k] = v
	}

	if err := m.createMetrics(allMetrics); err != nil {
		return nil, err
	}

	return m, nil
}

// createMetrics creates all the metrics based on their definitions
func (m *MetricRecorder) createMetrics(definitions map[string]MetricDefinition) error {
	for key, def := range definitions {
		metricName := m.buildMetricName(def.Name)

		switch def.Type {
		case Counter:
			counter, err := m.meter.Int64Counter(
				metricName,
				metric.WithDescription(def.Description),
				metric.WithUnit(string(def.Unit)),
			)
			if err != nil {
				return fmt.Errorf("failed to create counter %s: %w", metricName, err)
			}
			m.counters[key] = counter

		case UpDown:
			updown, err := m.meter.Int64UpDownCounter(
				metricName,
				metric.WithDescription(def.Description),
				metric.WithUnit(string(def.Unit)),
			)
			if err != nil {
				return fmt.Errorf("failed to create updown counter %s: %w", metricName, err)
			}
			m.updowns[key] = updown

		case Histogram:
			histogram, err := m.meter.Float64Histogram(
				metricName,
				metric.WithDescription(def.Description),
				metric.WithUnit(string(def.Unit)),
			)
			if err != nil {
				return fmt.Errorf("failed to create histogram %s: %w", metricName, err)
			}
			m.histograms[key] = histogram

		case Gauge:
			gauge, err := m.meter.Int64Gauge(
				metricName,
				metric.WithDescription(def.Description),
				metric.WithUnit(string(def.Unit)),
			)
			if err != nil {
				return fmt.Errorf("failed to create gauge %s: %w", metricName, err)
			}
			m.gauges[key] = gauge

		default:
			return fmt.Errorf("unsupported metric type: %v", def.Type)
		}
	}

	return nil
}

// buildMetricName constructs the full metric name with namespace and service
func (m *MetricRecorder) buildMetricName(name string) string {
	return fmt.Sprintf("%s_%s_%s", m.metricNamespace, m.serviceName, name)
}

// RecordCounter increments a counter metric
func (m *MetricRecorder) RecordCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if counter, exists := m.counters[name]; exists {
		counter.Add(ctx, value, metric.WithAttributes(attrs...))
	}
}

// RecordHistogram records a value in a histogram metric
func (m *MetricRecorder) RecordHistogram(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	if histogram, exists := m.histograms[name]; exists {
		histogram.Record(ctx, value, metric.WithAttributes(attrs...))
	}
}

// SetGauge sets the value of a gauge metric
func (m *MetricRecorder) SetGauge(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.Record(ctx, value, metric.WithAttributes(attrs...))
	}
}

// RecordUpDownCounter modifies an up/down counter metric
func (m *MetricRecorder) RecordUpDownCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if updown, exists := m.updowns[name]; exists {
		updown.Add(ctx, value, metric.WithAttributes(attrs...))
	}
}

// SetServiceInfo sets the service information metric (called once at startup)
func (m *MetricRecorder) SetServiceInfo(ctx context.Context, version string) {
	m.SetGauge(ctx, "service.info", 1,
		attribute.String(AttrServiceName, m.serviceName),
		attribute.String(AttrServiceVersion, version),
		attribute.String(AttrDeploymentEnvironment, m.environment), // or other environment
	)
}

// RecordServiceStartupAttempt increments the service startup attempt counter
func (m *MetricRecorder) RecordServiceStartupAttempt(ctx context.Context, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool("success", success),
	}
	m.RecordCounter(ctx, "service.startup_attempt", 1, attrs...)
}

// MetricDefinition represents a standardized metric definition
type MetricDefinition struct {
	Name        string
	Type        MetricType
	Unit        Unit
	Description string
	Attributes  []string
}

// MetricType represents the type of metric
type MetricType int

const (
	// Counter is a metric that counts occurrences
	Counter MetricType = iota
	// Gauge is a metric that represents a value at a point in time
	Gauge
	// Histogram is a metric that records distributions of values
	Histogram
	// UpDown is a metric that can increase or decrease
	UpDown
)

// String returns the string representation of the metric type
func (mt MetricType) String() string {
	switch mt {
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	case UpDown:
		return "updown"
	}
	return "unknown"
}

// Unit represents the unit of measurement for a metric
type Unit string

// Units for common metric types
const (
	// Dimensionless represents a dimensionless unit (e.g., counts)
	Dimensionless Unit = "1"
	// Milliseconds represents time in milliseconds
	Milliseconds Unit = "ms"
)

// CoreServiceMetrics defines core metrics that every diode service should have
var CoreServiceMetrics = map[string]MetricDefinition{
	"service.info": {
		Name:        "service_info",
		Type:        Gauge,
		Unit:        Dimensionless,
		Description: "Service information including name, version, and environment",
		Attributes:  []string{AttrServiceName, AttrServiceVersion, AttrDeploymentEnvironment},
	},
	"service.startup_attempt": {
		Name:        "service_startup_attempts_total",
		Type:        Counter,
		Unit:        Dimensionless,
		Description: "Number of service startup attempts",
		Attributes:  []string{"success"},
	},
}
