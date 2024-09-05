import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import (
    Cluster,
    Entity,
    ClusterGroup,
    ClusterType,
    Site,
)
from steps.utils import (
    get_object_by_name,
    get_object_state,
    send_delete_request,
    ingester,
)

endpoint = "virtualization/clusters/"


@given('a new cluster "{cluster_name}"')
def step_create_new_cluster(context, cluster_name):
    """Set the body of the request to create a new cluster."""
    context.cluster_name = cluster_name


@when("the cluster is ingested")
def ingest_cluster(context):
    """Ingest the cluster object using the Diode SDK"""

    entities = [
        Entity(cluster=Cluster(name=context.cluster_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the cluster, "{group_name}" group, "{type_name}" type and "{site_name}" site are created in the database'
)
def check_cluster_groups_and_types(context, group_name, type_name, site_name):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster = get_object_by_name(context.cluster_name, endpoint)
    cluster_group = get_object_by_name(group_name, "virtualization/cluster-groups/")
    cluster_type = get_object_by_name(type_name, "virtualization/cluster-types/")
    site = get_object_by_name(site_name, "dcim/sites/")
    assert cluster.get("name") == context.cluster_name
    assert cluster_group.get("name") == group_name
    assert cluster_type.get("name") == type_name
    assert site.get("name") == site_name


@then("the cluster remains the same")
def check_cluster_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None
    cluster = get_object_by_name(context.cluster_name, endpoint)
    assert cluster.get("name") == context.cluster_name


@given('cluster "{cluster_name}" already exists in the database')
def retrieve_existing_cluster(context, cluster_name):
    """Retrieve the cluster object from the database"""
    time.sleep(3)
    context.cluster_name = cluster_name
    cluster = get_object_by_name(context.cluster_name, endpoint)
    context.cluster_name = cluster.get("name")


@given(
    'cluster "{cluster_name}" with group "{group_name}", type "{type_name}", '
    'site "{site_name}" and description "{description}"'
)
def create_cluster_to_update(
    context, cluster_name, group_name, type_name, site_name, description
):
    """Create a cluster object with a description to update"""
    context.cluster_name = cluster_name
    context.group_name = group_name
    context.type_name = type_name
    context.site_name = site_name
    context.description = description


@then(
    'check if the group "{group_name}", type "{type_name}" and site "{site_name}" '
    "exist in the database and remove them"
)
def remove_cluster_objects(context, group_name, type_name, site_name):
    time.sleep(3)
    cluster_group = get_object_by_name(group_name, "virtualization/cluster-groups/")
    if cluster_group is not None:
        assert cluster_group.get("name") == group_name
        send_delete_request("virtualization/cluster-groups/", cluster_group.get("id"))

    cluster_type = get_object_by_name(type_name, "virtualization/cluster-types/")
    if cluster_type is not None:
        assert cluster_type.get("name") == type_name
        send_delete_request("virtualization/cluster-types/", cluster_type.get("id"))

    site = get_object_by_name(site_name, "dcim/sites/")
    if site is not None:
        assert site.get("name") == site_name
        send_delete_request("dcim/sites/", site.get("id"))


@when("the cluster object is ingested with the updates")
def ingest_to_update_cluster(context):
    """Update the object using the Diode SDK"""

    entities = [
        Entity(
            cluster=Cluster(
                name=context.cluster_name,
                group=ClusterGroup(name=context.group_name),
                type=ClusterType(name=context.type_name),
                site=Site(name=context.site_name),
                description=context.description,
            ),
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the group "{group_name}", type "{type_name}" and site "{site_name}" are created and the cluster updated'
)
def check_updated_cluster_objects(context, group_name, type_name, site_name):
    """Check if the response is not None and the object is updated in the database and cluster objects created."""
    time.sleep(3)
    assert context.response is not None

    cluster_group = get_object_by_name(group_name, "virtualization/cluster-groups/")
    assert cluster_group.get("name") == group_name

    cluster_type = get_object_by_name(type_name, "virtualization/cluster-types/")
    assert cluster_type.get("name") == type_name

    site = get_object_by_name(site_name, "dcim/sites/")
    assert site.get("name") == site_name
    
    params = {
        "object_type": "virtualization.cluster",
        "q": context.cluster_name,
        "site__name": context.site_name,
    }

    cluster = get_object_state(params)

    assert cluster.get("name") == context.cluster_name
    assert cluster.get("group").get("name") == cluster_group.get("name")
    assert cluster.get("type").get("name") == cluster_type.get("name")
    assert cluster.get("site").get("name") == site.get("name")
    assert cluster.get("description") == context.description
