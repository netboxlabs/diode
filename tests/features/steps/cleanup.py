from behave import given, when, then

from steps.utils import get_object_by_name, send_delete_request


@given('the site object "{site_name}" is deleted')
def delete_site_object(context, site_name):
    context.site_name = site_name
    endpoint = "dcim/sites/"
    site = get_object_by_name(context.site_name, endpoint)
    send_delete_request(endpoint, site.get("id"))


@then("the site object is removed from the database")
def check_site_object_deleted(context):
    """Check if the response status code is 200 and the result is success"""
    endpoint = "dcim/sites/"
    site = get_object_by_name(context.site_name, endpoint)
    assert site is None
