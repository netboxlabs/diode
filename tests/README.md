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

### Install Dependencies

Before running tests, create a virtual environment and install dependencies:

```bash
# Create and activate virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install test dependencies
pip install -r tests/requirements.txt
```

### Run Tests

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

## Linting

This project uses [Ruff](https://docs.astral.sh/ruff/) for Python linting and formatting.

### Run Linting

Check for linting issues:
```bash
ruff check .
```

Automatically fix linting issues:
```bash
ruff check --fix .
```

### Run Formatting

Format code:
```bash
ruff format .
```

Check formatting without making changes:
```bash
ruff format --check .
```

### Configuration

Linting configuration is defined in `pyproject.toml`. The configuration:
- Targets Python 3.12
- Sets line length to 120 characters
- Excludes protobuf generated files (`*.pb2.py`, `*.pb2_grpc.py`)
- Enables pycodestyle, pyflakes, isort, flake8-bugbear, flake8-comprehensions, and pyupgrade rules

## Test Structure

- `tests/`: Integration tests for the Diode SDK and NetBox plugin