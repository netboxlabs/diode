# Tests

This directory contains integration tests for the Diode project, using pytest.

## Prerequisites

To run the tests, you'll need:

- Python 3.9+
- A running Diode server
- A running NetBox instance with the Diode plugin installed

## Setup

### 1. Configure NetBox Connection

If you're using an external NetBox instance (not running locally on default port), you need to configure the connection.

#### 1.1. Copy the environment template

From the **root of the repository**:

```bash
cp tests/.env.example tests/.env
```

#### 1.2. Edit `tests/.env` with your NetBox details

**Required configurations:**

- **NETBOX_URL**: URL of your NetBox instance (e.g., `http://my-netbox-server:8000`)
- **NETBOX_API_TOKEN**: NetBox API token with appropriate permissions
- **NETBOX_USERNAME**: NetBox web UI username (for plugin UI tests)
- **NETBOX_PASSWORD**: NetBox web UI password (for plugin UI tests)

**Diode admin credentials (required for most tests):**

- **DIODE_ADMIN_CLIENT_ID**: Admin client ID from Diode
- **DIODE_ADMIN_CLIENT_SECRET**: Admin client secret from Diode

**Optional (only if using non-default values):**

- **DIODE_TARGET**: Diode gRPC server URL (default: `grpc://localhost:8080/diode`)

### 2. Obtain Diode Admin Credentials

To get the admin credentials for Diode:

1. Log into your NetBox instance
2. Navigate to **Plugins** → **Diode**
3. Go to **OAuth2 Clients**
4. Create a new client with scope: `diode:read diode:write`
5. Copy the `client_id` and `client_secret` to your `tests/.env` file

**Note**: If you don't configure admin credentials, tests that require them will be automatically skipped.

## Running Tests

Run all tests:
```bash
pytest tests/
```

Run specific test files:
```bash
pytest tests/test_ingestion.py
```

Run tests with verbose output:
```bash
pytest tests/ -v
```

Run tests with coverage:
```bash
pytest tests/ --cov=diode
```

## Test Structure

- `tests/`: Integration tests for the Diode SDK and NetBox plugin
- `tests/.env.example`: Template for test configuration
- `tests/.env`: Your local configuration (not tracked in git)