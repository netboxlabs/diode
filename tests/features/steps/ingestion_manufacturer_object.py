import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.manufacturer_pb2 import Manufacturer
from steps.utils import get_object_by_name, ingester

endpoint = "dcim/manufacturers/"


@given('a new manufacturer "{manufacturer_name}"')
def step_create_new_manufacturer(context, manufacturer_name):
    """Set the body of the request to create a new manufacturer."""
    context.manufacturer_name = manufacturer_name


@when("the manufacturer is ingested")
def ingest_manufacturer(context):
    """Ingest the manufacturer using the Diode SDK"""

    entities = [
        Entity(manufacturer=Manufacturer(name=context.manufacturer_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then("the manufacturer is created in the database")
@then("the manufacturer remains the same")
def check_manufacturer_(context):
    """Check if the response is not None and the is created in the database."""
    time.sleep(3)
    assert context.response is not None
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    assert manufacturer.get("name") == context.manufacturer_name
    assert manufacturer.get("slug") == "cisco"


@given('manufacturer "{manufacturer_name}" already exists in the database')
def retrieve_existing_manufacturer(context, manufacturer_name):
    """Retrieve the manufacturer from the database"""
    context.manufacturer_name = manufacturer_name
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    context.manufacturer_name = manufacturer.get("name")


@given('manufacturer {manufacturer_name} with description "{description}"')
def create_manufacturer_to_update(context, manufacturer_name, description):
    """Create a manufacturer with a description to update"""
    context.manufacturer_name = manufacturer_name
    context.description = description


@when("the manufacturer is ingested with the updates")
def ingest_to_update_manufacturer(context):
    """Update the manufacturer using the Diode SDK"""

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


@then("the manufacturer is updated in the database")
def check_manufacturer_updated(context):
    """Check if the response is not None and the is updated in the database."""
    assert context.response is not None
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    assert manufacturer.get("name") == context.manufacturer_name
    assert manufacturer.get("description") == context.description
