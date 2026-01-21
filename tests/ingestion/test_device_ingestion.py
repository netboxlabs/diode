"""Tests for device ingestion."""
import uuid
import time

import pytest
from netboxlabs.diode.sdk.ingester import (Entity, Site)


@pytest.mark.integration
def test_ingest_device_minimal(diode_client, netbox_api_client):
    """Test ingesting a site with minimal required fields.

    This test verifies the complete flow:
    1. Verifies site doesn't exist in NetBox
    2. Ingests site via Diode
    3. Verifies site was created in NetBox
    """
    site_name = f"pytest-test-{uuid.uuid4()}"

    # Step 1: Verify site doesn't exist in NetBox before ingestion
    sites_before = netbox_api_client.get_sites(name=site_name)
    assert sites_before.status_code == 200, \
        f"Failed to get sites: {sites_before.status_code}"
    assert sites_before.json()["count"] == 0, \
        f"Site '{site_name}' already exists in NetBox before test"

    # Step 2: Ingest site via Diode
    site = Site(name=site_name)
    entities = [Entity(site=site)]
    response = diode_client.ingest_entities(entities)
    assert not response.errors, f"Ingestion failed with errors: {response.errors}"

    # Step 3: Verify site was created in NetBox (with retry)
    max_retries = 5
    retry_delay = 1  # seconds
    sites_after = None

    for attempt in range(max_retries):
        sites_after = netbox_api_client.get_sites(name=site_name)
        assert sites_after.status_code == 200, \
            f"Failed to get sites after ingestion: {sites_after.status_code}"

        if sites_after.json()["count"] == 1:
            break

        if attempt < max_retries - 1:
            time.sleep(retry_delay)
    else:
        pytest.fail(f"Site '{site_name}' was not created in NetBox after {max_retries} attempts")

    assert sites_after.json()["results"][0]["name"] == site_name, \
        f"Retrieved site name doesn't match: expected '{site_name}'"


