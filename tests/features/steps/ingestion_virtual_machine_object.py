from behave import given, when, then
from netboxlabs.diode.sdk.ingester import (
    VirtualMachine,
    Entity,
)
from steps.utils import (
    get_object_state,
    ingester,
    send_delete_request,
)


endpoint = "virtualization/virtual-machines/"


@given('virtual machine "{virtual_machine_name}" with site not provided')
def set_virtual_machine_without_site(context, virtual_machine_name):
    """Set the body of the request to ingest the virtual machine."""
    context.virtual_machine_name = virtual_machine_name
    context.site_name = "undefined"
    context.virtual_machine_role_name = "undefined"

@given('virtual machine "{virtual_machine_name}" with site "{site_name}" does not exist')
def ensure_virtual_machine_does_not_exists(context, virtual_machine_name, site_name):
    """Ensure that the virtual machine does not exist."""
    virtual_machine = get_object_state(
        {
            "object_type": "virtualization.virtualmachine",
            "q": virtual_machine_name,
            "site__name": site_name,
        },
    )
    if virtual_machine:
        send_delete_request(endpoint, virtual_machine.get("id"))


@when("the virtual machine without site is ingested")
def ingest_virtual_machine_without_site(context):
    """Ingest the virtual machine using the Diode SDK"""

    entities = [
        Entity(virtual_machine=context.virtual_machine_name),
    ]

    response = ingester(entities)
    assert response.errors == []

    context.response = response
    return response


@then("the virtual machine is found")
def assert_virtual_machine_exists(context):
    """Assert that the virtual machine was created."""
    assert context.response is not None

    params = {
        "object_type": "virtualization.virtualmachine",
        "q": context.virtual_machine_name,
        "site__name": context.site_name,
    }
    
    virtual_machine = get_object_state(params)

    assert virtual_machine.get("name") == context.virtual_machine_name
    assert virtual_machine.get("site").get("name") == context.site_name
    context.existing_virtual_machine = virtual_machine


@then('device role is "{virtual_machine_role_name}"')
def assert_virtual_machine_role(context, virtual_machine_role_name):
    """Assert that the virtual machine role is correct."""
    assert context.existing_virtual_machine is not None
    assert context.existing_virtual_machine.get("role").get("name") == virtual_machine_role_name


@given('virtual machine "{virtual_machine_name}" with site "{site_name}" exists')
def assert_virtual_machine_exists_with_site(context, virtual_machine_name, site_name):
    """Assert that the virtual machine exists."""
    virtual_machine = get_object_state(
        {
            "object_type": "virtualization.virtualmachine",
            "q": virtual_machine_name,
            "site__name": site_name,
        },
    )

    assert virtual_machine.get("name") == virtual_machine_name
    assert virtual_machine.get("site").get("name") == site_name
    context.existing_virtual_machine = virtual_machine


@given('a new virtual machine "{virtual_machine_name}" with site "{site_name}"')
def create_new_virtual_machine_with_site(context, virtual_machine_name, site_name):
    """Set the body of the request to create a new virtual machine with site."""
    context.virtual_machine_name = virtual_machine_name
    context.site_name = site_name
    context.virtual_machine_role_name = "undefined"


@when("the virtual machine with site is ingested")
def ingest_virtual_machine_with_site(context):
    """Ingest the virtual machine using the Diode SDK"""
    entities = [
        Entity(
            virtual_machine=VirtualMachine(
                name=context.virtual_machine_name,
                site=context.site_name,
            ),
        ),
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@given(
    'virtual machine "{virtual_machine_name}" with site "{site_name}" and role "{virtual_machine_role_name}"'
)
def update_virtual_machine(context, virtual_machine_name, site_name, virtual_machine_role_name):
    """Set the body of the request to update a virtual machine."""
    context.virtual_machine_name = virtual_machine_name
    context.site_name = site_name
    context.virtual_machine_role_name = virtual_machine_role_name


@when("the virtual machine with site and device role is ingested")
def ingest_virtual_machine_with_site_device_role(context):
    """Ingest the virtual machine using the Diode SDK"""
    entities = [
        Entity(
            virtual_machine=VirtualMachine(
                name=context.virtual_machine_name,
                site=context.site_name,
                role=context.virtual_machine_role_name,
            ),
        ),
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response

@given('virtual machine "{virtual_machine_name}" with "{description}" description')
def set_virtual_machine_with_description(context, virtual_machine_name, description):
    """Set the body of the request to ingest the virtual machine."""
    context.virtual_machine_name = virtual_machine_name
    context.site_name = "undefined"
    if description == "empty":
        description = ""
    context.description = description


@when("the virtual machine with description is ingested")
def ingest_virtual_machine_with_description(context):
    """Ingest the virtual machine using the Diode SDK"""
    entities = [
        Entity(
            virtual_machine=VirtualMachine(
                name=context.virtual_machine_name,
                description=context.description,
            ),
        )
    ]

    context.response = ingester(entities)
    assert context.response.errors == []

    return context.response


@then('the virtual machine with ingested "{field_name}" field is found')
def assert_virtual_machine_with_ingested_field_is_found(context, field_name):
    """Assert that the virtual machine exists."""
    assert context.response is not None

    params = {
        "object_type": "virtualization.virtualmachine",
        "q": context.virtual_machine_name,
        "site__name": context.site_name,
        field_name: getattr(context, field_name),
    }

    if hasattr(context, "virtual_machine_role_name"):
        params["role__name"] = context.virtual_machine_role_name

    virtual_machine = get_object_state(params)

    assert virtual_machine.get("name") == context.virtual_machine_name
    assert virtual_machine.get("site").get("name") == context.site_name

    context.existing_virtual_machine = virtual_machine


@then('vm description is "{description}"')
def assert_description(context, description):
    """Assert that the description is correct."""
    if description == "empty":
        description = ""
    assert context.existing_virtual_machine is not None
    assert context.existing_virtual_machine.get("description") == description
