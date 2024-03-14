import uuid

from behave import given, when, then
from steps.utils import send_post_request, send_get_request


def get_site_id(site_name):
    endpoint = "dcim/sites/"
    site_id = (
        send_get_request(endpoint, {"name__ic": site_name})
        .json()
        .get("results")[0]
        .get("id")
    )
    return site_id


@given("I provide a correct payload to update the slug of the site")
def step_set_correct_update_payload(context):
    """Set the body of the request to create a new site. The site name is Site-Test. The site will be cleaned up after the test."""

    object_id = get_site_id("Site-Test-2")

    context.body = {
        "change_set_id": str(uuid.uuid4()),
        "change_set": [
            {
                "change_id": str(uuid.uuid4()),
                "change_type": "update",
                "object_version": None,
                "object_type": "dcim.site",
                "object_id": object_id,
                "data": {
                    "name": "Site-Test-2",
                    "slug": "slug-updated",
                    "facility": "Alpha",
                    "description": "",
                    "physical_address": "123 Fake St Lincoln NE 68588",
                    "shipping_address": "123 Fake St Lincoln NE 68588",
                    "comments": "Lorem ipsum etcetera",
                },
            },
        ],
    }
    context.sites_to_be_cleaned_up = [
        "Site-Test-2",
    ]


@given("I provide payload with object_type missing for update")
def set_incorrect_update_payload(context):
    """Set the body of the request to create a new site."""

    object_id = get_site_id("Site-Test-2")

    context.body = {
        "change_set_id": str(uuid.uuid4()),
        "change_set": [
            {
                "change_id": str(uuid.uuid4()),
                "change_type": "update",
                "object_version": None,
                "object_type": "",
                "object_id": object_id,
                "data": {
                    "name": "Site-Test-2",
                    "slug": "slug-updated",
                    "facility": "Alpha",
                    "description": "",
                    "physical_address": "123 Fake St Lincoln NE 68588",
                    "shipping_address": "123 Fake St Lincoln NE 68588",
                    "comments": "Lorem ipsum etcetera",
                },
            },
        ],
    }


# @then("I must get a response with status code 400 and a Json object with error message")
# def check_response_for_incorrect_payload(context):
#     """Check if the response status code is 400 and the change_id in error"""
#     change_id = context.body.get("change_set")[0]["change_id"]
#     assert context.response.status_code == 400
#     assert context.response.json()["errors"][0].get("change_id") == change_id
#     assert (
#         context.response.json()["errors"][0].get("object_type")
#         == "This field may not be blank."
#     )
