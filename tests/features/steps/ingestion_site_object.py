import time

from behave import given, when, then
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.site_pb2 import Site
from steps.utils import get_object_by_name, ingester

endpoint = "dcim/sites/"


@given('a new site "{site_name}"')
def step_create_new_site(context, site_name):
    """Set the body of the request to create a new site."""
    context.site_name = site_name


@when("the site is ingested")
def ingest_site(context):
    """Ingest the site using the Diode SDK"""
    entities = [
        Entity(site=Site(name=context.site_name)),
    ]
    context.response = ingester(entities)
    return context.response


@then("the site is created in the database")
@then("the site remains the same")
def check_site(context):
    """Check if the response is not None and the is created in the database."""
    time.sleep(3)
    assert context.response is not None
    site = get_object_by_name(context.site_name, endpoint)
    assert site.get("name") == context.site_name
    assert site.get("status").get("value") == "active"


@given('site "{site_name}" already exists in the database')
def retrieve_existing_site(context, site_name):
    """Retrieve the site from the database"""
    context.site_name = site_name
    context.site = get_object_by_name(context.site_name, endpoint)
    context.site_name = context.site.get("name")


@given('site "{site_name}" with status "{status}" and description "{description}"')
def create_site_to_update(context, site_name, status, description):
    """Create a site with a status and description to update"""
    context.site_name = site_name
    context.status = status
    context.description = description


@when("the site is ingested with the updates")
def ingest_to_update_site(context):
    """Update the site using the Diode SDK"""
    entities = [
        Entity(
            site=Site(
                name=context.site_name,
                status=context.status,
                description=context.description,
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the site is updated in the database")
def check_site_updated(context):
    """Check if the is updated in the database."""
    time.sleep(3)
    assert context.response is not None
    site = get_object_by_name(context.site_name, endpoint)
    assert site.get("name") == context.site_name
    assert site.get("status").get("value") == context.status
    assert site.get("description") == context.description
