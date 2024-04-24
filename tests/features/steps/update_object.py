import uuid

from behave import given
from steps.utils import get_site_id


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
