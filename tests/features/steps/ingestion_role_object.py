import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.role_pb2 import Role
from steps.utils import get_object_by_name, ingester

endpoint = "dcim/device-roles"


@given('a new device role "{device_role_name}" object')
def step_create_new_role_object(context, device_role_name):
    """Set the body of the request to create a new device role."""
    context.device_role_name = device_role_name


@when("the device role object is ingested")
def ingest_role_object(context):
    """Ingest the device role object using the Diode SDK"""
    entities = [
        Entity(device_role=Role(name=context.device_role_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then("the device role object is created in the database")
@then("the device role object remains the same")
def check_device_role_object(context):
    """Check if the response is not None and the device role object is created in the database."""
    assert context.response is not None
    # Wait for the device role object to be added to the cache
    time.sleep(3)
    device_role = get_object_by_name(context.device_role_name, endpoint)
    assert device_role.get("name") == context.device_role_name
    assert device_role.get("color") == "000000"


@given('device role "{device_role_name}" exists in the database')
def retrieve_existing_device_role_object(context, device_role_name):
    """Retrieve the device role object from the database"""
    context.device_role_name = device_role_name
    context.device_role = get_object_by_name(context.device_role_name, endpoint)
    context.device_role_name = context.device_role.get("name")


@given(
    'device role "{device_role_name}" with color "{color}" and description "{description}"'
)
def create_role_object_to_update(context, device_role_name, color, description):
    """Create a role object with a status and description to update"""
    context.device_role_name = device_role_name
    context.color = color
    context.description = description


@when("the device role object is ingested with the updates")
def update_role_object(context):
    """Update the role object using the Diode SDK"""

    entities = [
        Entity(
            device_role=Role(
                name=context.device_role_name,
                color=context.color,
                description=context.description,
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the device role object is updated in the database")
def check_role_object_updated(context):
    """Check if the role object is updated in the database"""
    assert context.response is not None
    role = get_object_by_name(context.device_role_name, endpoint)
    assert role.get("name") == context.device_role_name
    assert role.get("status").get("value") == context.status
    assert role.get("description") == context.description
