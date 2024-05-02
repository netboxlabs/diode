import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.interface_pb2 import Interface
from netboxlabs.diode.sdk.diode.v1.ip_address_pb2 import IPAddress
from steps.utils import (
    get_object_by_name,
    ingester,
)

endpoint = "ipam/ip-addresses/"


@given('a new IP address "{ip_address}" and interface "{interface_name}"')
def create_ip_address(context, ip_address, interface_name):
    """Set the body of the request to create an IP address."""
    context.ip_address = ip_address
    context.interface_name = interface_name


@when("the IP address is ingested")
def ingest_ip_address(context):
    """Ingest an IP address using the Diode SDK"""

    entities = [
        Entity(
            ip_address=IPAddress(
                address=context.ip_address,
                interface=Interface(
                    name=context.interface_name,
                ),
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the IP address is found")
def assert_ip_address_exists(context):
    """Assert that the IP address exists."""
    assert context.response is not None

    attempt = 0
    max_attempts = 3

    obj = None

    while obj is None and attempt < max_attempts:
        obj = get_object_by_name(context.ip_address, endpoint)
        if obj:
            break

        time.sleep(1)
        attempt += 1

    assert obj.get("address") == context.ip_address
    context.ip = obj


@then("the IP address is associated with the interface")
def assert_ip_address_associated_with_interface(
    context,
):
    """Assert that the IP address is associated with the interface."""
    assert context.ip is not None

    ip = context.ip

    assert ip.get("address") == context.ip_address
    assert ip.get("assigned_object").get("name") == context.interface_name


@then('the IP address status is "{status}"')
def assert_ip_address_status(context, status):
    """Assert that the IP address status is correct."""
    assert context.ip is not None

    ip = context.ip

    assert ip.get("address") == context.ip_address
    assert ip.get("status").get("value") == status


@given('an IP address "{ip_address}" with description "{description}"')
def update_ip_address_with_description(context, ip_address, description):
    """Set the body of the request to update an IP address with description."""
    context.ip_address = ip_address
    context.description = description


@when("the IP address with description is ingested")
def ingest_ip_address_with_description(context):
    """Ingest an IP address using the Diode SDK"""

    entities = [
        Entity(
            ip_address=IPAddress(
                address=context.ip_address,
                description=context.description,
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the IP address description is updated")
def assert_ip_address_description(context):
    """Assert that the IP address description is correct."""
    assert context.ip is not None

    ip = context.ip

    attempt = 0
    max_attempts = 3

    while ip.get("description") != context.description and attempt < max_attempts:
        ip = get_object_by_name(context.ip_address, endpoint)
        time.sleep(1)
        attempt += 1

    assert ip.get("address") == context.ip_address
    assert ip.get("description") == context.description
