"""Tests for device ingestion."""
import pytest
from helpers.api_helper import DiodeAPIClient
from netboxlabs.diode.sdk.ingester import (Entity, Site)


@pytest.fixture(scope="module")
def diode_client(test_config, diode_client_credentials):
    """Create a Diode API client with dynamically created credentials.

    This fixture uses the session-scoped diode_client_credentials fixture
    from conftest.py to obtain valid client credentials.
    """
    client = DiodeAPIClient(
        target=test_config["diode_target"],
        name="diode-test-client",
        client_id=diode_client_credentials["client_id"],
        client_secret=diode_client_credentials["client_secret"]
    )
    yield client
    client.close()


@pytest.mark.integration
def test_ingest_device_minimal(diode_client):
    """Test ingesting a device with minimal required fields."""
    device = Site(name="Test4"
    )

    entities = [Entity(site=device)]

    response = diode_client.ingest_entities(entities)
    assert not response.errors, f"Ingestion failed with errors: {response.errors}"

    # def test_ingest_device_with_all_fields(self, diode_client):
    #     """Test ingesting a device with all available fields."""
    #     entity_data = {
    #         "entity_type": "device",
    #         "data": {
    #             "name": "test-device-full",
    #             "device_type": "switch",
    #             "site": "dc1",
    #             "role": "access-switch",
    #             "serial": "ABC123456",
    #             "asset_tag": "TAG001",
    #             "status": "active",
    #         }
    #     }
    #
    #     response = diode_client.ingest_entity(entity_data)
    #     assert response.status_code in [200, 201, 202]
    #
    # def test_ingest_device_missing_required_field(self, diode_client):
    #     """Test that ingesting a device without required fields fails appropriately."""
    #     entity_data = {
    #         "entity_type": "device",
    #         "data": {
    #             "name": "test-device-incomplete",
    #             # Missing device_type and site
    #         }
    #     }
    #
    #     response = diode_client.ingest_entity(entity_data)
    #     assert response.status_code in [400, 422]