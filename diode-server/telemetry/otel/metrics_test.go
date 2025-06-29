package otel

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestMetricType_String(t *testing.T) {
	tests := []struct {
		metricType MetricType
		expected   string
	}{
		{Counter, "counter"},
		{Gauge, "gauge"},
		{Histogram, "histogram"},
		{UpDown, "updown"},
		{MetricType(999), "unknown"}, // invalid type
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("MetricType_%d", tt.metricType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.metricType.String())
		})
	}
}

func TestNewMetricRecorder(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	tests := []struct {
		name           string
		environment    string
		serviceName    string
		serviceMetrics map[string]MetricDefinition
		expectError    bool
	}{
		{
			name:           "valid recorder with core metrics only",
			environment:    "test",
			serviceName:    "test-service",
			serviceMetrics: map[string]MetricDefinition{},
			expectError:    false,
		},
		{
			name:        "valid recorder with additional metrics",
			environment: "production",
			serviceName: "ingester",
			serviceMetrics: map[string]MetricDefinition{
				"test.counter": {
					Name:        "test_counter",
					Type:        Counter,
					Unit:        Dimensionless,
					Description: "Test counter metric",
					Attributes:  []string{"test"},
				},
				"test.histogram": {
					Name:        "test_histogram",
					Type:        Histogram,
					Unit:        Milliseconds,
					Description: "Test histogram metric",
					Attributes:  []string{"test"},
				},
				"test.gauge": {
					Name:        "test_gauge",
					Type:        Gauge,
					Unit:        Dimensionless,
					Description: "Test gauge metric",
					Attributes:  []string{"test"},
				},
				"test.updown": {
					Name:        "test_updown",
					Type:        UpDown,
					Unit:        Dimensionless,
					Description: "Test updown counter metric",
					Attributes:  []string{"test"},
				},
			},
			expectError: false,
		},
		{
			name:        "recorder with invalid metric type",
			environment: "test",
			serviceName: "test-service",
			serviceMetrics: map[string]MetricDefinition{
				"invalid.metric": {
					Name:        "invalid_metric",
					Type:        MetricType(999), // invalid type
					Unit:        Dimensionless,
					Description: "Invalid metric",
					Attributes:  []string{},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder, err := NewMetricRecorder(meter, tt.environment, tt.serviceName, tt.serviceMetrics)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, recorder)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, recorder)
				assert.Equal(t, tt.environment, recorder.environment)
				assert.Equal(t, tt.serviceName, recorder.serviceName)
				assert.Equal(t, MetricsNamespace, recorder.metricNamespace)
				assert.NotNil(t, recorder.counters)
				assert.NotNil(t, recorder.histograms)
				assert.NotNil(t, recorder.gauges)
				assert.NotNil(t, recorder.updowns)

				// Verify core metrics are present
				assert.Contains(t, recorder.gauges, "service.info")

				// Verify service-specific metrics are present
				for key, def := range tt.serviceMetrics {
					switch def.Type {
					case Counter:
						assert.Contains(t, recorder.counters, key)
					case Histogram:
						assert.Contains(t, recorder.histograms, key)
					case Gauge:
						assert.Contains(t, recorder.gauges, key)
					case UpDown:
						assert.Contains(t, recorder.updowns, key)
					}
				}
			}
		})
	}
}

func TestMetricRecorder_buildMetricName(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	recorder, err := NewMetricRecorder(meter, "test", "test-service", map[string]MetricDefinition{})
	require.NoError(t, err)

	tests := []struct {
		name     string
		expected string
	}{
		{"test_metric", "diode_test-service_test_metric"},
		{"requests_total", "diode_test-service_requests_total"},
		{"", "diode_test-service_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := recorder.buildMetricName(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricRecorder_RecordCounter(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	serviceMetrics := map[string]MetricDefinition{
		"test.counter": {
			Name:        "test_counter",
			Type:        Counter,
			Unit:        Dimensionless,
			Description: "Test counter",
		},
	}

	recorder, err := NewMetricRecorder(meter, "test", "test-service", serviceMetrics)
	require.NoError(t, err)

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("test", "value"),
	}

	// Test recording existing counter
	recorder.RecordCounter(ctx, "test.counter", 5, attrs...)

	// Test recording non-existent counter (should not panic)
	recorder.RecordCounter(ctx, "nonexistent", 1, attrs...)
}

func TestMetricRecorder_RecordHistogram(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	serviceMetrics := map[string]MetricDefinition{
		"test.histogram": {
			Name:        "test_histogram",
			Type:        Histogram,
			Unit:        Milliseconds,
			Description: "Test histogram",
		},
	}

	recorder, err := NewMetricRecorder(meter, "test", "test-service", serviceMetrics)
	require.NoError(t, err)

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("test", "value"),
	}

	// Test recording existing histogram
	recorder.RecordHistogram(ctx, "test.histogram", 123.45, attrs...)

	// Test recording non-existent histogram (should not panic)
	recorder.RecordHistogram(ctx, "nonexistent", 1.0, attrs...)
}

