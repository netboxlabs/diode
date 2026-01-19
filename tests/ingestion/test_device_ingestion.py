"""Tests for device ingestion."""
import os
import pytest
import requests
from helpers.api_helper import DiodeAPIClient
from netboxlabs.diode.sdk.ingester import (Entity, Site)


@pytest.fixture(scope="module")
def diode_client_credentials(test_config):
    """Create Diode client credentials dynamically via Diode API.

    This fixture attempts to create new client credentials by calling
    the Diode server API directly. It requires admin credentials to be
    set in environment variables or test config.

    If creation fails, it falls back to using hardcoded test credentials.
    """
    # Try to get admin credentials for creating new clients
    admin_client_id = os.getenv("DIODE_ADMIN_CLIENT_ID")
    admin_client_secret = os.getenv("DIODE_ADMIN_CLIENT_SECRET")

    if admin_client_id and admin_client_secret:
        try:
            # Get the Diode API base URL (convert gRPC target to HTTP)
            diode_target = test_config["diode_target"]
            # Extract base URL from gRPC target (e.g., grpc://localhost:8080/diode -> http://localhost:8080)
            base_url = diode_target.replace("grpc://", "http://").rsplit("/", 1)[0]

            # First, authenticate to get a token
            auth_response = requests.post(
                f"{base_url}/token",
                data={
                    "grant_type": "client_credentials",
                    "client_id": admin_client_id,
                    "client_secret": admin_client_secret,
                    "scope": "diode:read diode:write"
                }
            )

            if auth_response.status_code == 200:
                token = auth_response.json()["access_token"]

                # Create new client credentials
                create_response = requests.post(
                    f"{base_url}/clients",
                    headers={
                        "Authorization": f"Bearer {token}",
                        "Content-Type": "application/json"
                    },
                    json={
                        "client_name": "pytest-test-client",
                        "scope": "diode:ingest"
                    }
                )

                if create_response.status_code == 201:
                    credentials = create_response.json()
                    yield credentials
                    return

        except Exception as e:
            print(f"Warning: Failed to create dynamic credentials: {e}")
            print("Falling back to hardcoded test credentials")

    # Fallback to hardcoded credentials
    yield {
        "client_id": "test-client-cre-63125734380a4cbf",
        "client_secret": "LTZ9si6rpRXcpEB5ZSrz6jZ4xtUsuMFE+EQt3r+IWog="
    }


@pytest.fixture(scope="module")
def diode_client(test_config, diode_client_credentials):
    """Create a Diode API client with dynamically created credentials."""
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