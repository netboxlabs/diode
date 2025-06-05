# Deploying Diode Server

This guide will help you set up and start using the Diode server for your NetBox data ingestion needs.

## Prerequisites

Before you begin, ensure you have:

- NetBox version 4.2.3 or later
- Docker version 27.0.3 or newer
- bash 4.x or newer
- jq

## Installation Steps

We provide a `quickstart.sh` script to automate the setup process. The script will download and configure all necessary files:

- `docker-compose.yaml` — Defines Diode server containers
- `.env` — Environment settings for customization
- `nginx.conf` — Nginx configuration for routing Diode endpoints
- `client-credentials.json` — Defines OAuth2 clients for secure communication

1. Create a working directory:
   ```bash
   mkdir /opt/diode
   cd /opt/diode
   ```

2. Download and prepare the quickstart script:
   ```bash
   curl -sSfLo quickstart.sh https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/scripts/quickstart.sh
   chmod +x quickstart.sh
   ```

3. Run the script with your NetBox server address:
   ```bash
   ./quickstart.sh https://{YOUR_NETBOX_FQDN}
   ```
   This should have created an `.env` file for your environment.

4. (Optional) Edit the `.env` file to adjust Diode server settings:

| Variable | Description | Default |
|:---|:---|:---|
| `DIODE_NGINX_PORT` | Port for the Nginx HTTP service. | `8080` |
| `RECONCILER_RATE_LIMITER_RPS` | Rate limit (requests per second) for reconciler change set generation. | `20` |
| `RECONCILER_RATE_LIMITER_BURST` | Burst limit for reconciler operations. | `1` |
| `DIODE_TO_NETBOX_RATE_LIMITER_RPC` | Rate limit for RPC calls to NetBox. | `20` |
| `DIODE_TO_NETBOX_RATE_LIMITER_BURST` | Burst limit for RPC calls to NetBox. | `1` |
| `LOGGING_LEVEL` | Log verbosity: `DEBUG`, `INFO`, `WARN`, `ERROR`. | `INFO` |
| `LOGGING_FORMAT` | Log output format: `json` or `text`. | `json` |

4. Start the Diode server:
   ```bash
   docker compose up -d
   ```
5. Extract the `netbox-to-diode` credential. This will be needed for the Diode NetBox plugin installation:
   ```bash
   echo $(jq -r '.[] | select(.client_id == "netbox-to-diode") | .client_secret' /opt/diode/oauth2/client/client-credentials.json)
   ```
   This should return a string. Save it somewhere safe.

## Next Step

- Installing the [Diode NetBox plugin](https://github.com/netboxlabs/diode-netbox-plugin/blob/develop/README.md)

## Managing the Diode Server

### Stopping Diode

To stop the Diode server:
```bash
docker compose down
```

## Troubleshooting

### Common Issues

1. **Connection Issues**
   - Verify network connectivity between Diode and NetBox
   - Check firewall rules
   - Validate URLs and ports

### Getting Help

If you encounter issues:

1. Search GitHub: [Issues](https://github.com/netboxlabs/diode/issues)
2. Find us in Slack: [NetDev Community #orb](https://https://netdev-community.slack.com/)