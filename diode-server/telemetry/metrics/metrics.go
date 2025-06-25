package metrics

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	// MetricNamespace is the namespace for all Diode metrics
	MetricNamespace = "diode"

	// AttrServiceName is the attribute key for a service name
	AttrServiceName = string(semconv.ServiceNameKey)

	// AttrServiceVersion is the attribute key for a service version
	AttrServiceVersion = string(semconv.ServiceVersionKey)

	// AttrDeploymentEnvironment is the attribute key for the deployment environment (e.g., production, staging)
	AttrDeploymentEnvironment = "deployment.environment"
)

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
		Description: "Service information and version (always 1)",
		Attributes:  []string{AttrServiceName, AttrServiceVersion, AttrDeploymentEnvironment},
	},
}
