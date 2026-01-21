"""Pytest configuration for integration tests.

This module provides shared fixtures and configuration for pytest-based tests.
"""
import sys
import logging
import os
from pathlib import Path
import pytest
import requests

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


@pytest.fixture(scope="session")
def diode_client_credentials(test_config):
    """Create Diode client credentials dynamically via Diode API.

    This is a session-scoped fixture that creates client credentials once
    for all tests in the session. Credentials are created by authenticating
    with admin credentials and calling the Diode API.

    Required environment variables:
        - DIODE_ADMIN_CLIENT_ID: Admin client ID with permission to create clients
        - DIODE_ADMIN_CLIENT_SECRET: Admin client secret

    The fixture will skip all tests if admin credentials are not configured.

    Returns:
        dict: Contains 'client_id' and 'client_secret' keys

    Raises:
        pytest.skip: If admin credentials are not configured
        pytest.fail: If credential creation fails
    """
    # Get admin credentials for creating new clients
    admin_client_id = os.getenv("DIODE_ADMIN_CLIENT_ID")
    admin_client_secret = os.getenv("DIODE_ADMIN_CLIENT_SECRET")

    if not admin_client_id or not admin_client_secret:
        pytest.fail(
            "Diode admin credentials not configured. "
            "Please set DIODE_ADMIN_CLIENT_ID and DIODE_ADMIN_CLIENT_SECRET "
            "environment variables to run integration tests. "
            "See tests/.env.example for configuration details."
        )

    # Get the Diode API base URL (convert gRPC target to HTTP)
    diode_target = test_config["diode_target"]
    # Extract base URL from gRPC target (e.g., grpc://localhost:8080/diode -> http://localhost:8080)
    base_url = diode_target.replace("grpc://", "http://")

    logger.info("Authenticating with Diode server to create test credentials...")

    # Authenticate to get a token
    try:
        auth_response = requests.post(
            f"{base_url}/auth/token",
            data={
                "grant_type": "client_credentials",
                "client_id": admin_client_id,
                "client_secret": admin_client_secret,
                "scope": "diode:ingest"
            },
            timeout=10
        )
    except requests.RequestException as e:
        pytest.fail(f"Failed to connect to Diode server at {base_url}: {e}")

    if auth_response.status_code != 200:
        pytest.fail(
            f"Failed to authenticate with Diode server: {auth_response.status_code} - {auth_response.text}"
        )

    token = auth_response.json()["access_token"]
    logger.info("Successfully authenticated with Diode server")

    # Create new client credentials
    try:
        create_response = requests.post(
            f"{base_url}/clients",
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json"
            },
            json={
                "client_name": f"pytest-session-{os.getpid()}",
                "scope": "diode:ingest"
            },
            timeout=10
        )
    except requests.RequestException as e:
        pytest.fail(f"Failed to create client credentials: {e}")

    if create_response.status_code != 201:
        pytest.fail(
            f"Failed to create client credentials: {create_response.status_code} - {create_response.text}"
        )

    credentials = create_response.json()
    logger.info(f"Created test credentials: {credentials['client_id']}")

    yield credentials

    # TODO: Optional cleanup - delete the created credentials after all tests
    # This would require implementing a delete endpoint call
    logger.info(f"Test session complete. Credentials {credentials['client_id']} may need manual cleanup.")
