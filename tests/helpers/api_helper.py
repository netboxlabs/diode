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

    def get_sites(self, name: Optional[str] = None) -> requests.Response:
        """Get sites from NetBox.

        Args:
            name: Optional site name to filter by

        Returns:
            Response object containing site data
        """
        url = f"{self.base_url}/api/dcim/sites/"
        params = {"name": name} if name else {}
        return self.session.get(url, params=params)


    def close(self):
        """Close the session."""
        self.session.close()


class NetBoxPluginWebClient:
    """Client for interacting with NetBox plugin web endpoints (not API).

    This client handles Django session authentication and CSRF tokens
    required for accessing plugin web views.
    """

    def __init__(
        self,
        base_url: str = "http://localhost:8000/netbox/",
        username: str = "admin",
        password: str = "admin"
    ):
        """Initialize the NetBox plugin web client.

        Args:
            base_url: Base URL for NetBox
            username: Username for authentication
            password: Password for authentication
        """
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.session = requests.Session()
        self._authenticated = False

    def _get_csrf_token(self) -> str:
        """Get CSRF token from login page.

        Returns:
            CSRF token string
        """
        response = self.session.get(f"{self.base_url}/login/")
        response.raise_for_status()

        # Extract CSRF token from cookies
        csrf_token = self.session.cookies.get("csrftoken")
        if not csrf_token:
            raise ValueError("Could not retrieve CSRF token from login page")

        return csrf_token

    def login(self) -> bool:
        """Authenticate with NetBox using username and password.

        Returns:
            True if login successful, False otherwise

        Raises:
            requests.HTTPError: If login request fails
        """
        if self._authenticated:
            return True

        # Get CSRF token
        csrf_token = self._get_csrf_token()

        # Perform login
        login_data = {
            "username": self.username,
            "password": self.password,
            "csrfmiddlewaretoken": csrf_token,
        }

        response = self.session.post(
            f"{self.base_url}/login/",
            data=login_data,
            headers={"Referer": f"{self.base_url}/login/"},
            allow_redirects=False
        )

        # Check if login was successful (redirect on success)
        if response.status_code in [302, 303]:
            self._authenticated = True
            return True

        return False

    def get(self, path: str, **kwargs) -> requests.Response:
        """Make authenticated GET request to plugin endpoint.

        Args:
            path: Path to endpoint (e.g., "/plugins/diode/settings/")
            **kwargs: Additional arguments to pass to requests.get()

        Returns:
            Response object

        Raises:
            ValueError: If not authenticated
        """
        if not self._authenticated:
            if not self.login():
                raise ValueError("Failed to authenticate with NetBox")

        url = f"{self.base_url}{path}"
        return self.session.get(url, **kwargs)

    def post(self, path: str, data: Optional[dict] = None, **kwargs) -> requests.Response:
        """Make authenticated POST request to plugin endpoint.

        Args:
            path: Path to endpoint
            data: POST data
            **kwargs: Additional arguments to pass to requests.post()

        Returns:
            Response object

        Raises:
            ValueError: If not authenticated
        """
        if not self._authenticated:
            if not self.login():
                raise ValueError("Failed to authenticate with NetBox")

        # Get fresh CSRF token from cookies
        csrf_token = self.session.cookies.get("csrftoken")
        if not csrf_token:
            raise ValueError("No CSRF token available")

        # Prepare data with CSRF token
        post_data = data.copy() if data else {}
        post_data["csrfmiddlewaretoken"] = csrf_token

        url = f"{self.base_url}{path}"
        headers = kwargs.pop("headers", {})
        headers["Referer"] = url

        return self.session.post(url, data=post_data, headers=headers, **kwargs)

    def get_settings(self) -> requests.Response:
        """Get Diode plugin settings page.

        Returns:
            Response object with settings page HTML
        """
        return self.get("/plugins/diode/settings/")

    def get_credentials_list(self) -> requests.Response:
        """Get client credentials list page.

        Returns:
            Response object with credentials list page HTML
        """
        return self.get("/plugins/diode/credentials/")

    def add_credential(self, client_name: str) -> requests.Response:
        """Add a new client credential.

        Args:
            client_name: Name for the new client

        Returns:
            Response object (should redirect to secret page)
        """
        return self.post(
            "/plugins/diode/credentials/add/",
            data={"client_name": client_name},
            allow_redirects=False
        )

    def delete_credential(self, client_credential_id: str, confirm: bool = True) -> requests.Response:
        """Delete a client credential.

        Args:
            client_credential_id: ID of the credential to delete
            confirm: Whether to confirm deletion

        Returns:
            Response object (should redirect to credentials list)
        """
        return self.post(
            f"/plugins/diode/credentials/delete/{client_credential_id}/",
            data={"confirm": confirm},
            allow_redirects=False
        )

    def close(self):
        """Close the session."""
        self.session.close()
