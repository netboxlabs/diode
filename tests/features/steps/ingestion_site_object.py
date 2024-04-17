import time

from behave import given, when, then

from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.diode.v1.ingester_pb2 import Entity
from netboxlabs.diode.sdk.diode.v1.site_pb2 import Site

from steps.config import configs

from steps.utils import get_object_by_name, send_delete_request


api_key = str(configs["api_key"])
endpoint = "dcim/sites/"


@given('a new site "{site_name}" object')
def step_create_new_site_object(context, site_name):
    """Set the body of the request to create a new site."""
    context.site_name = site_name


@when("the site object is ingested")
def ingest_site_object(context):
    """Ingest the site object using the Diode SDK"""
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        api_key=api_key,
    ) as client:
        entities = [
            Entity(site=Site(name=context.site_name)),
        ]

        context.response = client.ingest(entities=entities)
        return context.response


@then("the site object is created in the database")
@then("the site object remains the same")
def check_site_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    site = get_object_by_name(context.site_name, endpoint)
    assert site.get("name") == context.site_name
    assert site.get("status").get("value") == "active"


@given('site "{site_name}" already exists in the database')
def retrieve_existing_site_object(context, site_name):
    """Retrieve the site object from the database"""
    context.site_name = site_name
    context.site = get_object_by_name(context.site_name, endpoint)
    context.site_name = context.site.get("name")


@given('site {site_name} with status "{status}" and description "{description}"')
def create_site_object_to_update(context, site_name, status, description):
    """Create a site object with a status and description to update"""
    context.site_name = site_name
    context.status = status
    context.description = description


@when("the site object is ingested with the updates")
def update_site_object(context):
    """Update the site object using the Diode SDK"""
    with DiodeClient(
        target="localhost:8081",
        app_name="my-test-app",
        app_version="0.0.1",
        api_key=api_key,
    ) as client:
        entities = [
            Entity(
                site=Site(
                    name=context.site_name,
                    status=context.status,
                    description=context.description,
                )
            ),
        ]

        context.response = client.ingest(entities=entities)
        return context.response


@then("the site object is updated in the database")
def check_site_object_updated(context):
    """Check if the object is updated in the database."""
    assert context.response is not None
    site = get_object_by_name(context.site_name, endpoint)
    assert site.get("name") == context.site_name
    assert site.get("status").get("value") == context.status
    assert site.get("description") == context.description
