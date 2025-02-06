package telemetry

// Config holds OpenTelemetry configuration settings
// Note the environment variables may be prefixed with TELEMETRY_
// based on the application's configuration structure
type Config struct {
	// ServiceName is the name of the service being instrumented
	ServiceName string `envconfig:"SERVICE_NAME"`

	// Environment represents the deployment environment (e.g., prod, staging, dev)
	Environment string `envconfig:"ENVIRONMENT" default:"dev"`

	// MetricsExporter represents the type of exporter to use. oltp,console and none are supported
	MetricsExporter string `envconfig:"METRICS_EXPORTER" default:"none"`

	// MetricsPort is the port to serve the metrics on if MetricsExporter is prometheus
	MetricsPort int `envconfig:"METRICS_PORT" default:"9090"`

	// TracesExporter represents the type of exporter to use. oltp,console and none are supported
	TracesExporter string `envconfig:"TRACES_EXPORTER" default:"none"`

	// Additional environment variables used interally by otel
	// can be found in the otel exporters documentation eg
	// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
	// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
	// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
}