func TestMetricRecorder_SetGauge(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	serviceMetrics := map[string]MetricDefinition{
		"test.gauge": {
			Name:        "test_gauge",
			Type:        Gauge,
			Unit:        Dimensionless,
			Description: "Test gauge",
		},
	}

	recorder, err := NewMetricRecorder(meter, "test", "test-service", serviceMetrics)
	require.NoError(t, err)

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("test", "value"),
	}

	// Test setting existing gauge
	recorder.SetGauge(ctx, "test.gauge", 100, attrs...)

	// Test setting non-existent gauge (should not panic)
	recorder.SetGauge(ctx, "nonexistent", 1, attrs...)
}

func TestMetricRecorder_RecordUpDownCounter(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	serviceMetrics := map[string]MetricDefinition{
		"test.updown": {
			Name:        "test_updown",
			Type:        UpDown,
			Unit:        Dimensionless,
			Description: "Test updown counter",
		},
	}

	recorder, err := NewMetricRecorder(meter, "test", "test-service", serviceMetrics)
	require.NoError(t, err)

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("test", "value"),
	}

	// Test recording existing updown counter
	recorder.RecordUpDownCounter(ctx, "test.updown", -5, attrs...)

	// Test recording non-existent updown counter (should not panic)
	recorder.RecordUpDownCounter(ctx, "nonexistent", 1, attrs...)
}

func TestMetricRecorder_SetServiceInfo(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	recorder, err := NewMetricRecorder(meter, "production", "ingester", map[string]MetricDefinition{})
	require.NoError(t, err)

	ctx := context.Background()
	version := "v1.2.3"

	// This should not panic and should set the service.info gauge
	recorder.SetServiceInfo(ctx, version)

	// Verify the gauge exists (it's created as part of core metrics)
	assert.Contains(t, recorder.gauges, "service.info")
}

func TestCoreServiceMetrics(t *testing.T) {
	// Verify core metrics are properly defined
	assert.Contains(t, CoreServiceMetrics, "service.info")
	assert.Contains(t, CoreServiceMetrics, "service.startup_attempt")

	serviceInfo := CoreServiceMetrics["service.info"]
	assert.Equal(t, "service_info", serviceInfo.Name)
	assert.Equal(t, Gauge, serviceInfo.Type)
	assert.Equal(t, Dimensionless, serviceInfo.Unit)
	assert.NotEmpty(t, serviceInfo.Description)
	assert.Contains(t, serviceInfo.Attributes, AttrServiceName)
	assert.Contains(t, serviceInfo.Attributes, AttrServiceVersion)
	assert.Contains(t, serviceInfo.Attributes, AttrDeploymentEnvironment)

	startupAttempt := CoreServiceMetrics["service.startup_attempt"]
	assert.Equal(t, "service_startup_attempts_total", startupAttempt.Name)
	assert.Equal(t, Counter, startupAttempt.Type)
	assert.Equal(t, Dimensionless, startupAttempt.Unit)
	assert.NotEmpty(t, startupAttempt.Description)
	assert.Contains(t, startupAttempt.Attributes, "success")
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "diode", MetricsNamespace)
	assert.Equal(t, "service.name", AttrServiceName)
	assert.Equal(t, "service.version", AttrServiceVersion)
	assert.Equal(t, "deployment.environment", AttrDeploymentEnvironment)

	// Test units
	assert.Equal(t, Unit("1"), Dimensionless)
	assert.Equal(t, Unit("ms"), Milliseconds)
}

func TestMetricRecorder_createMetrics_Coverage(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	// Test all metric types to ensure complete coverage
	allTypeMetrics := map[string]MetricDefinition{
		"counter.metric": {
			Name:        "counter_metric",
			Type:        Counter,
			Unit:        Dimensionless,
			Description: "Counter metric",
		},
		"histogram.metric": {
			Name:        "histogram_metric",
			Type:        Histogram,
			Unit:        Milliseconds,
			Description: "Histogram metric",
		},
		"gauge.metric": {
			Name:        "gauge_metric",
			Type:        Gauge,
			Unit:        Dimensionless,
			Description: "Gauge metric",
		},
		"updown.metric": {
			Name:        "updown_metric",
			Type:        UpDown,
			Unit:        Dimensionless,
			Description: "UpDown counter metric",
		},
	}

	recorder, err := NewMetricRecorder(meter, "test", "test-service", allTypeMetrics)
	require.NoError(t, err)

	// Verify all metrics were created
	assert.Contains(t, recorder.counters, "counter.metric")
	assert.Contains(t, recorder.histograms, "histogram.metric")
	assert.Contains(t, recorder.gauges, "gauge.metric")
	assert.Contains(t, recorder.updowns, "updown.metric")
}

func TestMetricRecorder_RecordServiceStartupAttempt(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	recorder, err := NewMetricRecorder(meter, "test", "test-service", map[string]MetricDefinition{})
	require.NoError(t, err)

	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful startup attempt",
			success: true,
		},
		{
			name:    "failed startup attempt",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// This should not panic and should record the counter
			recorder.RecordServiceStartupAttempt(ctx, tt.success)

			// Verify the counter exists (it's created as part of core metrics)
			assert.Contains(t, recorder.counters, "service.startup_attempt")
		})
	}
}
