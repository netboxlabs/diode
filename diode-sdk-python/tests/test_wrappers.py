#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs - Tests."""

import pytest
from google.protobuf import timestamp_pb2 as _timestamp_pb2

from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Device as DevicePb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    DeviceType as DeviceTypePb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Entity as EntityPb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Interface as InterfacePb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    IPAddress as IPAddressPb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Manufacturer as ManufacturerPb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Platform as PlatformPb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Prefix as PrefixPb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Role as RolePb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Site as SitePb,
)
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Tag as TagPb,
)
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
)


def test_manufacturer_wrapper():
    """Ensure Manufacturer wrapper instantiates ManufacturerPb."""
    entity = Entity(manufacturer="Cisco")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.manufacturer, ManufacturerPb)

    manufacturer = Manufacturer(name="Cisco", tags=["tag1", "tag2"])
    assert isinstance(manufacturer, ManufacturerPb)
    assert manufacturer.name == "Cisco"
    assert len(manufacturer.tags) == 2
    for tag in manufacturer.tags:
        assert isinstance(tag, TagPb)


def test_platform_wrapper():
    """Ensure Platform wrapper instantiates PlatformPb."""
    entity = Entity(platform="Platform ABC")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.platform, PlatformPb)

    platform = Platform(
        name="Platform ABC", manufacturer="Cisco", tags=["tag1", "tag2"]
    )
    assert isinstance(platform, PlatformPb)
    assert platform.name == "Platform ABC"
    assert isinstance(platform.manufacturer, ManufacturerPb)
    assert platform.manufacturer.name == "Cisco"
    assert len(platform.tags) == 2
    for tag in platform.tags:
        assert isinstance(tag, TagPb)

    platform = Platform(name="Platform ABC", manufacturer=Manufacturer(name="Cisco"))
    assert isinstance(platform, PlatformPb)
    assert platform.name == "Platform ABC"
    assert isinstance(platform.manufacturer, ManufacturerPb)
    assert platform.manufacturer.name == "Cisco"

    platform = Platform(name="Platform ABC", manufacturer=ManufacturerPb(name="Cisco"))
    assert isinstance(platform, PlatformPb)
    assert platform.name == "Platform ABC"
    assert isinstance(platform.manufacturer, ManufacturerPb)
    assert platform.manufacturer.name == "Cisco"


def test_role_wrapper():
    """Ensure Role wrapper instantiates RolePb."""
    entity = Entity(device_role="Role ABC")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.device_role, RolePb)

    role = Role(name="Role ABC", color="red", slug="role-abc", tags=["tag1", "tag2"])
    assert isinstance(role, RolePb)
    assert role.name == "Role ABC"
    assert role.color == "red"
    assert role.slug == "role-abc"
    assert len(role.tags) == 2
    for tag in role.tags:
        assert isinstance(tag, TagPb)


def test_device_type_wrapper():
    """Ensure DeviceType wrapper instantiates DeviceTypePb."""
    entity = Entity(device_type="Device Type ABC")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.device_type, DeviceTypePb)

    device_type = DeviceType(
        model="Device Type ABC",
        manufacturer="Cisco",
        slug="device-type-abc",
        tags=["tag1", "tag2"],
    )
    assert isinstance(device_type, DeviceTypePb)
    assert device_type.model == "Device Type ABC"
    assert isinstance(device_type.manufacturer, ManufacturerPb)
    assert device_type.manufacturer.name == "Cisco"
    assert device_type.slug == "device-type-abc"
    assert len(device_type.tags) == 2
    for tag in device_type.tags:
        assert isinstance(tag, TagPb)

    device_type = DeviceType(
        model="Device Type ABC",
        manufacturer=Manufacturer(name="Cisco"),
        slug="device-type-abc",
    )
    assert isinstance(device_type, DeviceTypePb)
    assert device_type.model == "Device Type ABC"
    assert isinstance(device_type.manufacturer, ManufacturerPb)
    assert device_type.manufacturer.name == "Cisco"
    assert device_type.slug == "device-type-abc"

    device_type = DeviceType(
        model="Device Type ABC",
        manufacturer=ManufacturerPb(name="Cisco"),
        slug="device-type-abc",
    )
    assert isinstance(device_type, DeviceTypePb)
    assert device_type.model == "Device Type ABC"
    assert isinstance(device_type.manufacturer, ManufacturerPb)
    assert device_type.manufacturer.name == "Cisco"


