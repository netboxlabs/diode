"""Pytest configuration for integration tests.

This module provides shared fixtures and configuration for pytest-based tests.
"""
import sys
import logging
from pathlib import Path
import pytest

# Add project root and tests directory to Python path
project_root = Path(__file__).resolve().parent.parent
tests_dir = Path(__file__).resolve().parent
sys.path.insert(0, str(project_root))
sys.path.insert(0, str(tests_dir))

logger = logging.getLogger(__name__)


def pytest_configure(config):
    """Configure pytest with custom markers and settings."""
    # Add custom markers
    config.addinivalue_line(
        "markers",
        "integration: mark test as integration test requiring external services"
    )
    config.addinivalue_line(
        "markers",
        "unit: mark test as unit test (no external dependencies)"
    )
    config.addinivalue_line(
        "markers",
        "e2e: mark test as end-to-end test"
    )
    config.addinivalue_line(
        "markers",
        "slow: mark test as slow running"
    )


@pytest.fixture(scope="session")
def test_config():
    """Provide test configuration."""
    import os
    return {
        "diode_target": os.getenv("DIODE_TARGET", "grpc://localhost:8080/diode"),
        "netbox_url": os.getenv("NETBOX_URL", "http://localhost:8000"),
        "netbox_token": os.getenv("NETBOX_API_TOKEN", "0123456789abcdef0123456789abcdef01234567"),
        "timeout": 30,
    }


@pytest.fixture(scope="function")
def test_logger():
    """Provide a logger for tests."""
    return logging.getLogger("test")


@pytest.fixture(scope="function", autouse=True)
def log_test_name(request):
    """Log the name of each test as it runs."""
    test_name = request.node.name
    logger.info(f"Starting test: {test_name}")
    yield
    logger.info(f"Completed test: {test_name}")
