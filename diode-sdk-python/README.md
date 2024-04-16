# Diode SDK Python

## Installation

```bash
pip install netboxlabs-diode-sdk
```

## Development notes

```python
ruff netboxlabs/
black netboxlabs/
```

## Usage

### Environment variables

* `DIODE_API_KEY` - API key for the Diode service
* `DIODE_SDK_LOG_LEVEL` - Log level for the SDK (default: `INFO`)

### Example

```python

from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.diode.v1.device_type_pb2 import DeviceType
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.manufacturer_pb2 import Manufacturer
from netboxlabs.diode.sdk.diode.v1.site_pb2 import Site


def main():
    with DiodeClient(target="localhost:8081", app_name="my-test-app", app_version="0.0.1") as client:
        entities = [
            Entity(site=Site(name="Site 1")),
            Entity(
                device_type=DeviceType(
                    model="ISR4321",
                    manufacturer=Manufacturer(name="Cisco"),
                )
            ),
        ]

        resp = client.ingest(entities=entities)
        print(f"Response: {resp}")

if __name__ == "__main__":
    main()
```
