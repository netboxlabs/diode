package metrics

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricManager provides a standardized way to create and use metrics
type MetricManager struct {
	environment     string
	metricNamespace string
	meter           metric.Meter
	serviceName     string
	counters        map[string]metric.Int64Counter
	histograms      map[string]metric.Float64Histogram
	gauges          map[string]metric.Int64Gauge
	updowns         map[string]metric.Int64UpDownCounter
}

// NewMetricManager creates a new metric manager with the given service-specific metrics
func NewMetricManager(environment, metricNamespace string, meter metric.Meter, serviceName string, serviceMetrics map[string]MetricDefinition) (*MetricManager, error) {
	manager := &MetricManager{
		environment:     environment,
		metricNamespace: metricNamespace,
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

	if err := manager.createMetrics(allMetrics); err != nil {
		return nil, err
	}

	return manager, nil
}

// createMetrics creates all the metrics based on their definitions
func (m *MetricManager) createMetrics(definitions map[string]MetricDefinition) error {
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
func (m *MetricManager) buildMetricName(name string) string {
	return fmt.Sprintf("%s_%s_%s", m.metricNamespace, m.serviceName, name)
}

// RecordCounter increments a counter metric
func (m *MetricManager) RecordCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if counter, exists := m.counters[name]; exists {
		counter.Add(ctx, value, metric.WithAttributes(attrs...))
	}
}

// RecordHistogram records a value in a histogram metric
func (m *MetricManager) RecordHistogram(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	if histogram, exists := m.histograms[name]; exists {
		histogram.Record(ctx, value, metric.WithAttributes(attrs...))
	}
}

// SetGauge sets the value of a gauge metric
func (m *MetricManager) SetGauge(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.Record(ctx, value, metric.WithAttributes(attrs...))
	}
}

// RecordUpDownCounter modifies an up/down counter metric
func (m *MetricManager) RecordUpDownCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	if updown, exists := m.updowns[name]; exists {
		updown.Add(ctx, value, metric.WithAttributes(attrs...))
	}
}

// SetServiceInfo sets the service information metric (called once at startup)
func (m *MetricManager) SetServiceInfo(ctx context.Context, version string) {
	m.SetGauge(ctx, "service.info", 1,
		attribute.String(AttrServiceName, m.serviceName),
		attribute.String(AttrServiceVersion, version),
		attribute.String(AttrDeploymentEnvironment, m.environment), // or other environment
	)
}
