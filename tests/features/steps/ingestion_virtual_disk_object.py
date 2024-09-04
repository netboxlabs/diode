import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import VirtualDisk, Entity, VirtualMachine
from steps.utils import (
    get_object_by_name,
    send_delete_request,
    get_object_state,
    ingester,
)

endpoint = "virtualization/virtual-disks/"


@given('a new virtual disk "{virtual_disk_name}" with size "{virtual_disk_size}"')
def step_create_new_virtual_machine(context, virtual_disk_name, virtual_disk_size):
    """Set the body of the request to create a new virtual disk."""
    context.virtual_disk_name = virtual_disk_name
    context.virtual_disk_size = int(virtual_disk_size)


@when("the virtual disk is ingested")
def ingest_virtual_disk(context):
    """Ingest the virtual disk object using the Diode SDK"""

    entities = [
        Entity(
            virtual_disk=VirtualDisk(
                name=context.virtual_disk_name, size=context.virtual_disk_size
            )
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the virtual disk and "{virtual_machine_name}" virtual machine with "{site_name}" site are created in the database'
)
def check_virtual_disk_and_virtual_machine(context, virtual_machine_name, site_name):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None

    params = {
        "object_type": "virtualization.virtualdisk",
        "q": context.virtual_disk_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": site_name,
    }

    virtual_disk = get_object_state(params)

    assert virtual_disk.get("name") == context.virtual_disk_name
    assert virtual_disk.get("size") == context.virtual_disk_size
    assert virtual_disk.get("virtual_machine").get("name") == virtual_machine_name


@then("the virtual disk remains the same")
def check_virtual_disk_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None

    params = {
        "object_type": "virtualization.virtualdisk",
        "q": context.virtual_disk_name,
        "virtual_machine__name": context.virtual_machine_name,
        "virtual_machine__site__name": context.site_name,
    }

    virtual_disk = get_object_state(params)

    assert virtual_disk.get("name") == context.virtual_disk_name


@given(
    'virtual disk "{virtual_disk_name}" with "{virtual_machine_name}" virtual machine and "{site_name}" site already exists in the database'
)
def retrieve_existing_virtual_machine(
    context, virtual_disk_name, virtual_machine_name, site_name
):
    """Retrieve the virtual disk object from the database"""
    time.sleep(3)
    context.virtual_disk_name = virtual_disk_name
    context.virtual_machine_name = virtual_machine_name
    context.site_name = site_name

    params = {
        "object_type": "virtualization.virtualdisk",
        "q": context.virtual_disk_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": site_name,
    }

    virtual_disk = get_object_state(params)

    context.virtual_disk_name = virtual_disk.get("name")
    context.virtual_disk_size = virtual_disk.get("size")


@given(
    'virtual disk "{virtual_disk_name}" with virtual machine "{virtual_machine_name}", description "{description}" '
    'and size "{size}"'
)
def create_virtual_disk_to_update(
    context, virtual_disk_name, virtual_machine_name, description, size
):
    """Create a virtual disk object with a description to update"""
    context.virtual_disk_name = virtual_disk_name
    context.virtual_machine_name = virtual_machine_name
    context.description = description
    context.size = int(size)


@then(
    'check if the virtual machine "{virtual_machine_name}" related to disk exists in the database and remove it'
)
def remove_virtual_machine(context, virtual_machine_name):
    time.sleep(3)
    virtual_machine = get_object_by_name(
        virtual_machine_name, "virtualization/virtual-machines/"
    )
    if virtual_machine is not None:
        assert virtual_machine.get("name") == virtual_machine_name
        send_delete_request(
            "virtualization/virtual-machines/", virtual_machine.get("id")
        )


@when("the virtual disk object is ingested with the updates")
def ingest_to_update_virtual_disk(context):
    """Update the object using the Diode SDK"""

    entities = [
        Entity(
            virtual_disk=VirtualDisk(
                name=context.virtual_disk_name,
                virtual_machine=VirtualMachine(name=context.virtual_machine_name),
                description=context.description,
                size=context.size,
            ),
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then('the virtual machine "{virtual_machine_name}" is created and the disk updated')
def check_updated_virtual_disk_object(context, virtual_machine_name):
    """Check if the response is not None and the object is updated in the database and virtual machine created."""
    time.sleep(3)
    assert context.response is not None

    virtual_machine = get_object_by_name(
        virtual_machine_name, "virtualization/virtual-machines/"
    )
    assert virtual_machine.get("name") == virtual_machine_name

    params = {
        "object_type": "virtualization.virtualdisk",
        "q": context.virtual_disk_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": virtual_machine.get("site").get("name"),
    }

    virtual_disk = get_object_state(params)

    assert virtual_disk.get("name") == context.virtual_disk_name
    assert virtual_disk.get("virtual_machine").get("name") == virtual_machine.get(
        "name"
    )
    assert virtual_disk.get("description") == context.description
    assert virtual_disk.get("size") == context.size
