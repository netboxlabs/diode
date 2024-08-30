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

Diode requires a configuration file and an environment file to execute successfully:

* `docker-compose.yaml` - to configure and run the Diode server containers
* `.env` - to store the specific environmental settings

We recommend placing both files in a clean directory:

```bash
mkdir /opt/diode
cd /opt/diode
```

Download the default `docker-compose.yaml` and `.env` files from this repository:

```bash
curl -o docker-compose.yaml https://raw.githubusercontent.com/netboxlabs/diode/develop/diode-server/docker/docker-compose.yaml
curl -o .env https://raw.githubusercontent.com/netboxlabs/diode/develop/diode-server/docker/sample.env
```

Edit the `.env` to match your environment:

* `NETBOX_DIODE_PLUGIN_API_BASE_URL`: URL for the Diode NetBox plugin API
* `DIODE_TO_NETBOX_API_KEY`: API key generated with the Diode NetBox plugin installation
* `DIODE_API_KEY`: API key generated with the Diode NetBox plugin installation
* `NETBOX_TO_DIODE_API_KEY`: API key generated with the Diode NetBox plugin installation
* `INGESTER_TO_RECONCILER_API_KEY`: API key to authorize RPC calls between the Ingester and Reconciler services (at
  least 40 characters, example generation with shell command: `openssl rand -base64 40 | head -c 40`)

### Running the Diode server

Start the Diode server:

```bash
docker compose -f docker-compose.yaml up -d
```

## License

Distributed under the PolyForm Shield License 1.0.0 License. See [LICENSE.md](./LICENSE.md) for more information.
