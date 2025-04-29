# Diode Server

The Diode server is a required component of the [Diode](https://github.com/netboxlabs/diode) ingestion service.

Diode is a NetBox ingestion service that greatly simplifies and enhances the process to add and update network data
in NetBox, ensuring your network source of truth is always accurate and can be trusted to power your network automation
pipelines.

More information about Diode can be found
at [https://netboxlabs.com/blog/introducing-diode-streamlining-data-ingestion-in-netbox/](https://netboxlabs.com/blog/introducing-diode-streamlining-data-ingestion-in-netbox/).


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
It also requires the [Diode NetBox Plugin](https://github.com/netboxlabs/diode-netbox-plugin) **1.x.x**.

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

## License

Distributed under the NetBox Limited Use License 1.0.  
See [LICENSE.md](../LICENSE.md) for more information.
