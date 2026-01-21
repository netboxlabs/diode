"""Tests for device ingestion."""
import uuid
import re
import time

import pytest
from netboxlabs.diode.sdk.ingester import (Entity, Site)


@pytest.mark.integration
def test_ingest_device_minimal(diode_client, netbox_api_client, netbox_web_client):
    """Test ingesting a device with minimal required fields.

    This test verifies the complete flow:
    1. Verifies site doesn't exist in NetBox
    2. Ingests site via Diode
    3. Verifies deviation appears in deviations page
    4. Applies deviation to main branch
    5. Verifies site was created in NetBox
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

    # Step 3: Verify deviation appears in deviations page (with retry)
    max_retries = 5
    retry_delay = 1  # seconds
    deviations_page = None

    for attempt in range(max_retries):
        deviations_page = netbox_web_client.get("/plugins/assurance/deviations-all/")
        assert deviations_page.status_code == 200, \
            f"Failed to get deviations page: {deviations_page.status_code}"

        if site_name in deviations_page.text:
            break

        if attempt < max_retries - 1:
            time.sleep(retry_delay)
    else:
        pytest.fail(f"Site '{site_name}' did not appear in deviations page after {max_retries} attempts")

    # Step 4: Extract deviation ID from the page and apply it
    # Look for deviation ID in the HTML (usually in links or data attributes)
    deviation_id_match = re.search(
        rf'/plugins/assurance/deviations/([a-f0-9-]+)/.*{re.escape(site_name)}',
        deviations_page.text
    )
    if not deviation_id_match:
        # Try alternative pattern
        deviation_id_match = re.search(
            rf'deviation["\s-]+([a-f0-9-]+)["\s>].*{re.escape(site_name)}',
            deviations_page.text,
            re.IGNORECASE
        )

    assert deviation_id_match, \
        f"Could not find deviation ID for site '{site_name}' in deviations page"

    deviation_id = deviation_id_match.group(1)

    # Apply deviation to main branch (using web client, not API)
    apply_response = netbox_web_client.post(
        f"/plugins/assurance/deviations/{deviation_id}/apply/",
        data={"confirm": "1"},  # Confirmation required to apply
        allow_redirects=False
    )
    assert apply_response.status_code == 302, \
        f"Failed to apply deviation: {apply_response.status_code} - {apply_response.text}"

    # Step 5: Verify site was created in NetBox (with retry)
    sites_after = None
    for attempt in range(max_retries):
        sites_after = netbox_api_client.get_sites(name=site_name)
        assert sites_after.status_code == 200, \
            f"Failed to get sites after applying deviation: {sites_after.status_code}"

        if sites_after.json()["count"] == 1:
            break

        if attempt < max_retries - 1:
            time.sleep(retry_delay)
    else:
        pytest.fail(f"Site '{site_name}' was not created in NetBox after {max_retries} attempts")

    assert sites_after.json()["results"][0]["name"] == site_name, \
        f"Retrieved site name doesn't match: expected '{site_name}'"


