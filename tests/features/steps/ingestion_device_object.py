import time

from behave import given, when, then

from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.device_pb2 import Device

from steps.config import configs

from steps.utils import get_object_by_name, send_delete_request, get_object_by_model


api_key = str(configs["api_key"])
endpoint = "dcim/devices/"


@given('a new device "{device_name}" object')
def create_new_device_object(context, device_name):
    """Set the body of the request to create a new device."""
    context.device_name = device_name


@when("the device object is ingested")
def ingest_device_object(context):
    """Ingest the device object using the Diode SDK"""
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        api_key=api_key,
    ) as client:
        entities = [
            Entity(device=Device(name=context.device_name)),
        ]

        context.response = client.ingest(entities=entities)
        return context.response


@then(
    'the device, device_type "{device_type_model}" , role "{device_role_name}", and site "{site_name}" are created'
)
def verify_device_object(context, device_type_model, device_role_name, site_name):
    """Verify that the device object was created."""
    time.sleep(3)
    assert context.response is not None
    device = get_object_by_name(context.device_name, endpoint)
    # device_type = get_object_by_model(device_type_model, "dcim/device-types/")
    # device_role = get_object_by_name(device_role_name, "dcim/device-roles/")
    # site = get_object_by_name(site_name, "dcim/sites/")

    assert device.get("name") == context.device_name
    assert device.get("device_type") == device_type_model
    assert device.get("device_role") == device_role_name
    assert device.get("site") == site_name

    # assert device_type.get("model") == device_type_model
    # assert device_role.get("name") == device_role_name
    # assert site.get("name") == site_name
