import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.manufacturer_pb2 import Manufacturer
from steps.utils import get_object_by_name, ingester

endpoint = "dcim/manufacturers/"


@given('a new manufacturer "{manufacturer_name}" object')
def step_create_new_manufacturer_object(context, manufacturer_name):
    """Set the body of the request to create a new manufacturer."""
    context.manufacturer_name = manufacturer_name


@when("the manufacturer object is ingested")
def ingest_manufacturer_object(context):
    """Ingest the manufacturer object using the Diode SDK"""

    entities = [
        Entity(manufacturer=Manufacturer(name=context.manufacturer_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then("the manufacturer object is created in the database")
@then("the manufacturer object remains the same")
def check_manufacturer_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    assert manufacturer.get("name") == context.manufacturer_name
    assert manufacturer.get("slug") == "cisco"


@given('manufacturer "{manufacturer_name}" already exists in the database')
def retrieve_existing_manufacturer_object(context, manufacturer_name):
    """Retrieve the manufacturer object from the database"""
    context.manufacturer_name = manufacturer_name
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    context.manufacturer_name = manufacturer.get("name")


@given('manufacturer {manufacturer_name} with description "{description}"')
def create_manufacturer_object_to_update(context, manufacturer_name, description):
    """Create a manufacturer object with a description to update"""
    context.manufacturer_name = manufacturer_name
    context.description = description


@when("the manufacturer object is ingested with the updates")
def update_manufacturer_object(context):
    """Update the object using the Diode SDK"""

    entities = [
        Entity(
            manufacturer=Manufacturer(
                name=context.manufacturer_name,
                description=context.description,
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the manufacturer object is updated in the database")
def check_manufacturer_object_updated(context):
    """Check if the response is not None and the object is updated in the database."""
    assert context.response is not None
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    assert manufacturer.get("name") == context.manufacturer_name
    assert manufacturer.get("description") == context.description
