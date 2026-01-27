"""Pytest configuration for integration tests.

This module provides shared fixtures and configuration for pytest-based tests.
"""
import sys
import logging
import os
import uuid
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
    return {
        "diode_target": os.getenv("DIODE_TARGET", "grpc://localhost:8080/diode"),
        "netbox_url": os.getenv("NETBOX_URL", "http://localhost:8000/netbox/"),
        "timeout": 30,
    }


@pytest.fixture(scope="function")
def test_logger():
    """Provide a logger for tests."""
    return logging.getLogger("test")


@pytest.fixture(scope="session")
def netbox_credentials():
    """Provide NetBox web authentication credentials.

    Returns:
        dict: Contains 'username' and 'password' keys for NetBox login

    Note:
        Override these values using environment variables:
        - NETBOX_USERNAME (default: "admin")
        - NETBOX_PASSWORD (default: "admin")
    """
    return {
        "username": os.getenv("NETBOX_USERNAME", "admin"),
        "password": os.getenv("NETBOX_PASSWORD", "admin"),
    }


@pytest.fixture(scope="function")
def netbox_web_client(test_config, netbox_credentials):
    """Create authenticated NetBox web client for plugin endpoints.

    This fixture creates a client that can interact with NetBox plugin
    web views (not REST API). It handles Django session authentication
    and CSRF tokens automatically.

    Returns:
        NetBoxPluginWebClient: Authenticated client ready to use

    Example:
        def test_get_settings(netbox_web_client):
            response = netbox_web_client.get_settings()
            assert response.status_code == 200
    """
    from helpers.api_helper import NetBoxPluginWebClient

    client = NetBoxPluginWebClient(
        base_url=test_config["netbox_url"],
        username=netbox_credentials["username"],
        password=netbox_credentials["password"]
    )

    # Perform login
    if not client.login():
        pytest.fail(f"Failed to login to NetBox at {test_config['netbox_url']}")

    yield client
    client.close()


@pytest.fixture(scope="function", autouse=True)
def log_test_name(request):
    """Log the name of each test as it runs."""
    test_name = request.node.name
    logger.info(f"Starting test: {test_name}")
    yield
    logger.info(f"Completed test: {test_name}")


@pytest.fixture(scope="function")
def diode_client_credential(netbox_web_client):
    """Create a test client credential and return its details.

    This fixture creates a new client credential via the NetBox web interface,
    follows the redirect to the secret page, and extracts the client_id and
    client_secret from the response.

    Returns:
        dict: Contains 'client_id', 'client_secret', and 'client_name' keys

    Example:
        def test_something(diode_client_credential):
            client_id = diode_client_credential['client_id']
            client_secret = diode_client_credential['client_secret']
    """
    import re

    client_name = f"pytest-test-{uuid.uuid4()}"

    # Create credential
    response = netbox_web_client.add_credential(client_name)

    assert response.status_code == 302, pytest.fail(f"Failed to create test credential: {response.status_code}")

    # Follow redirect to secret page
    secret_url = response.headers["Location"]
    secret_response = netbox_web_client.session.get(
        f"{netbox_web_client.base_url.rstrip('netbox/')}{secret_url}"
    )

    assert secret_response.status_code == 200, pytest.fail(f"Failed to get secret page: {secret_response.status_code}")

    # Extract client_id and client_secret from HTML input fields
    # Find the input tag with data-clipboard="client-id" or "client-secret"
    client_id_input = re.search(r'<input[^>]*data-clipboard=["\']client-id["\'][^>]*>', secret_response.text)
    client_secret_input = re.search(r'<input[^>]*data-clipboard=["\']client-secret["\'][^>]*>', secret_response.text)

    if not client_id_input or not client_secret_input:
        pytest.fail("Failed to find client_id or client_secret input fields in secret page")

    # Extract value attribute from the input tags
    client_id_match = re.search(r'value=["\']([^"\']+)["\']', client_id_input.group(0))
    client_secret_match = re.search(r'value=["\']([^"\']+)["\']', client_secret_input.group(0))

    if not client_id_match or not client_secret_match:
        pytest.fail("Failed to extract value from client_id or client_secret input fields")

    credential = {
        "client_name": client_name,
        "client_id": client_id_match.group(1),
        "client_secret": client_secret_match.group(1),
    }

    logger.info(f"Created test credential: {credential['client_id']}")

    yield credential

    # Cleanup: Delete the credential
    try:
        netbox_web_client.delete_credential(credential['client_id'])
        logger.info(f"Deleted test credential: {credential['client_id']}")
    except Exception as e:
        logger.warning(f"Failed to delete test credential {credential['client_id']}: {e}")


@pytest.fixture(scope="function")
def diode_client(test_config, diode_client_credential):
    """Create a Diode API client with dynamically created credentials.

    This fixture uses the diode_client_credential fixture
    from conftest.py to obtain valid client credentials.

    Returns:
        DiodeAPIClient: Configured Diode API client ready to use

    Example:
        def test_ingest(diode_client):
            response = diode_client.ingest_entities(entities)
            assert not response.errors
    """
    from helpers.api_helper import DiodeAPIClient

    client = DiodeAPIClient(
        target=test_config["diode_target"],
        name="diode-test-client",
        client_id=diode_client_credential["client_id"],
        client_secret=diode_client_credential["client_secret"]
    )
    yield client
    client.close()


@pytest.fixture(scope="function")
def netbox_api_client(netbox_web_client):
    """Create a NetBox API client using web client's authenticated session.

    This fixture reuses the authenticated session from netbox_web_client,
    avoiding the need for a separate API token.

    Returns:
        NetBoxAPIClient: Configured NetBox API client ready to use

    Example:
        def test_get_sites(netbox_api_client):
            response = netbox_api_client.get_sites()
            assert response.status_code == 200
    """
    from helpers.api_helper import NetBoxAPIClient

    # Create client and replace its session with the authenticated web client session
    client = NetBoxAPIClient(
        base_url=netbox_web_client.base_url,
        token=None
    )
    # Use the web client's authenticated session instead of creating a new one
    client.session = netbox_web_client.session

    yield client
    # Don't close the session since it belongs to netbox_web_client
