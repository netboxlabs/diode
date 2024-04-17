import time

from behave import given, when, then

from steps.utils import get_object_by_name, send_delete_request, get_object_by_model


@given('the site object "{site_name}" is deleted')
def delete_site_object(context, site_name):
    """Delete the site object with the given name."""
    endpoint = "dcim/sites/"
    site = get_object_by_name(site_name, endpoint)
    if site:
        send_delete_request(endpoint, site.get("id"))


@given('the device role object "{device_role_name}" is deleted')
def delete_device_role_object(context, device_role_name):
    """Delete the device role object with the given name."""
    endpoint = "dcim/device-roles"
    device_role = get_object_by_name(device_role_name, endpoint)
    if device_role:
        send_delete_request(endpoint, device_role.get("id"))


@given('the manufacturer object "{manufacturer_name}" is deleted')
def delete_manufacturer_object(context, manufacturer_name):
    """Delete the manufacturer object with the given name."""
    endpoint = "dcim/manufacturers"
    manufacturer = get_object_by_name(manufacturer_name, endpoint)
    if manufacturer:
        send_delete_request(endpoint, manufacturer.get("id"))


@given('the device type object "{device_type_model}" is deleted')
def delete_device_type_object(context, device_type_model):
    """Delete the device type object with the given model."""
    endpoint = "dcim/device-types"
    device_type = get_object_by_model(device_type_model, endpoint)
    if device_type:
        send_delete_request(endpoint, device_type.get("id"))


@given('the platform object "{platform_name}" is deleted')
def delete_platform_object(context, platform_name):
    """Delete the manufacturer object with the given name."""
    endpoint = "dcim/platforms"
    platform = get_object_by_name(platform_name, endpoint)
    if platform:
        send_delete_request(endpoint, platform.get("id"))


@given('the device "{device_name}" is deleted')
def delete_device_object(context, device_name):
    """Delete the device object"""
    endpoint = "dcim/devices"
    device = get_object_by_name(device_name, endpoint)
    if device:
        send_delete_request(endpoint, device.get("id"))
