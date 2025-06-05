# Diode Server

The Diode server is a required component of the [Diode](https://github.com/netboxlabs/diode) ingestion service.

Diode is a data ingestion service for NetBox that greatly simplifies and enhances the process of adding and updating data in NetBox, ensuring your network source of truth is always accurate and up to date. Our guiding principle in designing Diode has been to make it as easy as possible to get data into NetBox, removing as much burden as possible from the user while shifting that effort to technology.

---

## Overview of Diode Services

The Diode server is composed of three core services:

### Auth Service

- Issues and introspects OAuth2 tokens

### Ingester Service

- Accepts and pushes ingested data into Redis streams for further processing

### Reconciler Service

- Processes and reconciles ingested data against existing NetBox objects, detecting and storing any changes

---

## Compatibility

The Diode server has been tested with NetBox version **4.2.3**.  
It also requires the [Diode NetBox Plugin](https://github.com/netboxlabs/diode-netbox-plugin) **1.1.0**.

---

## Getting Started

### Requirements

- Docker version **27.0.3** or newer
- bash 4.x or newer
- jq

### Quick Installation

We provide a `quickstart.sh` script to automate the setup process.

The following files will be downloaded:

- `docker-compose.yaml` — Defines Diode server containers
- `.env` — Environment settings for customization
- `nginx.conf` — Nginx configuration for routing Diode endpoints
- `client-credentials.json` — Defines OAuth2 clients for secure communication between the Orb Agent, Diode SDK, Diode server, and Diode NetBox plugin

We recommend placing these files in a clean working directory, for example:

```bash
mkdir /opt/diode
cd /opt/diode
```

Download and prepare the quickstart script:

```bash
curl -sSfLo quickstart.sh https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/scripts/quickstart.sh
chmod +x quickstart.sh
```

Run the script to download and configure required files with your NetBox server address:

```bash
./quickstart.sh http://my.netbox:8080
```

### Starting Diode

Once setup is complete, start the Diode server:

```bash
docker compose up -d
```

---

## Configuration

Edit the `.env` file to adjust Diode server settings as needed:

| Variable | Description | Default |
|:---|:---|:---|
| `DIODE_NGINX_PORT` | Port for the Nginx HTTP service. | `8080` |
| `RECONCILER_RATE_LIMITER_RPS` | Rate limit (requests per second) for reconciler change set generation. | `20` |
| `RECONCILER_RATE_LIMITER_BURST` | Burst limit for reconciler operations. | `1` |
| `DIODE_TO_NETBOX_RATE_LIMITER_RPC` | Rate limit for RPC calls to NetBox. | `20` |
| `DIODE_TO_NETBOX_RATE_LIMITER_BURST` | Burst limit for RPC calls to NetBox. | `1` |
| `LOGGING_LEVEL` | Log verbosity: `DEBUG`, `INFO`, `WARN`, `ERROR`. | `INFO` |
| `LOGGING_FORMAT` | Log output format: `json` or `text`. | `json` |

---

### Stopping Diode

To stop the Diode:

```bash
docker compose down
```

To stop the Diode and also delete PostgeSQL and Redis volumes:

```bash
docker compose down --volumes
```

---

### Provisioning and Managing Agent Credentials via Command Line

Additional agent credentials may be provisioned by calling the `authmanager` command in the auth service container.

**Create a new client with the right to ingest**
```bash
docker compose run --rm --no-deps diode-auth authmanager create-client --client-id my-agent-001 --allow-ingest

** NOTE: The client secret is only displayed once and cannot be retrieved later.
  Store credentials in a secure location. If you lose them, you will need to
  destroy and regenerate the client.
{
  "client_id": "my-agent-001",
  "scope": "diode:ingest",
  "client_secret": "a_new_generated_secret"
}
```

**Create a client with a supplied secret** (in this case the secret is not returned)
```bash
docker compose run --rm --no-deps diode-auth authmanager create-client --client-id my-agent-002 --allow-ingest --client-secret="a_secret_key_from_some_other_source"

client created successfully.
{
  "client_id": "my-agent-002",
  "scope": "diode:ingest"
}
```

**Retrieve info of an existing client**
```bash
docker compose run --rm --no-deps diode-auth authmanager get-client --client-id my-agent-001

{
  "client_id": "my-agent-001",
  "scope": "diode:ingest"
}
```

**List existing clients**
```bash
docker compose run --rm --no-deps diode-auth authmanager list-clients

[
  {
    "client_id": "diode-ingest",
    "scope": "diode:ingest"
  },
  {
    "client_id": "diode-to-netbox",
    "scope": "netbox:read netbox:write"
  },
  {
    "client_id": "my-agent-001",
    "scope": "diode:ingest"
  },
  {
    "client_id": "my-agent-002",
    "scope": "diode:ingest"
  },
  {
    "client_id": "netbox-to-diode",
    "scope": "diode:read diode:write"
  }
]
```

**Delete an existing client**
```bash
docker compose run --rm --no-deps diode-auth authmanager delete-client --client-id my-agent-002

client my-agent-002 deleted successfully
```

**List auth utility subcommands**
```bash
docker compose run --rm --no-deps diode-auth authmanager

usage: authmanager <subcommand>
subcommands: list-clients get-client delete-client create-client
```

**Additional help on a subcommand**
```bash
docker compose run --rm --no-deps diode-auth authmanager create-client --help

Usage of create-client:
  -allow-ingest
    	include scopes that allow the client to ingest data
  -client-id string
    	client id
  -client-secret string
    	client secret [generated if not provided]
  -scope string
    	space separated list of scopes to allow
```
---

## License

Distributed under the NetBox Limited Use License 1.0.  
See [LICENSE.md](../LICENSE.md) for more information.
