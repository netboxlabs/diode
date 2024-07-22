# Diode

Diode is a NetBox data ingestion service that greatly simplifies and enhances the process to add and update network data in NetBox, ensuring your network source of truth is always accurate and can be trusted to power your network automation pipelines. Our guiding principle in designing Diode has been to make it as easy as possible to get data into NetBox, removing as much burden as possible from the user while shifting that effort to technology. 

To achieve this, Diode sits in front of NetBox and provides an API purpose built for ingestion of complex network data. Diode eliminates the need to preprocess data to make it conform to the strict object hierarchy imposed by the NetBox data model. This allows data to be sent to NetBox in a more freeform manner, in blocks that are intuitive for network engineers (such as by device or by interface) with much of the related information treated as attributes or properties of these components of interest. Then, Diode takes care of the heavy lifting, automatically transforming the data to align it with NetBox’s structured and comprehensive data model. Diode can even create placeholder objects to compensate for missing information, which means even fragmented information about the network can be captured in NetBox. 

## Project status

The Diode project is currently in the _Public Preview_ stage. Please see [NetBox Labs Product and Feature Lifecycle](https://docs.netboxlabs.com/product_feature_lifecycle/) for more details. We actively welcome feedback to help identify and prioritize bugs, new features and areas of improvement. 

## Get started

Diode runs as a sidecar service to NetBox and can run anywhere with network connectivity to NetBox, whether on the same host or elsewhere. The overall Diode service is delivered through three main components (and a fourth optional component):

1. Diode plugin - see how to [install the Diode plugin](https://github.com/netboxlabs/diode-netbox-plugin)
2. Diode server - see how to [run the Diode server](https://github.com/netboxlabs/diode/tree/develop/diode-server#readme)
3. Diode SDK - see how to [install the Diode client SDK](https://github.com/netboxlabs/diode-sdk-python) and [download Diode Python script examples](https://github.com/netboxlabs/netbox-learning/tree/develop/diode)
4. Diode agent (optional) - see how to [install and run the Diode NAPALM discovery agent](https://github.com/netboxlabs/diode-agent/tree/develop/diode-napalm-agent)

## Related Projects

- [diode-netbox-plugin](https://github.com/netboxlabs/diode-netbox-plugin) - The Diode NetBox plugin is a NetBox plugin and a required component of the Diode ingestion service.
- [diode-sdk-python](https://github.com/netboxlabs/diode-sdk-python) - Diode SDK Python is a Python library for interacting with the Diode ingestion service utilizing gRPC.
- [diode-agent](https://github.com/netboxlabs/diode-agent) - A collection of agents that leverage the Diode SDK to interact with the Diode server.

## License

Distributed under the PolyForm Shield License 1.0.0 License. See [LICENSE.md](./LICENSE.md) for more information.

Diode protocol buffers are distributed under the Apache 2.0 License. See [LICENSE.txt](./diode-proto/LICENSE.txt) for more information.

## Required Notice

Copyright NetBox Labs, Inc.

