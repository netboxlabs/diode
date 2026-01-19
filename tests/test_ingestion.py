"""Tests for Diode ingestion functionality."""
import pytest
from helpers.api_helper import DiodeAPIClient, NetBoxAPIClient


@pytest.fixture(scope="module")
def diode_client(test_config):
    """Create a Diode API client for testing."""
    client = DiodeAPIClient(base_url=test_config["diode_server_url"])
    yield client
    client.close()


@pytest.fixture(scope="module")
def netbox_client(test_config):
    """Create a NetBox API client for testing."""
    # Note: Add proper token handling when needed
    client = NetBoxAPIClient(base_url=test_config["netbox_url"])
    yield client
    client.close()


@pytest.mark.integration
@pytest.mark.slow
def test_ingest_device(diode_client):
    """Test ingesting a device entity."""
    entity_data = {
        "entity_type": "device",
        "data": {
            "name": "test-device-01",
            "device_type": "test-type",
            "site": "test-site",
        }
    }

    response = diode_client.ingest_entity(entity_data)
    # Note: Adjust assertions based on actual API response format
    assert response.status_code in [200, 201, 202]


@pytest.mark.e2e
@pytest.mark.slow
def test_device_ingestion_e2e(diode_client, netbox_client):
    """Test end-to-end device ingestion from Diode to NetBox."""
    device_name = "test-device-e2e-01"

    # Ingest device via Diode
    entity_data = {
        "entity_type": "device",
        "data": {
            "name": device_name,
            "device_type": "test-type",
            "site": "test-site",
        }
    }

    response = diode_client.ingest_entity(entity_data)
    assert response.status_code in [200, 201, 202]

    # Verify device was created in NetBox
    # Note: May need to add wait/retry logic for async processing
    netbox_response = netbox_client.get_device(device_name)
    assert netbox_response.status_code == 200
    devices = netbox_response.json().get("results", [])
    assert len(devices) > 0
    assert devices[0]["name"] == device_name
