# Tests

This directory contains integration tests for the Diode project, using pytest.

## Prerequisites

To run the tests, you'll need:

- Python 3.9+
- A running Diode server
- A running NetBox instance with the Diode plugin installed

## Setup

### 1. Configure Environment Variables

The tests read configuration from environment variables. Configure them according to your setup:

#### Environment Variables

**NetBox connection (required for external NetBox):**
- `NETBOX_URL`: URL of your NetBox instance (default: `http://localhost:8000/netbox/`)
- `NETBOX_USERNAME`: NetBox web UI username (default: `admin`)
- `NETBOX_PASSWORD`: NetBox web UI password (default: `admin`)

**Diode server (optional):**
- `DIODE_TARGET`: Diode gRPC server URL (default: `grpc://localhost:8080/diode`)

**Note**: The tests automatically create Diode client credentials dynamically via the NetBox plugin web interface during test execution. You don't need to manually configure `DIODE_ADMIN_CLIENT_ID` or `DIODE_ADMIN_CLIENT_SECRET`.

#### Setting Environment Variables

You can set these variables in your shell before running tests:

```bash
export NETBOX_URL="http://my-netbox-server:8000/netbox/"
export NETBOX_USERNAME="admin"
export NETBOX_PASSWORD="admin"
```

## Running Tests

Run all tests:
```bash
pytest tests/
```

Run specific test files:
```bash
pytest tests/test_ingestion.py
```

Run tests with verbose output:
```bash
pytest tests/ -v
```

Run tests with coverage:
```bash
pytest tests/ --cov=diode
```

## Test Structure

- `tests/`: Integration tests for the Diode SDK and NetBox plugin
- `tests/.env.example`: Template for test configuration
- `tests/.env`: Your local configuration (not tracked in git)