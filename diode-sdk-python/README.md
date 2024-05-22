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
* `DIODE_TLS_VERIFY` - Verify TLS certificate
* `DIODE_SDK_LOG_LEVEL` - Log level for the SDK (default: `INFO`)
* `DIODE_SENTRY_DSN` - Optional Sentry DSN for error reporting

### Examples

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Device,
    DeviceType,
    Entity,
    Interface,
    IPAddress,
    Manufacturer,
    Platform,
    Prefix,
    Role,
    Site,
)


def main():
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        tls_verify=False,
    ) as client:
        entities = []

        # device - simplified definition
        entities.append(
            Entity(
                device="Device A",
            )
        )

        # device - expanded definition (1)
        entities.append(
            Entity(
                device=Device(
                    name="Device A",
                    device_type="Device Type A",
                    platform="Platform A",
                    manufacturer="Manufacturer A",
                    site="Site ABC",
                    role="Role ABC",
                    serial="123456",
                    asset_tag="123456",
                    status="active",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # device - expanded definition (2)
        entities.append(
            Entity(
                device=Device(
                    name="Device A",
                    device_type=DeviceType(
                        model="Device Type A", manufacturer="Manufacturer A"
                    ),
                    platform=Platform(name="Platform A", manufacturer="Manufacturer A"),
                    site=Site(name="Site ABC"),
                    role=Role(name="Role ABC", tags=["tag 1", "tag 3"]),
                    serial="123456",
                    asset_tag="123456",
                    status="active",
                    comments="Lorem ipsum dolor sit amet",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # device type - simplified definition
        entities.append(
            Entity(
                device_type="Device Type A",
            )
        )

        # device type - expanded definition
        entities.append(
            Entity(
                device_type=DeviceType(
                    model="Device Type A",
                    manufacturer="Manufacturer A",
                    part_number="123456",
                    description="Device Type A description",
                    comments="Lorem ipsum dolor sit amet",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # interface - simplified definition
        entities.append(
            Entity(
                interface="Interface A",
            )
        )

        # interface - expanded definition (1)
        entities.append(
            Entity(
                interface=Interface(
                    name="Interface A",
                    device="Device A",
                    device_type="Device Type A",
                    role="Role ABC",
                    platform="Platform A",
                    site="Site ABC",
                    type="virtual",
                    enabled=True,
                    mtu=1500,
                    mac_address="00:00:00:00:00:00",
                    description="Interface A description",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # interface - expanded definition (2)
        entities.append(
            Entity(
                interface=Interface(
                    name="Interface A",
                    device=Device(
                        name="Device A",
                        device_type=DeviceType(
                            model="Device Type A", manufacturer="Manufacturer A"
                        ),
                        platform=Platform(
                            name="Platform A", manufacturer="Manufacturer A"
                        ),
                        site=Site(name="Site ABC"),
                        role=Role(name="Role ABC", tags=["tag 1", "tag 3"]),
                    ),
                    type="virtual",
                    enabled=True,
                    mtu=1500,
                    mac_address="00:00:00:00:00:00",
                    description="Interface A description",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # ip address - simplified definition
        entities.append(
            Entity(
                ip_address="192.168.0.1/24",
            )
        )

        # ip address - expanded definition (1)
        entities.append(
            Entity(
                ip_address=IPAddress(
                    address="192.168.0.1/24",
                    interface="Interface ABC",
                    device="Device ABC",
                    device_type="Device Type ABC",
                    device_role="Role ABC",
                    platform="Platform ABC",
                    manufacturer="Cisco",
                    site="Site ABC",
                    status="active",
                    role="Role ABC",
                    description="IP Address A description",
                    comments="Lorem ipsum dolor sit amet",
                    tags=["tag1", "tag2"],
                )
            )
        )

        # ip address - expanded definition (2)
        entities.append(
            Entity(
                ip_address=IPAddress(
                    address="192.168.0.1/24",
                    interface=Interface(
                        name="Interface ABC",
                        device=Device(
                            name="Device ABC",
                            device_type=DeviceType(
                                model="Device Type ABC", manufacturer="Cisco"
                            ),
                            platform=Platform(
                                name="Platform ABC", manufacturer="Cisco"
                            ),
                            site=Site(name="Site ABC"),
                            role=Role(name="Role ABC", tags=["tag1", "tag3"]),
                        ),
                    ),
                    status="active",
                    role="Role ABC",
                    description="IP Address A description",
                    comments="Lorem ipsum dolor sit amet",
                    tags=["tag1", "tag2"],
                )
            )
        )

        # manufacturer - simplified definition
        entities.append(
            Entity(
                manufacturer="Manufacturer A",
            )
        )

        # manufacturer - expanded definition
        entities.append(
            Entity(
                manufacturer=Manufacturer(
                    name="Manufacturer A",
                    description="Manufacturer A description",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # platform - simplified definition
        entities.append(
            Entity(
                platform="Platform A",
            )
        )

        # platform - expanded definition (1)
        entities.append(
            Entity(
                platform=Platform(
                    name="Platform A",
                    manufacturer="Manufacturer A",
                    description="Platform A description",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # platform - expanded definition (2)
        entities.append(
            Entity(
                platform=Platform(
                    name="Platform A",
                    manufacturer=Manufacturer(
                        name="Manufacturer A", tags=["tag 1", "tag 3"]
                    ),
                    description="Platform A description",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        # prefix - simplified definition
        entities.append(Entity(prefix="192.168.0.0/32"))

        # prefix - expanded definition (1)
        entities.append(
            Entity(
                prefix=Prefix(
                    prefix="192.168.0.0/32",
                    site="Site ABC",
                    is_pool=True,
                )
            )
        )

        # prefix - expanded definition (2)
        entities.append(
            Entity(
                prefix=Prefix(
                    prefix="192.168.0.0/32",
                    site=Site(
                        name="Site ABC",
                        status="active",
                        tags=["tag 1", "tag 2"],
                    ),
                    is_pool=True,
                )
            )
        )

        # site - simplified definition
        entities.append(Entity(site="Site A"))

        # site - expanded definition
        entities.append(
            Entity(
                site=Site(
                    name="Site A",
                    status="active",
                    facility="Data Center 1",
                    time_zone="UTC",
                    description="Site A description",
                    comments="Lorem ipsum dolor sit amet",
                    tags=["tag 1", "tag 2"],
                )
            )
        )

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```
