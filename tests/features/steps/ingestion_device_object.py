from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import (
    Device,
    DeviceType,
    Entity,
    Role,
    Site,
)
from steps.utils import (
    get_object_state,
    ingester,
    send_delete_request,
)


endpoint = "dcim/devices/"


@given('device "{device_name}" with site not provided')
def set_device_without_site(context, device_name):
    """Set the body of the request to ingest the device."""
    context.device_name = device_name
    context.site_name = "undefined"
    context.device_type_model = "undefined"
    context.device_role_name = "undefined"


@given('device "{device_name}" with site "{site_name}" does not exist')
def ensure_device_does_not_exists(context, device_name, site_name):
    """Ensure that the device does not exist."""
    device = get_object_state(
        {
            "object_type": "dcim.device",
            "q": device_name,
            "site__name": site_name,
        },
    )
    if device:
        send_delete_request(endpoint, device.get("id"))


@when("the device without site is ingested")
def ingest_device_without_site(context):
    """Ingest the device using the Diode SDK"""

    entities = [
        Entity(device=Device(name=context.device_name)),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return response


@then("the device is found")
def assert_device_exists(context):
    """Assert that the device was created."""
    assert context.response is not None

    params = {
        "object_type": "dcim.device",
        "q": context.device_name,
        "site__name": context.site_name,
    }

    if hasattr(context, "device_type_model"):
        params["device_type__model"] = context.device_type_model
    if hasattr(context, "device_role_name"):
        params["role__name"] = context.device_role_name

    device = get_object_state(params)

    assert device.get("name") == context.device_name
    assert device.get("site").get("name") == context.site_name
    context.existing_device = device


@then('device type is "{device_type_model}"')
def assert_device_type(context, device_type_model):
    """Assert that the device type is correct."""
    assert context.existing_device is not None
    assert context.existing_device.get("device_type").get("model") == device_type_model


@then('role is "{device_role_name}"')
def assert_device_role(context, device_role_name):
    """Assert that the device role is correct."""
    assert context.existing_device is not None
    assert context.existing_device.get("device_role").get("name") == device_role_name


@given('device "{device_name}" with site "{site_name}" exists')
def assert_device_exists_with_site(context, device_name, site_name):
    """Assert that the device exists."""
    device = get_object_state(
        {
            "object_type": "dcim.device",
            "q": device_name,
            "site__name": site_name,
        },
    )

    assert device.get("name") == device_name
    assert device.get("site").get("name") == site_name
    context.existing_device = device


@given('a new device "{device_name}" with site "{site_name}"')
def create_new_device_with_site(context, device_name, site_name):
    """Set the body of the request to create a new device with site."""
    context.device_name = device_name
    context.site_name = site_name
    context.device_type_model = "undefined"
    context.device_role_name = "undefined"


@when("the device with site is ingested")
def ingest_device_with_site(context):
    """Ingest the device using the Diode SDK"""
    entities = [
        Entity(
            device=Device(
                name=context.device_name,
                site=Site(name=context.site_name),
            ),
        ),
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@given(
    'device "{device_name}" with site "{site_name}", device type "{device_type_model}" and role "{device_role_name}"'
)
def update_device(context, device_name, site_name, device_type_model, device_role_name):
    """Set the body of the request to update a device."""
    context.device_name = device_name
    context.site_name = site_name
    context.device_type_model = device_type_model
    context.device_role_name = device_role_name


@when("the device with site, device type and role is ingested")
def ingest_device_with_site_device_type_and_role(context):
    """Ingest the device using the Diode SDK"""
    entities = [
        Entity(
            device=Device(
                name=context.device_name,
                site=Site(name=context.site_name),
                device_type=DeviceType(model=context.device_type_model),
                role=Role(name=context.device_role_name),
            ),
        ),
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response
