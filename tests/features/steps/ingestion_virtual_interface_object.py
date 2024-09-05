import time

from behave import given, when, then
from netboxlabs.diode.sdk.ingester import VMInterface, Entity, VirtualMachine
from steps.utils import (
    get_object_by_name,
    send_delete_request,
    get_object_state,
    ingester,
)

endpoint = "virtualization/virtual-interfaces/"


@given('a new virtual interface "{vminterface_name}"')
def step_create_new_virtual_machine(context, vminterface_name):
    """Set the body of the request to create a new virtual interface."""
    context.vminterface_name = vminterface_name


@when("the virtual interface is ingested")
def ingest_virtual_interface(context):
    """Ingest the virtual interface object using the Diode SDK"""

    entities = [
        Entity(vminterface=VMInterface(name=context.vminterface_name)),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the virtual interface and "{virtual_machine_name}" virtual machine with "{site_name}" site are created in the database'
)
def check_virtual_interface_and_virtual_machine(
    context, virtual_machine_name, site_name
):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None

    params = {
        "object_type": "virtualization.vminterface",
        "q": context.vminterface_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": site_name,
    }

    virtual_interface = get_object_state(params)

    assert virtual_interface.get("name") == context.vminterface_name
    assert virtual_interface.get("virtual_machine").get("name") == virtual_machine_name


@then("the virtual interface remains the same")
def check_virtual_interface_object(context):
    """Check if the response is not None and the object is created in the database."""
    time.sleep(3)
    assert context.response is not None

    params = {
        "object_type": "virtualization.vminterface",
        "q": context.vminterface_name,
        "virtual_machine__name": context.virtual_machine_name,
        "virtual_machine__site__name": context.site_name,
    }

    virtual_interface = get_object_state(params)

    assert virtual_interface.get("name") == context.vminterface_name


@given(
    'virtual interface "{vminterface_name}" with "{virtual_machine_name}" virtual machine and "{site_name}" site already exists in the database'
)
def retrieve_existing_virtual_machine(
    context, vminterface_name, virtual_machine_name, site_name
):
    """Retrieve the virtual interface object from the database"""
    time.sleep(3)
    context.vminterface_name = vminterface_name
    context.virtual_machine_name = virtual_machine_name
    context.site_name = site_name

    params = {
        "object_type": "virtualization.vminterface",
        "q": context.vminterface_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": site_name,
    }

    virtual_interface = get_object_state(params)

    context.vminterface_name = virtual_interface.get("name")


@given(
    'virtual interface "{vminterface_name}" with virtual machine "{virtual_machine_name}", description "{description}" '
    'and MTU "{mtu}"'
)
def create_virtual_interface_to_update(
    context, vminterface_name, virtual_machine_name, description, mtu
):
    """Create a virtual interface object with a description to update"""
    context.vminterface_name = vminterface_name
    context.virtual_machine_name = virtual_machine_name
    context.description = description
    context.mtu = int(mtu)


@then(
    'check if the virtual machine "{virtual_machine_name}" related interface exists in the database and remove it'
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


@when("the virtual interface object is ingested with the updates")
def ingest_to_update_virtual_interface(context):
    """Update the object using the Diode SDK"""

    entities = [
        Entity(
            vminterface=VMInterface(
                name=context.vminterface_name,
                virtual_machine=VirtualMachine(name=context.virtual_machine_name),
                description=context.description,
                mtu=context.mtu,
            ),
        ),
    ]

    context.response = ingester(entities)
    return context.response


@then(
    'the virtual machine "{virtual_machine_name}" is created and the interface updated'
)
def check_updated_virtual_interface_object(context, virtual_machine_name):
    """Check if the response is not None and the object is updated in the database and virtual machine created."""
    time.sleep(3)
    assert context.response is not None

    virtual_machine = get_object_by_name(
        virtual_machine_name, "virtualization/virtual-machines/"
    )
    assert virtual_machine.get("name") == virtual_machine_name

    params = {
        "object_type": "virtualization.vminterface",
        "q": context.vminterface_name,
        "virtual_machine__name": virtual_machine_name,
        "virtual_machine__site__name": virtual_machine.get("site").get("name"),
    }

    virtual_interface = get_object_state(params)

    assert virtual_interface.get("name") == context.vminterface_name
    assert virtual_interface.get("virtual_machine").get("name") == virtual_machine.get(
        "name"
    )
    assert virtual_interface.get("description") == context.description
    assert virtual_interface.get("mtu") == context.mtu
