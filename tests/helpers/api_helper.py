"""API helper functions for testing Diode server endpoints."""
import requests
from typing import Any, Optional, List
from netboxlabs.diode.sdk.client import DiodeClient
from netboxlabs.diode.sdk.ingester import Entity


class DiodeAPIClient:
    """Client for interacting with Diode server via gRPC."""

    def __init__(
        self,
        target: str = "grpc://localhost:8080/diode",
        name: str = "diode-test-client",
        client_id: str = None,
        client_secret: str = None,
    ):
        """Initialize the Diode gRPC client.

        Args:
            target: gRPC target URL (e.g., "grpc://localhost:8080/diode")
            name: Client name
            client_id: Client ID for authentication (required)
            client_secret: Client secret for authentication (required)

        Raises:
            ValueError: If client_id or client_secret is not provided
        """
        if not client_id or not client_secret:
            raise ValueError(
                "client_id and client_secret are required. "
                "Please set DIODE_ADMIN_CLIENT_ID and DIODE_ADMIN_CLIENT_SECRET "
                "environment variables, or ensure the test fixtures generate credentials."
            )

        self.target = target
        self.name = name
        self.client_id = client_id
        self.client_secret = client_secret
        self._client = None

    def _get_client(self) -> DiodeClient:
        """Get or create a Diode client instance."""
        if self._client is None:
            self._client = DiodeClient(
                target=self.target,
                app_name=self.name,
                app_version="1.0.0",
                client_id=self.client_id,
                client_secret=self.client_secret
            )
        return self._client

    def ingest_entities(
        self,
        entities: List[Entity],
        stream: str = "latest"
    ) -> Any:
        """Ingest entities to Diode.

        Args:
            entities: List of Entity objects to ingest
            stream: Stream name (default: "latest")

        Returns:
            IngestResponse from the Diode server
        """
        client = self._get_client()
        return client.ingest(entities=entities, stream=stream)

    def close(self):
        """Close the gRPC client connection."""
        if self._client is not None:
            self._client.close()
            self._client = None


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
