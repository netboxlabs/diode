import time
from behave import given, when, then
from netboxlabs.diode.sdk.ingester import Entity, Interface, IPAddress
from steps.utils import (
    get_object_state,
    ingester,
)

endpoint = "ipam/ip-addresses/"


@given('an IP address "{ip_address}"')
def set_ip_address(context, ip_address):
    """Set the body of the request to ingest the device."""
    context.ip_address = ip_address
    context.site_name = "undefined"
    context.device_name = "undefined"


@given('interface "{interface_name}"')
def set_interface_name(context, interface_name):
    """Set the body of the request to ingest the device."""
    context.interface_name = interface_name


@given('description "{description}"')
def set_description(context, description):
    """Set the body of the request to ingest the device."""
    context.description = description


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

    params = {
        "object_type": "ipam.ipaddress",
        "q": context.ip_address,
        "interface__name": context.interface_name,
        "interface__device__name": context.device_name,
        "interface__device__site__name": context.site_name,
    }

    if hasattr(context, "description"):
        params["description"] = context.description

    time.sleep(1)

    ip_address = get_object_state(params)

    assert ip_address.get("address") == context.ip_address
    assigned_object = ip_address.get("assigned_object")
    assert assigned_object.get("interface").get("name") == context.interface_name
    assert (
        assigned_object.get("interface").get("device").get("name")
        == context.device_name
    )
    assert (
        assigned_object.get("interface").get("device").get("site").get("name")
        == context.site_name
    )
    context.existing_ip_address = ip_address


@then("the IP address is associated with the interface")
def assert_ip_address_associated_with_interface(
    context,
):
    """Assert that the IP address is associated with the interface."""
    assert context.existing_ip_address is not None
    assert context.existing_ip_address.get("address") == context.ip_address
    assert (
        context.existing_ip_address.get("assigned_object").get("interface").get("name")
        == context.interface_name
    )


@then('the IP address status is "{status}"')
def assert_ip_address_status(context, status):
    """Assert that the IP address status is correct."""
    assert context.existing_ip_address is not None
    assert context.existing_ip_address.get("address") == context.ip_address
    assert context.existing_ip_address.get("status").get("value") == status


@when("the IP address with description is ingested")
def ingest_ip_address_with_description(context):
    """Ingest an IP address using the Diode SDK"""

    entities = [
        Entity(
            ip_address=IPAddress(
                address=context.ip_address,
                interface=Interface(
                    name=context.interface_name,
                ),
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
    assert context.existing_ip_address is not None
    assert context.existing_ip_address.get("address") == context.ip_address
    assert context.existing_ip_address.get("description") == context.description
