import time

from behave import given, when, then

from steps.utils import get_object_by_name, send_delete_request, get_object_by_model


@given('the site object "{site_name}" is deleted')
def delete_site_object(context, site_name):
    """Delete the site object with the given name."""
    context.site_name = site_name
    endpoint = "dcim/sites/"
    site = get_object_by_name(context.site_name, endpoint)
    if site:
        send_delete_request(endpoint, site.get("id"))


@then("the site object is removed from the database")
def check_site_object_deleted(context):
    """Check if the site object is removed from the database."""
    endpoint = "dcim/sites/"
    site = get_object_by_name(context.site_name, endpoint)
    assert site is None


@given('the device role object "{device_role_name}" is deleted')
def delete_device_role_object(context, device_role_name):
    """Delete the device role object with the given name."""
    context.device_role_name = device_role_name
    endpoint = "dcim/device-roles"
    device_role = get_object_by_name(context.device_role_name, endpoint)
    if device_role:
        send_delete_request(endpoint, device_role.get("id"))


@then("the device role object is removed from the database")
def check_device_role_object_deleted(context):
    """Check if the device role object is removed from the database."""
    endpoint = "dcim/device-roles"
    device_role = get_object_by_name(context.device_role_name, endpoint)
    assert device_role is None


@given('the manufacturer object "{manufacturer_name}" is deleted')
def delete_manufacturer_object(context, manufacturer_name):
    """Delete the manufacturer object with the given name."""
    context.manufacturer_name = manufacturer_name
    endpoint = "dcim/manufacturers"
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    if manufacturer:
        send_delete_request(endpoint, manufacturer.get("id"))


@then("the manufacturer object is removed from the database")
def check_manufacturer_object_deleted(context):
    """Check if the manufacturer object is removed from the database."""
    endpoint = "dcim/manufacturers"
    manufacturer = get_object_by_name(context.manufacturer_name, endpoint)
    assert manufacturer is None


@given('the device type object "{device_type_model}" is deleted')
def delete_device_type_object(context, device_type_model):
    """Delete the device type object with the given model."""
    context.device_type_model = device_type_model
    endpoint = "dcim/device-types"
    device_type = get_object_by_model(context.device_type_model, endpoint)
    if device_type:
        send_delete_request(endpoint, device_type.get("id"))


@then("the device type object is removed from the database")
def check_device_type_object_deleted(context):
    """Check if the device type object is removed from the database."""
    endpoint = "dcim/device-types"
    device_type = get_object_by_model(context.device_type_model, endpoint)
    assert device_type is None
