import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import Entity, ClusterType
from steps.utils import get_object_by_name, ingester

endpoint = "virtualization/cluster-types/"


@given('a new cluster type "{cluster_type_name}"')
def step_create_new_cluster_type(context, cluster_type_name):
    """Set the body of the request to create a new cluster type."""
    context.cluster_type_name = cluster_type_name


@when("the cluster type is ingested")
def ingest_cluster_type(context):
    """Ingest the cluster type using the Diode SDK"""

    entities = [
        Entity(cluster_type=ClusterType(name=context.cluster_type_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then("the cluster type is created in the database")
@then("the cluster type remains the same")
def check_cluster_type_(context):
    """Check if the response is not None and the is created in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster_type = get_object_by_name(context.cluster_type_name, endpoint)
    assert cluster_type.get("name") == context.cluster_type_name
    assert cluster_type.get("slug") == "vmware"


@given('cluster type "{cluster_type_name}" already exists in the database')
def retrieve_existing_cluster_type(context, cluster_type_name):
    """Retrieve the cluster type from the database"""
    context.cluster_type_name = cluster_type_name
    cluster_type = get_object_by_name(context.cluster_type_name, endpoint)
    context.cluster_type_name = cluster_type.get("name")


@given('cluster type "{cluster_type_name}" with description "{description}"')
def create_cluster_type_to_update(context, cluster_type_name, description):
    """Create a cluster type with a description to update"""
    context.cluster_type_name = cluster_type_name
    context.description = description


@when("the cluster type is ingested with the updates")
def ingest_to_update_cluster_type(context):
    """Update the cluster type using the Diode SDK"""

    entities = [
        Entity(
            cluster_type=ClusterType(
                name=context.cluster_type_name,
                description=context.description,
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then("the cluster type is updated in the database")
def check_cluster_type_updated(context):
    """Check if the response is not None and the is updated in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster_type = get_object_by_name(context.cluster_type_name, endpoint)
    assert cluster_type.get("name") == context.cluster_type_name
    assert cluster_type.get("description") == context.description
