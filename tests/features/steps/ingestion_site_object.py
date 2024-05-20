from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity, Site
from steps.utils import get_object_state, ingester, send_delete_request

endpoint = "dcim/sites/"


@given('a site "{site_name}"')
def set_site(context, site_name):
    """Set the site name."""
    context.site_name = site_name


@given('the site status "{status}"')
def set_status(context, status):
    """Set the status of the site."""
    context.status = status


@given('the site description "{description}"')
def set_description(context, description):
    """Set the description of the site."""
    context.description = description


@given('site "{site_name}" does not exist')
def ensure_site_does_not_exists(context, site_name):
    """Ensure that the site does not exist."""
    site = get_object_state(
        {
            "object_type": "dcim.site",
            "q": site_name,
        },
    )
    if site:
        send_delete_request(endpoint, site.get("id"))


@when("the site is ingested")
def ingest_site(context):
    """Ingest the site using the Diode SDK"""
    entities = [
        Entity(site=Site(name=context.site_name)),
    ]
    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@then("the site is found")
def assert_site_exists(context):
    """Assert that the site was created."""
    assert context.response is not None

    params = {
        "object_type": "dcim.site",
        "q": context.site_name,
    }
    if hasattr(context, "status"):
        params["status"] = context.status
    if hasattr(context, "description"):
        params["description"] = context.description

    site = get_object_state(params)
    assert site.get("name") == context.site_name

    context.existing_site = site


@when("the site with status and description is ingested")
def ingest_site_with_status_and_description(context):
    """Ingest the site using the Diode SDK"""
    entities = [
        Entity(
            site=Site(
                name=context.site_name,
                status=context.status,
                description=context.description,
            ),
        ),
    ]
    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@then('the site status is "{status}"')
def assert_site_status(context, status):
    """Assert that the site status is correct."""
    assert context.existing_site is not None
    assert context.existing_site.get("status") == status


@then("the site description is empty")
def assert_site_description_empty(context):
    """Assert that the site description is empty."""
    assert context.existing_site is not None
    assert context.existing_site.get("description") == ""


@then('the site description is "{description}"')
def assert_site_description(context, description):
    """Assert that the site description is correct."""
    assert context.existing_site is not None
    assert context.existing_site.get("description") == description


@then("the site remains the same")
def assert_site_remains(context):
    """Assert that the site remains the same."""
    assert context.existing_site is not None
    assert context.existing_site.get("name") == context.site_name
    assert context.existing_site.get("status") == context.status
    assert context.existing_site.get("description") == context.description
