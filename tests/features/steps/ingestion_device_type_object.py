import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.device_type_pb2 import DeviceType
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.manufacturer_pb2 import Manufacturer
from steps.utils import (
    get_object_by_name,
    send_delete_request,
    get_object_by_model,
    ingester,
)

endpoint = "dcim/device-types/"


@given('a new device type "{device_type_model}"')
def step_create_new_manufacturer(context, device_type_model):
    """Set the body of the request to create a new device type."""
    context.device_type_model = device_type_model


@when("the device type is ingested")
def ingest_device_type(context):
    """Ingest the device type object using the Diode SDK"""

    entities = [
        Entity(device_type=DeviceType(model=context.device_type_model)),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the device type and "{manufacturer_name}" manufacturer are created in the database'
)
def check_device_type_and_manufacturers(context, manufacturer_name):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    device_type = get_object_by_model(context.device_type_model, endpoint)
    manufacturer = get_object_by_name(manufacturer_name, "dcim/manufacturers/")
    assert device_type.get("model") == context.device_type_model
    assert manufacturer.get("name") == manufacturer_name


@then("the device type remains the same")
def check_device_type_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    device_type = get_object_by_model(context.device_type_model, endpoint)
    assert device_type.get("model") == context.device_type_model


@given('device type "{device_type_model}" already exists in the database')
def retrieve_existing_manufacturer(context, device_type_model):
    """Retrieve the device type object from the database"""
    time.sleep(3)
    context.device_type_model = device_type_model
    device_type = get_object_by_model(context.device_type_model, endpoint)
    context.device_type_model = device_type.get("model")


@given(
    'device type "{device_type_model}" with manufacturer "{manufacturer_name}", description "{description}" '
    'and part number "{part_number}"'
)
def create_device_type_to_update(
    context, device_type_model, manufacturer_name, description, part_number
):
    """Create a device type object with a description to update"""
    context.device_type_model = device_type_model
    context.manufacturer_name = manufacturer_name
    context.description = description
    context.part_number = part_number


@then(
    'check if the manufacturer "{manufacturer_name}" exists in the database and remove it'
)
def remove_manufacturer(context, manufacturer_name):
    time.sleep(3)
    manufacturer = get_object_by_name(manufacturer_name, "dcim/manufacturers/")
    if manufacturer is not None:
        assert manufacturer.get("name") == manufacturer_name
        send_delete_request("dcim/manufacturers/", manufacturer.get("id"))


@when("the device type object is ingested with the updates")
def ingest_to_update_device_type(context):
    """Update the object using the Diode SDK"""

    entities = [
        Entity(
            device_type=DeviceType(
                model=context.device_type_model,
                manufacturer=Manufacturer(name=context.manufacturer_name),
                description=context.description,
                part_number=context.part_number,
            ),
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then('the manufacturer "{manufacturer_name}" is created and the device updated')
def check_updated_device_type_object(context, manufacturer_name):
    """Check if the response is not None and the object is updated in the database and manufacturer created."""
    time.sleep(3)
    assert context.response is not None

    manufacturer = get_object_by_name(manufacturer_name, "dcim/manufacturers/")
    assert manufacturer.get("name") == manufacturer_name

    device_type = get_object_by_model(context.device_type_model, endpoint)
    assert device_type.get("model") == context.device_type_model
    assert device_type.get("manufacturer").get("name") == manufacturer.get("name")
    assert device_type.get("description") == context.description
    assert device_type.get("part_number") == context.part_number
