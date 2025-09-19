# Diode

Diode is a data ingestion service for NetBox that greatly simplifies and enhances the process of adding and updating data in NetBox, ensuring your network source of truth is always accurate and up to date. Our guiding principle in designing Diode has been to make it as easy as possible to get data into NetBox, removing as much burden as possible from the user while shifting that effort to technology.

## Project Status

The Diode project is currently in the _General Availability (GA)_ stage. Please see [NetBox Labs Product and Feature Lifecycle](https://netboxlabs.com/docs/console/product_feature_lifecycle/) for more details. We actively welcome feedback to help identify and prioritize bugs, new features and areas of improvement.

## Prerequisites

- NetBox 4.2.3 or later
- Python 3.8 or later (for Python SDK)
- Go 1.18 or later (for Go SDK)
- Network connectivity between Diode and NetBox

## Quick Start

For a quick step-by-step guide, see our [Getting Started Guide](./GET_STARTED.md).

1. **Deploy the Diode Server**
   See [deployment instructions](https://github.com/netboxlabs/diode/blob/develop/diode-server/README.md)

2. **Install the Diode NetBox Plugin**
   See [installation instructions](https://github.com/netboxlabs/diode-netbox-plugin/blob/develop/README.md)

3. **Choose Your Data Ingestion Method**
   - Use the NetBox Discovery agent for automated network discovery: see [instructions to run Orb agent](https://github.com/netboxlabs/orb-agent)
   - Build custom integrations using our SDKs:
     - [Python SDK](https://github.com/netboxlabs/diode-sdk-python)
     - [Go SDK](https://github.com/netboxlabs/diode-sdk-go)

## Documentation

- [Diode Protocol Documentation](https://github.com/netboxlabs/diode/blob/develop/docs/diode-proto.md)
- [Metrics](https://github.com/netboxlabs/diode/blob/develop/docs/metrics.md)

## Related Projects

- [diode-netbox-plugin](https://github.com/netboxlabs/diode-netbox-plugin) - The Diode NetBox plugin is a NetBox plugin and a required component of the Diode ingestion service.
- [diode-sdk-python](https://github.com/netboxlabs/diode-sdk-python) - Diode SDK Python is a Python library for interacting with the Diode ingestion service utilizing gRPC.
- [diode-sdk-go](https://github.com/netboxlabs/diode-sdk-go) - Diode SDK Go is a Go module for interacting with the Diode ingestion service utilizing gRPC.
- [orb-agent](https://github.com/netboxlabs/orb-agent) - The NetBox Discovery agent.

## Support

- [GitHub Issues](https://github.com/netboxlabs/diode/issues)
- [Slack NetDev Community (#orb channel)](https://https://netdev-community.slack.com/)

## License

Distributed under the NetBox Limited Use License 1.0. See [LICENSE.md](./LICENSE.md) for more information.

Diode protocol buffers are distributed under the Apache 2.0 License. See [LICENSE.txt](./diode-proto/LICENSE.txt) for more information.

## Required Notice

Copyright NetBox Labs, Inc.

