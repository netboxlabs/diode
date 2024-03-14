import uuid

from behave import given, when, then
from steps.utils import send_post_request


@given("I provide a correct payload to create a new site")
def step_set_correct_payload(context):
    """Set the body of the request to create a new site. The site name is Site-Test. The site will be cleaned up after the test."""
    context.body = {
        "change_set_id": str(uuid.uuid4()),
        "change_set": [
            {
                "change_id": str(uuid.uuid4()),
                "change_type": "create",
                "object_version": None,
                "object_type": "dcim.site",
                "object_id": None,
                "data": {
                    "name": "Site-Test",
                    "slug": "site-test",
                    "facility": "Alpha",
                    "description": "",
                    "physical_address": "123 Fake St Lincoln NE 68588",
                    "shipping_address": "123 Fake St Lincoln NE 68588",
                    "comments": "Lorem ipsum etcetera",
                },
            },
        ],
    }


@when("I send a POST request to the endpoint")
def get_response(context):
    """Send a POST request to the endpoint with the payload"""
    context.response = send_post_request(context.body)


@then(
    "I must get a response with status code 200 and a Json object with success message"
)
def check_response_for_correct_payload(context):
    """Check if the response status code is 200 and the result is success"""
    assert context.response.status_code == 200
    assert context.response.json()["result"] == "success"


@given("I provide payload with object_type missing")
def set_incorrect_payload(context):
    """Set the body of the request to create a new site."""
    context.body = {
        "change_set_id": str(uuid.uuid4()),
        "change_set": [
            {
                "change_id": str(uuid.uuid4()),
                "change_type": "create",
                "object_version": None,
                "object_type": "",
                "object_id": None,
                "data": {
                    "name": "Site-Test",
                    "slug": "site-test",
                    "facility": "Alpha",
                    "description": "",
                    "physical_address": "123 Fake St Lincoln NE 68588",
                    "shipping_address": "123 Fake St Lincoln NE 68588",
                    "comments": "Lorem ipsum etcetera",
                },
            },
        ],
    }


@then("I must get a response with status code 400 and a Json object with error message")
def check_response_for_incorrect_payload(context):
    """Check if the response status code is 400 and the change_id in error"""
    change_id = context.body.get("change_set")[0]["change_id"]
    assert context.response.status_code == 400
    assert context.response.json()["errors"][0].get("change_id") == change_id
    assert (
        context.response.json()["errors"][0].get("object_type")
        == "This field may not be blank."
    )
