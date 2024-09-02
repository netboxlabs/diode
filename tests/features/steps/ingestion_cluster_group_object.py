import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import Entity, ClusterGroup
from steps.utils import get_object_by_name, ingester

endpoint = "virtualization/cluster-groups/"


@given('a new cluster group "{cluster_group_name}"')
def step_create_new_cluster_group(context, cluster_group_name):
    """Set the body of the request to create a new cluster group."""
    context.cluster_group_name = cluster_group_name


@when("the cluster group is ingested")
def ingest_cluster_group(context):
    """Ingest the cluster group using the Diode SDK"""

    entities = [
        Entity(cluster_group=ClusterGroup(name=context.cluster_group_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then("the cluster group is created in the database")
@then("the cluster group remains the same")
def check_cluster_group_(context):
    """Check if the response is not None and the is created in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster_group = get_object_by_name(context.cluster_group_name, endpoint)
    assert cluster_group.get("name") == context.cluster_group_name
    assert cluster_group.get("slug") == "north-america"


@given('cluster group "{cluster_group_name}" already exists in the database')
def retrieve_existing_cluster_group(context, cluster_group_name):
    """Retrieve the cluster group from the database"""
    context.cluster_group_name = cluster_group_name
    cluster_group = get_object_by_name(context.cluster_group_name, endpoint)
    context.cluster_group_name = cluster_group.get("name")


@given('cluster group "{cluster_group_name}" with description "{description}"')
def create_cluster_group_to_update(context, cluster_group_name, description):
    """Create a cluster group with a description to update"""
    context.cluster_group_name = cluster_group_name
    context.description = description


@when("the cluster group is ingested with the updates")
def ingest_to_update_cluster_group(context):
    """Update the cluster group using the Diode SDK"""

    entities = [
        Entity(
            cluster_group=ClusterGroup(
                name=context.cluster_group_name,
                description=context.description,
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the cluster group is updated in the database")
def check_cluster_group_updated(context):
    """Check if the response is not None and the is updated in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster_group = get_object_by_name(context.cluster_group_name, endpoint)
    assert cluster_group.get("name") == context.cluster_group_name
    assert cluster_group.get("description") == context.description
