import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import Entity, Interface
from steps.utils import (
    get_object_state,
    get_object_by_name,
    ingester,
)

endpoint = "dcim/interfaces/"


@given('a new interface "{interface_name}"')
def create_interface(context, interface_name):
    """Set the body of the request to create an interface."""
    context.interface_name = interface_name


@when("the interface is ingested")
def ingest_interface(context):
    """Ingest an interface using the Diode SDK"""

    entities = [
        Entity(
            interface=Interface(
                name=context.interface_name,
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the interface is found")
def assert_interface_exists(context):
    """Assert that the interface exists."""
    assert context.response is not None

    attempt = 0
    max_attempts = 3

    obj = None

    while obj is None and attempt < max_attempts:
        obj = get_object_by_name(context.interface_name, endpoint)
        if obj:
            break

        time.sleep(1)
        attempt += 1

    assert obj.get("name") == context.interface_name
    context.interface = obj


@then('the interface is associated with the device "{device_name}"')
def assert_interface_associated_with_device(context, device_name):
    """Assert that the interface was associated with the device."""
    assert context.interface is not None

    interface = context.interface

    assert interface.get("name") == context.interface_name
    assert interface.get("device").get("name") == device_name


@then("the interface is enabled")
def assert_interface_enabled(context):
    """Assert that the interface was enabled."""
    assert context.interface is not None

    interface = context.interface

    assert interface.get("name") == context.interface_name
    assert interface.get("enabled") is True


@then('the interface type is "{interface_type}"')
def assert_interface_type(context, interface_type):
    """Assert that the interface type is correct."""
    assert context.interface is not None

    interface = context.interface

    assert interface.get("name") == context.interface_name
    assert interface.get("type").get("value") == interface_type


@given('an interface "{interface_name}" with MTU "{mtu}"')
def update_interface_with_mtu(context, interface_name, mtu):
    """Set the body of the request to update an interface with an MTU."""
    context.interface_name = interface_name
    context.mtu = int(mtu)


@when("the interface with MTU is ingested")
def ingest_interface_with_mtu(context):
    """Ingest an interface using the Diode SDK"""

    entities = [
        Entity(
            interface=Interface(
                name=context.interface_name,
                mtu=context.mtu,
            ),
        ),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return context.response


@then("the interface MTU is updated")
def assert_interface_mtu(context):
    """Assert that the interface MTU is correct."""
    assert context.interface is not None

    interface = context.interface

    attempt = 0
    max_attempts = 3

    while interface.get("mtu") != context.mtu and attempt < max_attempts:
        interface = get_object_by_name(context.interface_name, endpoint)
        time.sleep(1)
        attempt += 1

    assert interface.get("name") == context.interface_name
    assert interface.get("mtu") == context.mtu


@given('an interface "{interface_name}" with enabled field set to "{field_value}"')
def set_interface_with_enabled_field(context, interface_name, field_value):
    """Set the body of the request to ingest an interface with enabled field."""
    context.interface_name = interface_name
    context.enabled = str.lower(field_value) == "true"


@when("the interface with enabled field is ingested")
def ingest_interface_enabled_with_field(context):
    """Ingest an interface using the Diode SDK"""
    entities = [
        Entity(
            interface=Interface(
                name=context.interface_name,
                enabled=context.enabled,
            )
        )
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@then("the interface with enabled field is found")
def assert_interface_with_ingested_enabled_field_is_found(context):
    """Assert that the interface exists."""
    assert context.response is not None

    params = {
        "object_type": "dcim.interface",
        "q": context.interface_name,
        "enabled": context.enabled,
    }

    interface = get_object_state(params)

    assert interface.get("name") == context.interface_name

    context.existing_interface = interface


@then('enabled field is "{field_value}"')
def assert_enabled_field(context, field_value):
    """Assert that the enabled field is correct."""
    assert context.existing_interface is not None
    assert context.existing_interface.get("enabled") == (
        str.lower(field_value) == "true"
    )
