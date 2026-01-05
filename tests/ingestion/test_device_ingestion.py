"""Tests for device ingestion."""
import pytest
from helpers.api_helper import DiodeAPIClient


@pytest.fixture(scope="module")
def diode_client(test_config):
    """Create a Diode API client for testing."""
    client = DiodeAPIClient(base_url=test_config["diode_server_url"])
    yield client
    client.close()


@pytest.mark.integration
class TestDeviceIngestion:
    """Test cases for device ingestion."""

    def test_ingest_device_minimal(self, diode_client):
        """Test ingesting a device with minimal required fields."""
        entity_data = {
            "entity_type": "device",
            "data": {
                "name": "test-device-minimal",
                "device_type": "switch",
                "site": "dc1",
            }
        }

        response = diode_client.ingest_entity(entity_data)
        assert response.status_code in [200, 201, 202]

    def test_ingest_device_with_all_fields(self, diode_client):
        """Test ingesting a device with all available fields."""
        entity_data = {
            "entity_type": "device",
            "data": {
                "name": "test-device-full",
                "device_type": "switch",
                "site": "dc1",
                "role": "access-switch",
                "serial": "ABC123456",
                "asset_tag": "TAG001",
                "status": "active",
            }
        }

        response = diode_client.ingest_entity(entity_data)
        assert response.status_code in [200, 201, 202]

    def test_ingest_device_missing_required_field(self, diode_client):
        """Test that ingesting a device without required fields fails appropriately."""
        entity_data = {
            "entity_type": "device",
            "data": {
                "name": "test-device-incomplete",
                # Missing device_type and site
            }
        }

        response = diode_client.ingest_entity(entity_data)
        assert response.status_code in [400, 422]