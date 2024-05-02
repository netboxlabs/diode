import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.prefix_pb2 import Prefix
from steps.utils import (
    get_object_by_name,
    ingester,
)

endpoint = "ipam/prefixes/"


@given('a new prefix "{prefix_prefix}"')
def create_prefix(context, prefix_prefix):
    """Set the body of the request to create a prefix."""
    context.prefix_prefix = prefix_prefix


@when("the prefix is ingested")
def ingest_prefix(context):
    """Ingest a prefix using the Diode SDK"""

    entities = [
        Entity(
            prefix=Prefix(
                prefix=context.prefix_prefix,
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the prefix is found")
def assert_prefix_exists(context):
    """Assert that the prefix exists."""
    assert context.response is not None

    attempt = 0
    max_attempts = 3

    obj = None

    while obj is None and attempt < max_attempts:
        obj = get_object_by_name(context.prefix_prefix, endpoint)
        if obj:
            break

        time.sleep(1)
        attempt += 1

    assert obj.get("prefix") == context.prefix_prefix
    context.prefix = obj


@then('the prefix is associated with the site "{site_name}"')
def assert_prefix_associated_with_site(context, site_name):
    """Assert that the prefix was associated with the site."""
    assert context.prefix is not None

    prefix = context.prefix

    assert prefix.get("prefix") == context.prefix_prefix
    assert prefix.get("site").get("name") == site_name


@then("the prefix is active")
def assert_prefix_active(context):
    """Assert that the prefix is active."""
    assert context.prefix is not None

    assert context.prefix.get("prefix") == context.prefix_prefix
    assert context.prefix.get("status").get("value") == "active"


@given('a prefix "{prefix_prefix}" with description "{description}"')
def update_prefix_with_description(context, prefix_prefix, description):
    """Set the body of the request to update a prefix with description."""
    context.prefix_prefix = prefix_prefix
    context.description = description


@when("the prefix with description is ingested")
def ingest_prefix_with_description(context):
    """Ingest a prefix using the Diode SDK"""

    entities = [
        Entity(
            prefix=Prefix(
                prefix=context.prefix_prefix,
                description=context.description,
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the prefix description is updated")
def assert_prefix_description(context):
    """Assert that the prefix description is updated."""
    assert context.prefix is not None

    prefix = context.prefix

    attempt = 0
    max_attempts = 3

    while prefix.get("description") != context.description and attempt < max_attempts:
        prefix = get_object_by_name(context.prefix_prefix, endpoint)
        time.sleep(1)
        attempt += 1

    assert prefix.get("prefix") == context.prefix_prefix
    assert prefix.get("description") == context.description
