"""Tests for NetBox Diode Plugin web endpoints.

These tests verify that the plugin web UI endpoints are accessible
and functioning correctly. They use Django session authentication
(not API tokens) to access the endpoints.
"""
import pytest
import re


@pytest.mark.integration
def test_get_settings_page(netbox_web_client):
    """Test accessing the Diode plugin settings page.

    This test verifies that an authenticated user can access
    the settings page and see the Diode target configuration.
    """
    response = netbox_web_client.get_settings()

    assert response.status_code == 200, \
        f"Expected 200 OK, got {response.status_code}: {response.text}"

    # Verify page contains expected content
    assert "Settings" in response.text, \
        "Settings page should have 'Settings' title"

    # Verify Diode target format: starts with grpc:// and ends with /diode
    diode_target_pattern = r'grpc://[^\s<]+/diode'
    assert re.search(diode_target_pattern, response.text), \
        "Settings page should contain Diode target starting with 'grpc://' and ending with '/diode'"

    # Verify Branch shows Main (default)
    assert "Main (default)" in response.text, \
        "Settings page should display 'Main (default)' as the branch"


@pytest.mark.integration
def test_get_credentials_list_page(netbox_web_client):
    """Test accessing the client credentials list page.

    This test verifies that an authenticated user can access
    the credentials list page.
    """
    response = netbox_web_client.get_credentials_list()

    assert response.status_code == 200, \
        f"Expected 200 OK, got {response.status_code}: {response.text}"

    # Verify page contains expected content
    assert "credential" in response.text.lower(), \
        "Credentials list page should mention credentials"


@pytest.mark.integration
def test_add_credential_flow(netbox_web_client):
    """Test the complete flow of adding a new client credential.

    This test:
    1. Creates a new client credential
    2. Verifies redirect to secret page
    3. Checks that the secret is displayed
    """
    client_name = "pytest-test-client"

    # Step 1: Add new credential
    response = netbox_web_client.add_credential(client_name)

    # Should redirect to secret page
    assert response.status_code == 302, \
        f"Expected redirect after creating credential, got {response.status_code}"

    # Check redirect location
    assert "credentials/secret" in response.headers.get("Location", ""), \
        "Should redirect to secret page after creation"

    # Step 2: Follow redirect to secret page
    secret_url = response.headers["Location"]
    secret_response = netbox_web_client.session.get(
        f"{netbox_web_client.base_url.rstrip('netbox/')}{secret_url}"
    )

    assert secret_response.status_code == 200, \
        f"Expected 200 OK on secret page, got {secret_response.status_code}"

    # Step 3: Verify secret is displayed
    assert "client_secret" in secret_response.text.lower() or "secret" in secret_response.text.lower(), \
        "Secret page should display the client secret"

    assert client_name in secret_response.text, \
        f"Secret page should display client name '{client_name}'"


@pytest.mark.integration
def test_add_credential_requires_client_name(netbox_web_client):
    """Test that adding a credential without client_name fails validation.

    This test verifies that form validation works correctly.
    """
    # Try to add credential without client_name
    response = netbox_web_client.add_credential(None)

    # Should NOT redirect (form validation should fail)
    assert response.status_code == 200, \
        f"Expected form validation error, got {response.status_code}"

    # Check for error message in response
    assert ("enter a name for the client credential that will be created for authentication to "
            "the diode ingestion service") in response.text.lower() or "error" in response.text.lower(), \
        "Should show validation error for missing client_name"


@pytest.mark.integration
def test_delete_credential_flow(netbox_web_client, diode_client_credential):
    """Test the complete flow of deleting a client credential.

    This test:
    1. Creates a credential (via fixture)
    2. Deletes the credential
    3. Verifies redirect to list page
    """
    client_id = diode_client_credential['client_id']

    # Step 1: Verify credential exists in list before deletion
    list_response_before = netbox_web_client.get_credentials_list()
    assert list_response_before.status_code == 200, \
        f"Failed to get credentials list: {list_response_before.status_code}"
    assert f"<td class=\"align-middle\">{client_id}</td>" in list_response_before.text, \
        f"Credential {client_id} should exist in list before deletion"

    # Step 2: Delete the credential
    delete_response = netbox_web_client.delete_credential(client_id)

    # Should redirect to credentials list page after deletion
    assert delete_response.status_code in [302, 303], \
        f"Expected redirect after deleting credential, got {delete_response.status_code}"

    # Verify redirect goes to credentials list
    location = delete_response.headers.get("Location", "")
    assert "credentials" in location, \
        f"Should redirect to credentials list page, got: {location}"

    # Step 3: Verify credential no longer exists in list after deletion
    list_response_after = netbox_web_client.get_credentials_list()
    assert list_response_after.status_code == 200, \
        f"Failed to get credentials list after deletion: {list_response_after.status_code}"
    assert f"<td class=\"align-middle\">{client_id}</td>" not in list_response_after.text, \
        f"Credential {client_id} should not exist in list after deletion"


@pytest.mark.integration
def test_unauthenticated_access_redirects_to_login(test_config):
    """Test that unauthenticated requests redirect to login page.

    This test verifies that the plugin endpoints are properly
    protected and require authentication.
    """
    import requests

    # Create a new session without authentication
    session = requests.Session()

    # Try to access settings page without authentication
    response = session.get(
        f"{test_config['netbox_url']}/plugins/diode/settings/",
        allow_redirects=False
    )

    # Should redirect to login page
    assert response.status_code == 302, \
        f"Expected redirect for unauthenticated user, got {response.status_code}"

    assert "login" in response.headers.get("Location", "").lower(), \
        "Should redirect to login page"
