# Diode

Diode is a data ingestion service for NetBox that greatly simplifies and enhances the process of adding and updating data in NetBox, ensuring your network source of truth is always accurate and up to date. Our guiding principle in designing Diode has been to make it as easy as possible to get data into NetBox, removing as much burden as possible from the user while shifting that effort to technology.

## Project status

The Diode project is currently in the _Public Preview_ stage. Please see [NetBox Labs Product and Feature Lifecycle](https://netboxlabs.com/docs/console/product_feature_lifecycle/) for more details. We actively welcome feedback to help identify and prioritize bugs, new features and areas of improvement.

## Get started

Diode runs as a sidecar service to NetBox and can run anywhere with network connectivity to NetBox, whether on the same
host or elsewhere. To get started with Diode, you need to run the Diode server and install the NetBox Diode plugin:

1. Run the Diode server - see how to [run the Diode server](https://github.com/netboxlabs/diode/tree/develop/diode-server#readme)
2. Install the Diode plugin - see how to [install the Diode plugin](https://github.com/netboxlabs/diode-netbox-plugin)

To start ingesting data you'll need a Diode client:

1. Run the NetBox Discovery agent - see how to [run discovery with the Orb agent](https://github.com/netboxlabs/orb-agent)
2. Build your own client using one of the Diode SDKs - see how
   to [install the Diode Python client SDK](https://github.com/netboxlabs/diode-sdk-python), [download Diode Python script examples](https://github.com/netboxlabs/netbox-learning/tree/develop/diode)
   or [use the Diode SDK Go](https://github.com/netboxlabs/diode-sdk-go)

## Related Projects

- [diode-netbox-plugin](https://github.com/netboxlabs/diode-netbox-plugin) - The Diode NetBox plugin is a NetBox plugin
  and a required component of the Diode ingestion service.
- [diode-sdk-python](https://github.com/netboxlabs/diode-sdk-python) - Diode SDK Python is a Python library for
  interacting with the Diode ingestion service utilizing gRPC.
- [diode-sdk-go](https://github.com/netboxlabs/diode-sdk-go) - Diode SDK Go is a Go module for interacting with the
  Diode ingestion service utilizing gRPC.
- [orb-agent](https://github.com/netboxlabs/orb-agent) - The NetBox Discovery agent.

## License

Distributed under the NetBox Limited Use License 1.0. See [LICENSE.md](./LICENSE.md) for more information.

Diode protocol buffers are distributed under the Apache 2.0 License. See [LICENSE.txt](./diode-proto/LICENSE.txt) for
more information.

## Required Notice

Copyright NetBox Labs, Inc.

