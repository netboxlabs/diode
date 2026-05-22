# Getting Started with Diode

This guide will help you set up and start using Diode to ingest data into NetBox.

## Prerequisites

Before you begin, ensure you have:

- NetBox version 4.2.3 or later
- Docker version 27.0.3 or newer
- bash 4.x or newer
- jq

## Installation Steps

### Deploy Diode server

> **Host**: These steps should be performed on the host where you want to run the Diode server.

> **Note**: For the complete installation instructions, please refer to the [official Diode Server documentation](https://github.com/netboxlabs/diode/tree/develop/diode-server).

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
   ./quickstart.sh https://<netbox-server>
   ```
   This should have created an `.env` file for your environment.

4. Start the Diode server:
   ```bash
   docker compose up -d
   ```
5. Extract the `netbox-to-diode` client secret. This will be needed for the Diode NetBox plugin installation:
   ```bash
   echo $(jq -r '.[] | select(.client_id == "netbox-to-diode") | .client_secret' /opt/diode/oauth2/client/client-credentials.json)
   ```
   > **Note**: This will return a credential that will be used by the Diode NetBox plugin to connect to the Diode server. Store it safely.

### Install Diode NetBox Plugin

> **Host**: These steps should be performed on the host where NetBox is installed.

> **Note**: For the complete installation instructions, please refer to the [official Diode NetBox Plugin documentation](https://github.com/netboxlabs/diode-netbox-plugin/blob/develop/README.md).

1. **Source the NetBox Python Virtual Environment**
   ```bash
   cd /opt/netbox
   source venv/bin/activate
   ```

2. **Install the Plugin Package**
   ```bash
   pip install netboxlabs-diode-netbox-plugin
   ```

3. **Configure NetBox Settings**
   Add the following to your `configuration.py`:
   ```python
   PLUGINS = [
       "netbox_diode_plugin",
   ]

   PLUGINS_CONFIG = {
       "netbox_diode_plugin": {
           # Diode gRPC target for communication with Diode server
           "diode_target_override": "grpc://<diode-server:port>/diode",
           # NetBox username associated with changes applied via plugin
           "diode_username": "diode",
           # netbox-to-diode client secret from earlier step
           "netbox_to_diode_client_secret": "<netbox-to-diode-secret>"
       },
   }
   ```

4. **Apply Database Migrations**
   ```bash
   cd /opt/netbox/netbox
   ./manage.py migrate netbox_diode_plugin
   ```

5. **Restart NetBox Services**
   ```bash
   sudo systemctl restart netbox netbox-rq
   ```

6. **Generate Diode Client Credentials**
   > **Note**: These credentials will be used by the Orb agent to send discovery results to NetBox via Diode.

   1. Go to your NetBox instance (https://<netbox-server>)
   2. In the left-hand pane, navigate to **Diode -> Client Credentials**
   3. Click on **+ Add a Credential**
   4. For Client Name, enter any name and click **Create**
   5. **IMPORTANT**: Copy the _Client ID_ and _Client Secret_ and save them securely
   6. Click **Return to List**

   You have now created your Diode client credentials. These will be used as environment variables when running the Orb agent.

### Ingest Data with Orb Agent

> **Host**: These steps should be performed on the host where you want to run the Orb agent for network discovery.

> **Note**: For the complete installation instructions, please refer to the [official Orb Agent documentation](https://github.com/netboxlabs/orb-agent).

1. **Export Client Credentials**
   ```bash
   # Export the client credentials you generated in NetBox
   export DIODE_CLIENT_ID="<your-client-id>"
   export DIODE_CLIENT_SECRET="<your-client-secret>"
   ```

2. **Create Agent Configuration File**
   Create an `agent.yaml` file with the following content:
   ```yaml
   orb:
     config_manager:
       active: local
     backends:
       network_discovery:  # Enable network discovery backend
       common:
         diode:
           target: grpc://<diode-server:port>/diode
           client_id: ${DIODE_CLIENT_ID}
           client_secret: ${DIODE_CLIENT_SECRET}
           agent_name: my_agent
     policies:
       network_discovery:
         loopback_policy:
           config:
           scope:
             targets: 
               - 127.0.0.1
   ```

3. **Run the Agent**
   
   Using host network mode (recommended):
   ```bash
   docker run --net=host \
     -v $(pwd):/opt/orb/ \
     -e DIODE_CLIENT_ID \
     -e DIODE_CLIENT_SECRET \
     netboxlabs/orb-agent:latest run -c /opt/orb/agent.yaml
   ```

   Alternative using root user:
   ```bash
   docker run -u root \
     -v $(pwd):/opt/orb/ \
     -e DIODE_CLIENT_ID \
     -e DIODE_CLIENT_SECRET \
     netboxlabs/orb-agent:latest run -c /opt/orb/agent.yaml
   ```

   > **Note**: The container needs sufficient permissions to send ICMP and TCP packets. This can be achieved either by:
   > - Setting the network mode to `host` (recommended)
   > - Running the container as root user

4. **Verify Agent Operation**
   - Check the agent logs for successful startup
   - Verify data appears in NetBox

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
