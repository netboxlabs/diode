from steps.utils import send_post_request, send_get_request, get_site_id
from behave import given, then


@given('the site id "{site_name}" and object_type "{object_type}"')
def get_object_state_using_id(context, site_name, object_type):
    site_id = get_site_id(site_name)
    endpoint = "plugins/diode/object-state/"
    params = {"id": site_id, "object_type": object_type}
    response = send_get_request(endpoint, params)
    context.response = response


@then('the object state "{site_name}" is returned successfully')
def check_object_state_response(context, site_name):
    assert context.response.status_code == 200
    assert context.response.json().get("object").get("name") == site_name


@given('the site name "{site_name}" and object_type "{object_type}"')
@given('the site name "{site_name}" and not object_type')
def get_object_state_using_name(context, site_name, object_type=None):
    endpoint = "plugins/diode/object-state/"
    params = {"q": site_name, "object_type": object_type}
    response = send_get_request(endpoint, params)
    context.response = response


@then("endpoint return 200 and empty response")
def check_object_state_response(context):
    assert context.response.status_code == 200
    assert context.response.json() == {}


@then("endpoint return 400")
def check_object_state_response(context):
    assert context.response.status_code == 400
