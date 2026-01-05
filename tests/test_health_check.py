"""Tests for Diode server health check endpoint."""
import pytest
from helpers.api_helper import DiodeAPIClient


@pytest.fixture(scope="module")
def diode_client(test_config):
    """Create a Diode API client for testing."""
    client = DiodeAPIClient(base_url=test_config["diode_server_url"])
    yield client
    client.close()


@pytest.mark.integration
def test_health_check_returns_200(diode_client):
    """Test that health check endpoint returns 200 status."""
    response = diode_client.health_check()
    assert response.status_code == 200


@pytest.mark.integration
def test_health_check_response_format(diode_client):
    """Test that health check response has expected format."""
    response = diode_client.health_check()
    assert response.status_code == 200
    data = response.json()
    assert "status" in data
