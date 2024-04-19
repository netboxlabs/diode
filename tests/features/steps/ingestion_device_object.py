import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.device_pb2 import Device
from netboxlabs.diode.sdk.diode.v1.device_type_pb2 import DeviceType
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.platform_pb2 import Platform
from netboxlabs.diode.sdk.diode.v1.role_pb2 import Role
from netboxlabs.diode.sdk.diode.v1.site_pb2 import Site
from steps.utils import get_object_by_name, get_object_by_model, ingester

endpoint = "dcim/devices/"


@given('a new device "{device_name}"')
def create_new_device(context, device_name):
    """Set the body of the request to create a new device."""
    context.device_name = device_name


@when("the device is ingested")
def ingest_device(context):
    """Ingest the device using the Diode SDK"""

    entities = [
        Entity(device=Device(name=context.device_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the device, device type "{device_type_model}" , role "{device_role_name}", and site "{site_name}" are created'
)
def verify_device_created(context, device_type_model, device_role_name, site_name):
    """Verify that the device was created."""
    time.sleep(3)
    assert context.response is not None

    device = get_object_by_name(context.device_name, endpoint)

    device_type = get_object_by_model(device_type_model, "dcim/device-types/")
    device_role = get_object_by_name(device_role_name, "dcim/device-roles/")
    site = get_object_by_name(site_name, "dcim/sites/")

    assert device.get("name") == context.device_name
    assert device.get("device_type").get("model") == device_type_model
    assert device.get("device_role").get("name") == device_role_name
    assert device.get("site").get("name") == site_name

    assert device_type.get("model") == device_type_model
    assert device_role.get("name") == device_role_name
    assert site.get("name") == site_name


@given(
    'device type "{device_type_model}", role "{device_role_name}", and site "{site_name}" for device "{device_name}"'
)
def create_device_to_update(
    context, device_type_model, device_role_name, site_name, device_name
):
    """Create device type, device role, site, and device_name context."""
    context.device_type_model = device_type_model
    context.device_role_name = device_role_name
    context.site_name = site_name
    context.device_name = device_name


@when("the device is ingested with the updates")
def ingest_device_with_updates(context):
    """Ingest the device using the Diode SDK."""

    entities = [
        Entity(
            device=Device(
                name=context.device_name,
                device_type=DeviceType(model=context.device_type_model),
                role=Role(name=context.device_role_name),
                site=Site(name=context.site_name),
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("device type, role and site are created")
def verify_device_type_role_site_created(context):
    """Verify that the device was created."""
    time.sleep(3)
    assert context.response is not None

    device_type = get_object_by_model(context.device_type_model, "dcim/device-types/")
    device_role = get_object_by_name(context.device_role_name, "dcim/device-roles/")
    site = get_object_by_name(context.site_name, "dcim/sites/")

    assert device_type.get("model") == context.device_type_model
    assert device_role.get("name") == context.device_role_name
    assert site.get("name") == context.site_name


@then("the device is updated")
def verify_device_updated(context):
    """Verify that the device was created."""
    time.sleep(3)
    assert context.response is not None

    device = get_object_by_name(context.device_name, endpoint)

    assert device.get("name") == context.device_name
    assert device.get("device_type").get("model") == context.device_type_model
    assert device.get("device_role").get("name") == context.device_role_name
    assert device.get("site").get("name") == context.site_name


@given(
    'platform "{platform_name}", device type "{device_type_model}", role "{device_role_name}", and site "{site_name}" for device "{device_name}"'
)
def create_device_context_to_update(
    context, platform_name, device_type_model, device_role_name, site_name, device_name
):
    """Create device type, device role, site, and device_name context."""
    context.platform_name = platform_name
    context.device_type_model = device_type_model
    context.device_role_name = device_role_name
    context.site_name = site_name
    context.device_name = device_name


@when("the device is ingested with the platform update")
def ingest_device_with_platform_update(context):
    """Ingest the device using the Diode SDK."""

    entities = [
        Entity(
            device=Device(
                name=context.device_name,
                device_type=DeviceType(model=context.device_type_model),
                role=Role(name=context.device_role_name),
                site=Site(name=context.site_name),
                platform=Platform(name=context.platform_name),
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the device is updated and platform is created")
def verify_device_platform_updated(context):
    time.sleep(3)
    assert context.response is not None

    device = get_object_by_name(context.device_name, endpoint)

    device_type = get_object_by_model(context.device_type_model, "dcim/device-types/")
    device_role = get_object_by_name(context.device_role_name, "dcim/device-roles/")
    site = get_object_by_name(context.site_name, "dcim/sites/")
    platform = get_object_by_name(context.platform_name, "dcim/platforms/")

    assert device.get("name") == context.device_name
    assert device.get("device_type").get("model") == context.device_type_model
    assert device.get("device_role").get("name") == context.device_role_name
    assert device.get("site").get("name") == context.site_name
    assert device.get("platform").get("name") == context.platform_name

    assert device_type.get("model") == context.device_type_model
    assert device_role.get("name") == context.device_role_name
    assert site.get("name") == context.site_name
    assert platform.get("name") == context.platform_name
