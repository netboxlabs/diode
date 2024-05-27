# Supported Entities

## Device

Attributes:
* `name` (str) - device name
* `device_type` (str, [DeviceType](#device-type)) - device type name or DeviceType entity
* `platform` (str, [Platform](#platform)) - platform name or Platform entity
* `manufacturer` (str, [Manufacturer](#manufacturer)) - manufacturer name or Manufacturer entity
* `site` (str, [Site](#site)) - site name or Site entity
* `role` (str, [Role](#role)) - role name or Role entity
* `serial` (str) - serial number
* `asset_tag` (str) - asset tag
* `status` (str) - status (e.g. `active`, `planned`, `staged`, `failed`, `inventory`, `decommissioning`, `offine`)
* `comments` (str) - comments
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Device,
    DeviceType,
    Entity,
    Manufacturer,
    Platform,
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

        """
        Device entity with only a name provided will attempt to create or update a device with
        the given name and placeholders (i.e. "undefined") for other nested objects types
        (e.g. DeviceType, Platform, Site, Role) required by NetBox.
        """

        device = Device(name="Device A")

        entities.append(Entity(device=device))

        """
        Device entity using flat data structure.
        """

        device_flat = Device(
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

        entities.append(Entity(device=device_flat))

        """
        Device entity using explicit data structure.
        """

        device_explicit = Device(
            name="Device A",
            device_type=DeviceType(
                model="Device Type A", manufacturer=Manufacturer(name="Manufacturer A")
            ),
            platform=Platform(
                name="Platform A", manufacturer=Manufacturer(name="Manufacturer A")
            ),
            site=Site(name="Site ABC"),
            role=Role(name="Role ABC", tags=["tag 1", "tag 3"]),
            serial="123456",
            asset_tag="123456",
            status="active",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(device=device_explicit))


        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()
```

## Interface

Attributes:
* `name` (str) - interface name
* `device` (str, [Device](#device)) - device name or Device entity
* `device_type` (str, [DeviceType](#device-type)) - device type name or DeviceType entity
* `role` (str, [Role](#role)) - role name or Role entity
* `platform` (str, [Platform](#platform)) - platform name or Platform entity
* `manufacturer` (str, [Manufacturer](#manufacturer)) - manufacturer name or Manufacturer entity
* `site` (str, [Site](#site)) - site name or Site entity
* `type` (str) - interface type (e.g. `virtual`, `other`, etc.)
* `enabled` (bool) - is the interface enabled
* `mtu` (int) - maximum transmission unit
* `mac_address` (str) - MAC address
* `speed` (int) - speed
* `wwn` (str) - world wide name
* `mgmt_only` (bool) - is the interface for management only
* `description` (str) - description
* `mark_connected` (bool) - mark connected
* `mode` (str) - mode (`access`, `tagged`, `tagged-all`)
* `tags` (list) - tags

### Example

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

        """
        Interface entity with only a name provided will attempt to create or update an interface with
        the given name and placeholders (i.e. "undefined") for other nested objects types
        (e.g. Device, DeviceType, Platform, Site, Role) required by NetBox.
        """

        interface = Interface(name="Interface A")

        entities.append(Entity(interface=interface))

        """
        Interface entity using flat data structure.
        """

        interface_flat = Interface(
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

        entities.append(Entity(interface=interface_flat))

        """
        Interface entity using explicit data structure.
        """

        interface_explicit = Interface(
            name="Interface A",
            device=Device(
                name="Device A",
                device_type=DeviceType(
                    model="Device Type A",
                    manufacturer=Manufacturer(name="Manufacturer A"),
                ),
                platform=Platform(
                    name="Platform A", manufacturer=Manufacturer(name="Manufacturer A")
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

        entities.append(Entity(interface=interface_explicit))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Device Type

Attributes:
* `model` (str) - device type model
* `slug` (str) - slug
* `manufacturer` (str, [Manufacturer](#manufacturer)) - manufacturer name or Manufacturer entity
* `description` (str) - description
* `comments` (str) - comments
* `part_number` (str) - part number
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    DeviceType,
    Entity,
    Manufacturer,
)


def main():
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        tls_verify=False,
    ) as client:
        entities = []

        """
        DeviceType entity with only a name provided will attempt to create or update a device type with
        the given name and placeholder (i.e. "undefined") for nested Manufacturer object type
        required by NetBox.
        """

        device_type = DeviceType(model="Device Type A")

        entities.append(Entity(device_type=device_type))

        """
        DeviceType entity using flat data structure.
        """

        device_type_flat = DeviceType(
            model="Device Type A",
            manufacturer="Manufacturer A",
            part_number="123456",
            description="Device Type A description",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(device_type=device_type_flat))

        """
        DeviceType entity using explicit data structure.
        """

        device_type_explicit = DeviceType(
            model="Device Type A",
            manufacturer=Manufacturer(
                name="Manufacturer A",
                description="Manufacturer A description",
                tags=["tag 1", "tag 2"],
            ),
            part_number="123456",
            description="Device Type A description",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(device_type=device_type_explicit))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Platform

Attributes:
* `name` (str) - platform name
* `slug` (str) - slug
* `manufacturer` (str, [Manufacturer](#manufacturer)) - manufacturer name or Manufacturer entity
* `description` (str) - description
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Entity,
    Manufacturer,
    Platform,
)


def main():
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        tls_verify=False,
    ) as client:
        entities = []

        """
        Platform entity with only a name provided will attempt to create or update a platform with
        the given name and placeholders (i.e. "undefined") for other nested objects types (e.g. Manufacturer)
        required by NetBox.
        """

        platform = Platform(
            name="Platform A",
        )

        entities.append(Entity(platform=platform))

        """
        Platform entity using flat data structure.
        """

        platform_flat = Platform(
            name="Platform A",
            manufacturer="Manufacturer A",
            description="Platform A description",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(platform=platform_flat))

        """
        Platform entity using explicit data structure.
        """

        platform_explicit = Platform(
            name="Platform A",
            manufacturer=Manufacturer(name="Manufacturer A", tags=["tag 1", "tag 3"]),
            description="Platform A description",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(platform=platform_explicit))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Manufacturer

Attributes:
* `name` (str) - manufacturer name
* `slug` (str) - slug
* `description` (str) - description
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Entity,
    Manufacturer,
)


def main():
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        tls_verify=False,
    ) as client:
        entities = []

        """
        Manufacturer entity.
        """

        manufacturer = Manufacturer(
            name="Manufacturer A",
            description="Manufacturer A description",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(manufacturer=manufacturer))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Site

Attributes:
* `name` (str) - site name
* `slug` (str) - slug
* `status` (str) - status (`active`, `planned`, `retired`, `staging`, `decommissioning`)
* `facility` (str) - facility
* `time_zone` (str) - time zone
* `description` (str) - description
* `comments` (str) - comments
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Entity,
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

        """
        Site entity.
        """

        site = Site(
            name="Site A",
            status="active",
            facility="Data Center 1",
            time_zone="UTC",
            description="Site A description",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(site=site))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Role

Attributes:
* `name` (str) - role name
* `slug` (str) - slug
* `color` (str) - color
* `description` (str) - description
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Entity,
    Role,
)


def main():
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        tls_verify=False,
    ) as client:
        entities = []

        """
        Role entity.
        """

        role = Role(
            name="Role A",
            slug="role-a",
            color="ffffff",
            description="Role A description",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(device_role=role))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## IP Address

Attributes:
* `address` (str) - IP address
* `interface` (str, [Interface](#interface)) - interface name or Interface entity
* `device` (str, [Device](#device)) - device name or Device entity
* `device_type` (str, [DeviceType](#device-type)) - device type name or DeviceType entity
* `device_role` (str, [Role](#role)) - device role name or Role entity
* `platform` (str, [Platform](#platform)) - platform name or Platform entity
* `manufacturer` (str, [Manufacturer](#manufacturer)) - manufacturer name or Manufacturer entity
* `site` (str, Site) - site name or Site entity
* `status` (str) - status (`active`, `reserved`, `deprecated`, `dhcp`, `slaac`)
* `role` (str) - role (`loopback`, `secondary`, `anycast`, `vip`, `vrrp`, `hsrp`, `glbp`, `carp`)
* `dns_name` (str) - DNS name
* `description` (str) - description
* `comments` (str) - comments
* `tags` (list) - tags

### Example

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

        """
        IPAddress entity with only an address provided will attempt to create or update an IP address with
        the given address and placeholders (i.e. "undefined") for other nested objects types
        (e.g. Interface, Device, DeviceType, Platform, Site, Role) required by NetBox.
        """

        ip_address = IPAddress(
            address="192.168.0.1/24",
        )

        entities.append(Entity(ip_address=ip_address))

        """
        IPAddress entity using flat data structure.
        """

        ip_address_flat = IPAddress(
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

        entities.append(Entity(ip_address=ip_address_flat))

        """
        IPAddress entity using explicit data structure.
        """

        ip_address_explicit = IPAddress(
            address="192.168.0.1/24",
            interface=Interface(
                name="Interface ABC",
                device=Device(
                    name="Device ABC",
                    device_type=DeviceType(
                        model="Device Type ABC", manufacturer=Manufacturer(name="Cisco")
                    ),
                    platform=Platform(
                        name="Platform ABC", manufacturer=Manufacturer(name="Cisco")
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

        entities.append(Entity(ip_address=ip_address_explicit))

        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```

## Prefix

Attributes:
* `prefix` (str) - prefix
* `site` (str, [Site](#site)) - site name or Site entity
* `status` (str) - status (`active`, `reserved`, `deprecated`, `container`)
* `is_pool` (bool) - is pool
* `mark_utilized` (bool) - mark utilized
* `description` (str) - description
* `comments` (str) - comments
* `tags` (list) - tags

### Example

```python
from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.ingester import (
    Entity,
    Prefix,
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

        """
        Prefix entity with only a prefix provided will attempt to create or update a prefix with
        the given prefix and placeholders (i.e. "undefined") for other nested objects types (e.g. Site)
        required by NetBox.
        """

        prefix = Prefix(
            prefix="192.168.0.0/32",
        )

        entities.append(Entity(prefix=prefix))

        """
        Prefix entity using flat data structure.
        """

        prefix_flat = Prefix(
            prefix="192.168.0.0/32",
            site="Site ABC",
            status="active",
            is_pool=True,
            mark_utilized=True,
            description="Prefix A description",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(prefix=prefix_flat))

        """
        Prefix entity using explicit data structure.
        """

        prefix_explicit = Prefix(
            prefix="192.168.0.0/32",
            site=Site(
                name="Site ABC",
                status="active",
                facility="Data Center 1",
                time_zone="UTC",
                description="Site A description",
                comments="Lorem ipsum dolor sit amet",
                tags=["tag 1", "tag 2"],
            ),
            is_pool=True,
            mark_utilized=True,
            description="Prefix A description",
            comments="Lorem ipsum dolor sit amet",
            tags=["tag 1", "tag 2"],
        )

        entities.append(Entity(prefix=prefix_explicit))
        
        response = client.ingest(entities=entities)
        if response.errors:
            print(f"Errors: {response.errors}")


if __name__ == "__main__":
    main()

```
