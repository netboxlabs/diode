# Diode Pytest Tests

This directory contains pytest-based integration and unit tests for the Diode project.

## Directory Structure

```
tests/
├── conftest.py                     # Pytest configuration and shared fixtures
├── requirements.txt                # Test dependencies
├── helpers/                        # Test helper modules
│   ├── __init__.py
│   ├── api_helper.py              # API client helpers
│   └── utils.py                   # Utility functions
├── ingestion/                      # Ingestion-specific tests
│   ├── __init__.py
│   └── test_device_ingestion.py
├── test_health_check.py           # Health check tests
└── test_ingestion.py              # General ingestion tests
```

## Setup

### Install Test Dependencies

```bash
cd /home/amanda/netboxlabs/diode/tests
pip install -r requirements.txt
```

### Prerequisites

- Diode server running on `http://localhost:8081` (or configure in conftest.py)
- NetBox running on `http://localhost:8000` (for integration tests)

## Running Tests

### Run All Tests

```bash
cd /home/amanda/netboxlabs/diode
pytest tests/
```

### Run Specific Test Categories

```bash
# Run only unit tests
pytest -m unit tests/

# Run only integration tests
pytest -m integration tests/

# Run only e2e tests
pytest -m e2e tests/

# Exclude slow tests
pytest -m "not slow" tests/
```

### Run Specific Test Files

```bash
# Run health check tests
pytest tests/test_health_check.py

# Run device ingestion tests
pytest tests/ingestion/test_device_ingestion.py
```

### Run with Coverage

```bash
pytest --cov=diode-server --cov-report=html tests/
```

Coverage reports will be generated in `htmlcov/` directory.

### Run Tests in Parallel

```bash
pytest -n auto tests/
```

## Test Markers

The following markers are available:

- `@pytest.mark.unit` - Unit tests (no external dependencies)
- `@pytest.mark.integration` - Integration tests (require external services)
- `@pytest.mark.e2e` - End-to-end tests
- `@pytest.mark.slow` - Slow running tests

## Writing Tests

### Example Test

```python
import pytest
from helpers.api_helper import DiodeAPIClient


@pytest.mark.integration
def test_example(test_config):
    """Example test using fixtures."""
    client = DiodeAPIClient(base_url=test_config["diode_server_url"])
    response = client.health_check()
    assert response.status_code == 200
```

### Using Fixtures

Common fixtures are available from `conftest.py`:

- `test_config` - Test configuration dictionary
- `test_logger` - Logger instance for tests

## Configuration

Test configuration can be modified in `conftest.py`:

```python
@pytest.fixture(scope="session")
def test_config():
    return {
        "diode_server_url": "http://localhost:8081",
        "netbox_url": "http://localhost:8000",
        "timeout": 30,
    }
```

## CI/CD Integration

Add to your CI/CD pipeline:

```yaml
- name: Run tests
  run: |
    pip install -r tests/requirements.txt
    pytest tests/ -v --cov=diode-server --cov-report=xml
```