"""API helper functions for testing Diode server endpoints."""
import requests
from typing import Dict, Any, Optional


class DiodeAPIClient:
    """Client for interacting with Diode server API."""

    def __init__(self, base_url: str = "http://localhost:8081", timeout: int = 30):
        """Initialize the API client.

        Args:
            base_url: Base URL for the Diode server
            timeout: Request timeout in seconds
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.session = requests.Session()

    def health_check(self) -> requests.Response:
        """Check if the Diode server is healthy.

        Returns:
            Response object from the health check endpoint
        """
        url = f"{self.base_url}/health"
        return self.session.get(url, timeout=self.timeout)

    def ingest_entity(
        self,
        entity_data: Dict[str, Any],
        api_key: Optional[str] = None
    ) -> requests.Response:
        """Send entity data to Diode for ingestion.

        Args:
            entity_data: Entity data to ingest
            api_key: Optional API key for authentication

        Returns:
            Response object from the ingestion endpoint
        """
        url = f"{self.base_url}/diode/v1/ingest"
        headers = {"Content-Type": "application/json"}
        if api_key:
            headers["Authorization"] = f"Bearer {api_key}"

        return self.session.post(
            url,
            json=entity_data,
            headers=headers,
            timeout=self.timeout
        )

    def close(self):
        """Close the session."""
        self.session.close()


class NetBoxAPIClient:
    """Client for interacting with NetBox API."""

    def __init__(self, base_url: str = "http://localhost:8000", token: Optional[str] = None):
        """Initialize the NetBox API client.

        Args:
            base_url: Base URL for NetBox
            token: API token for authentication
        """
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.session = requests.Session()
        if token:
            self.session.headers.update({"Authorization": f"Token {token}"})

    def get_device(self, device_name: str) -> requests.Response:
        """Get a device from NetBox by name.

        Args:
            device_name: Name of the device

        Returns:
            Response object containing device data
        """
        url = f"{self.base_url}/api/dcim/devices/"
        params = {"name": device_name}
        return self.session.get(url, params=params)

    def close(self):
        """Close the session."""
        self.session.close()
