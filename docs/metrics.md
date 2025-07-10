# Metrics

## Available Metrics

### Core Metrics (All Services)
- `diode_{service}_service_info` - Service information
- `diode_{service}_service_startup_attempts_total` - Startup attempts

### Service-Specific Metrics

#### Auth Service
- `diode_auth_service_info` - Service information
- `diode_auth_service_startup_attempts_total` - Startup attempts

#### Ingester Service
- `diode_ingester_service_info` - Service information
- `diode_ingester_service_startup_attempts_total` - Startup attempts
- `diode_ingester_ingest_requests_total` - Ingest requests (success)
- `diode_ingester_ingest_entities_total` - Entities ingested

#### Reconciler Service
- `diode_reconciler_service_info` - Service information
- `diode_reconciler_service_startup_attempts_total` - Startup attempts
- `diode_reconciler_handle_message_total` - Messages handled (success)
- `diode_reconciler_ingestion_log_create_total` - Ingestion logs created (success)
- `diode_reconciler_change_set_create_total` - Change sets created (success)
- `diode_reconciler_change_set_apply_total` - Change sets applied (success)
- `diode_reconciler_change_create_total` - Individual changes created
- `diode_reconciler_change_apply_total` - Individual changes applied


## Configuration Overview

Diode's telemetry implementation supports multiple configuration methods and exporters.

### Exporters
- **prometheus**: Serve metrics on `/metrics` endpoint
- **otlp**: Send to OpenTelemetry Collector
- **console**: Output to stdout (debugging)
- **none**: Disable metrics

## Environment Variables

### Core Configuration

| Variable | Description | Default | Required | Example |
|----------|-------------|---------|----------|---------|
| `TELEMETRY_SERVICE_NAME` | Service identifier | Auto-detected | No | `netboxlabs/diode/auth` |
| `TELEMETRY_ENVIRONMENT` | Deployment environment | `dev` | No | `production` |
| `TELEMETRY_METRICS_EXPORTER` | Metrics export type | `none` | No | `prometheus` |
| `TELEMETRY_METRICS_PORT` | Prometheus metrics port | `9090` | No | `9090` |

### OTLP Configuration

If the otlp

| Variable | Description | Default | Required | Example |
|----------|-------------|---------|----------|---------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint | `http://localhost:4317` | No | `http://collector:4317` |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | OTLP protocol | `grpc` | No | `grpc` |
| `OTEL_EXPORTER_OTLP_HEADERS` | OTLP headers | None | No | `authorization=Bearer token` |
| `OTEL_EXPORTER_OTLP_TIMEOUT` | OTLP timeout | `10s` | No | `30s` |

For further details refer to the [OpenTelemetry Environment Variable Specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration).

## Docker Compose Example
```yaml
# docker-compose.yaml
services:
  diode-auth:
    environment:
      - TELEMETRY_METRICS_EXPORTER=prometheus
      - TELEMETRY_METRICS_PORT=9090
      - TELEMETRY_ENVIRONMENT=dev
    ports:
      - "8080:8080"
      - "9090:9090"

  diode-ingester:
    environment:
      - TELEMETRY_METRICS_EXPORTER=prometheus
      - TELEMETRY_METRICS_PORT=9090
      - TELEMETRY_ENVIRONMENT=dev
    ports:
      - "8081:8081"
      - "9091:9090"

  diode-reconciler:
    environment:
      - TELEMETRY_METRICS_EXPORTER=prometheus
      - TELEMETRY_METRICS_PORT=9090
      - TELEMETRY_ENVIRONMENT=dev
    ports:
      - "8082:8081"
      - "9092:9090"
```

## Troubleshooting

### Metrics Not Appearing
1. Check `TELEMETRY_METRICS_EXPORTER=prometheus`
2. Verify endpoint accessibility
3. Check service logs for errors

### Debug using console exporter
```bash
export TELEMETRY_METRICS_EXPORTER=console
```