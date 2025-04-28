# Diode server

The Diode server is a required component of the [Diode](https://github.com/netboxlabs/diode) ingestion service.

Diode is a NetBox ingestion service that greatly simplifies and enhances the process to add and update network data
in NetBox, ensuring your network source of truth is always accurate and can be trusted to power your network automation
pipelines.

More information about Diode can be found
at [https://netboxlabs.com/blog/introducing-diode-streamlining-data-ingestion-in-netbox/](https://netboxlabs.com/blog/introducing-diode-streamlining-data-ingestion-in-netbox/).

## Diode services

Diode server is composed of two services:

### Ingester Service

- Responsible for receiving and validating ingestion data.
- Utilizes `IngesterService.Ingest` RPC method.
- Supports single API key for data source authorization.
- Validates incoming data and pushes it into Redis streams.

### Reconciler Service

- Processes data from Redis streams and converts it for storage.
- Manages data sources and their API keys.
- Implements a reconciliation engine to detect and store deltas between ingested data and the current NetBox object
  state.

## Compatibility

The Diode server has been tested with NetBox versions 3.7.2 and above. The Diode server also requires
the [Diode NetBox Plugin](https://github.com/netboxlabs/diode-netbox-plugin).

## Running the Diode server

### Requirements

Diode server requires Docker version 27.0.3 or above.

### Installation

We prepared a `quickstart.sh` file to download all required Diode configuration files:

* `docker-compose.yaml` - to configure and run the Diode server containers
* `.env` - to store the specific environmental settings
* `nginx.conf` - to configure Nginx with Diode endpoints
* `client-credentials.json` - to create OAuth2 clients required for communication between Orb Agent / Diode SDK, diode server and Diode NetBox plugins

We recommend placing all files in a clean directory, e.g:

```bash
mkdir /opt/diode
cd /opt/diode
```

Download `quickstart.sh`:

```bash
curl -sSfL https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/scripts/quickstart.sh
chmod +x quickstart.sh
```

Run `quickstart.sh` with your NetBox address (e.g. `http://my.netbox:8080`):
```bash
./quickstart.sh http://my.netbox:8080
```

### Running the Diode server

Start the Diode server:

```bash
docker compose up -d
```

### Environment variables

Edit the `.env` to match your environment:

* `DIODE_NGINX_PORT`: Port number for the Nginx service that handles incoming HTTP requests, default: `8080`
* `RECONCILER_RATE_LIMITER_RPS`: Rate limit for the reconciler service for generating and applying change sets concurrently, default: `20`
* `RECONCILER_RATE_LIMITER_BURST`: Burst limit for the reconciler service for generating and applying change sets concurrently, default: `1`
* `DIODE_TO_NETBOX_RATE_LIMITER_RPC`: Rate limit for the number of RPC calls per second from Diode to NetBox, default: `20`
* `DIODE_TO_NETBOX_RATE_LIMITER_BURST`: Burst limit for the number of RPC calls from Diode to NetBox, default: `1`
* `LOGGING_LEVEL`: Controls the verbosity of logs, options include: `DEBUG`, `INFO`, `WARN`, `ERROR`, default: `INFO`
* `LOGGING_FORMAT`: Controls the format of log output, options include: `json`, `text`, default: `json`

## License

Distributed under the NetBox Limited Use License 1.0. See [LICENSE.md](../LICENSE.md) for more information.
