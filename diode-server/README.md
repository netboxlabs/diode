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
* `client-credentials.json` - to create OAuth2 clients required for communication between Orb Agent / Diode SDK, diode server and Diode NetBox plugins

We recommend placing both files in a clean directory:

```bash
mkdir /opt/diode
cd /opt/diode
```

Download the default `docker-compose.yaml` and other required files from this repository:

```bash
curl -o docker-compose.yaml https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/docker-compose.yaml
curl -o .env https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/sample.env
curl -o generate-client-credentials.sh https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/scripts/generate-client-credentials.sh
curl -o generate-env-secrets.sh https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/scripts/generate-env-secrets.sh
```

Generate OAuth2 client credentials into `client-credentials.json` file:

```bash
chmod +x generate-client-credentials.sh
mkdir oauth2/client
./generate-client-credentials.sh > ./oauth2/client/client-credentials.json
```

Generate secrets and replace placeholders in `.env`:

```bash
chmod +x generate-env-secrets.sh
./generate-env-secrets.sh .env
```

Set `DIODE_TO_NETBOX_CLIENT_SECRET` in `.env` extracted from generated `client-credentials.json` file:

```bash
DIODE_TO_NETBOX_CLIENT_SECRET=$(jq -r '.[] | select(.client_id == "diode-to-netbox") | .client_secret' ./oauth2/client/client-credentials.json)

# linux
sed -i "s|<PLACEHOLDER_DIODE_TO_NETBOX_CLIENT_SECRET>|$DIODE_TO_NETBOX_CLIENT_SECRET|g" .env

# macos
sed -i '' "s|<PLACEHOLDER_DIODE_TO_NETBOX_CLIENT_SECRET>|$DIODE_TO_NETBOX_CLIENT_SECRET|g" .env
```


Edit the `.env` to match your environment:

* `DIODE_NGINX_PORT`: Port number for the Nginx service that handles incoming HTTP requests, default: `8080`
* `NETBOX_DIODE_PLUGIN_API_BASE_URL`: URL for the Diode NetBox plugin API, replace `<http://NETBOX_HOST>` with your NetBox URL
* `RECONCILER_RATE_LIMITER_RPS`: Rate limit for the reconciler service for generating and applying change sets concurrently, default: `20`
* `RECONCILER_RATE_LIMITER_BURST`: Burst limit for the reconciler service for generating and applying change sets concurrently, default: `1`
* `DIODE_TO_NETBOX_RATE_LIMITER_RPC`: Rate limit for the number of RPC calls per second from Diode to NetBox, default: `20`
* `DIODE_TO_NETBOX_RATE_LIMITER_BURST`: Burst limit for the number of RPC calls from Diode to NetBox, default: `1`
* `LOGGING_LEVEL`: Controls the verbosity of logs, options include: `DEBUG`, `INFO`, `WARN`, `ERROR`, default: `INFO`
* `LOGGING_FORMAT`: Controls the format of log output, options include: `json`, `text`, default: `json`

### Running the Diode server

Start the Diode server:

```bash
docker compose -f docker-compose.yaml up -d
```

## License

Distributed under the PolyForm Shield License 1.0.0 License. See [LICENSE.md](./LICENSE.md) for more information.