def test_device_wrapper():
    """Ensure Device wrapper instantiates DevicePb."""
    entity = Entity(device="Device ABC")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.device, DevicePb)

    device = Device(
        name="Device ABC",
        device_type="Device Type ABC",
        platform="Platform ABC",
        manufacturer="Cisco",
        site="Site ABC",
        role="Role ABC",
        serial="123456",
        asset_tag="123456",
        status="active",
    )
    assert isinstance(device, DevicePb)
    assert device.name == "Device ABC"
    assert isinstance(device.device_type, DeviceTypePb)
    assert device.device_type.model == "Device Type ABC"
    assert isinstance(device.platform, PlatformPb)
    assert device.platform.name == "Platform ABC"
    assert isinstance(device.platform.manufacturer, ManufacturerPb)
    assert device.platform.manufacturer.name == "Cisco"
    assert isinstance(device.site, SitePb)
    assert device.site.name == "Site ABC"
    assert isinstance(device.role, RolePb)
    assert device.role.name == "Role ABC"
    assert device.serial == "123456"
    assert device.asset_tag == "123456"
    assert device.status == "active"

    device = Device(
        name="Device ABC",
        device_type=DeviceType(model="Device Type ABC"),
        platform=Platform(name="Platform ABC", manufacturer="Cisco"),
        site=SitePb(name="Site ABC"),
        role=Role(name="Role ABC"),
        serial="123456",
        asset_tag="123456",
        status="active",
    )
    assert isinstance(device, DevicePb)
    assert device.name == "Device ABC"
    assert isinstance(device.device_type, DeviceTypePb)
    assert device.device_type.model == "Device Type ABC"
    assert isinstance(device.platform, PlatformPb)
    assert device.platform.name == "Platform ABC"
    assert isinstance(device.platform.manufacturer, ManufacturerPb)
    assert device.platform.manufacturer.name == "Cisco"
    assert isinstance(device.site, SitePb)
    assert device.site.name == "Site ABC"
    assert isinstance(device.role, RolePb)
    assert device.role.name == "Role ABC"
    assert device.serial == "123456"
    assert device.asset_tag == "123456"
    assert device.status == "active"


def test_interface_wrapper():
    """Ensure Interface wrapper instantiates InterfacePb."""
    entity = Entity(interface="Interface ABC")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.interface, InterfacePb)

    interface = Interface(
        name="Interface ABC",
        device="Device ABC",
        device_type="Device Type ABC",
        role="Role ABC",
        platform="Platform ABC",
        site="Site ABC",
        type="type",
        enabled=True,
        mtu=1500,
        mac_address="00:00:00:00:00:00",
        description="Description ABC",
    )
    assert isinstance(interface, InterfacePb)
    assert interface.name == "Interface ABC"
    assert isinstance(interface.device, DevicePb)
    assert interface.device.name == "Device ABC"
    assert isinstance(interface.device.device_type, DeviceTypePb)
    assert interface.device.device_type.model == "Device Type ABC"
    assert isinstance(interface.device.role, RolePb)
    assert interface.device.role.name == "Role ABC"
    assert isinstance(interface.device.platform, PlatformPb)
    assert interface.device.platform.name == "Platform ABC"
    assert isinstance(interface.device.platform.manufacturer, ManufacturerPb)
    assert interface.device.platform.manufacturer == ManufacturerPb()
    assert isinstance(interface.device.site, SitePb)
    assert interface.device.site.name == "Site ABC"
    assert interface.type == "type"
    assert interface.enabled is True
    assert interface.mtu == 1500
    assert interface.mac_address == "00:00:00:00:00:00"
    assert interface.description == "Description ABC"


def test_ip_address_wrapper():
    """Ensure IPAddress wrapper instantiates IPAddressPb."""
    entity = Entity(ip_address="192.168.0.1/24")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.ip_address, IPAddressPb)
    assert entity.ip_address.address == "192.168.0.1/24"

    ip_address = IPAddress(
        address="192.168.0.1/24",
        interface="Interface ABC",
        tags=["tag1", "tag2"],
    )
    assert isinstance(ip_address, IPAddressPb)
    assert ip_address.address == "192.168.0.1/24"
    assert isinstance(ip_address.interface, InterfacePb)
    assert ip_address.interface.name == "Interface ABC"
    assert ip_address.interface.device == DevicePb()
    assert len(ip_address.tags) == 2
    for tag in ip_address.tags:
        assert isinstance(tag, TagPb)

    ip_address = IPAddress(
        address="192.168.0.1/24",
        interface="Interface ABC",
        device="Device ABC",
        device_type="Device Type ABC",
        device_role="Role ABC",
        platform="Platform ABC",
        manufacturer="Cisco",
        site="Site ABC",
    )
    assert isinstance(ip_address, IPAddressPb)
    assert ip_address.address == "192.168.0.1/24"
    assert isinstance(ip_address.interface, InterfacePb)
    assert ip_address.interface.name == "Interface ABC"
    assert isinstance(ip_address.interface.device, DevicePb)
    assert ip_address.interface.device.name == "Device ABC"
    assert isinstance(ip_address.interface.device.device_type, DeviceTypePb)
    assert ip_address.interface.device.device_type.model == "Device Type ABC"
    assert isinstance(ip_address.interface.device.role, RolePb)
    assert ip_address.interface.device.role.name == "Role ABC"
    assert isinstance(ip_address.interface.device.platform, PlatformPb)
    assert ip_address.interface.device.platform.name == "Platform ABC"
    assert isinstance(ip_address.interface.device.platform.manufacturer, ManufacturerPb)
    assert ip_address.interface.device.platform.manufacturer.name == "Cisco"


def test_prefix_wrapper():
    """Ensure Prefix wrapper instantiates PrefixPb."""
    entity = Entity(prefix="192.168.0.0/32")
    assert isinstance(entity, EntityPb)
    assert isinstance(entity.prefix, PrefixPb)

    prefix = Prefix(
        prefix="192.168.0.0/32",
        site="Site ABC",
        is_pool=True,
    )
    assert isinstance(prefix, PrefixPb)
    assert prefix.prefix == "192.168.0.0/32"
    assert isinstance(prefix.site, SitePb)
    assert prefix.site.name == "Site ABC"
    assert prefix.is_pool is True
