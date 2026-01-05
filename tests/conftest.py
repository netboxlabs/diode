"""Pytest configuration for integration tests.

This module provides shared fixtures and configuration for pytest-based tests.
"""
import sys
import logging
from pathlib import Path
import pytest

# Add project root to Python path
project_root = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(project_root))

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
    return {
        "diode_server_url": "http://localhost:8081",
        "netbox_url": "http://localhost:8000",
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
