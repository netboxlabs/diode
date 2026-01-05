Feature: Tests for ingestion of virtual interface
    Validate the behavior of the ingestion of virtual interface

@smoke
@ingestion.virtual_interface
Scenario: Ingestion of new virtual interface
    Given a new virtual interface "eth0"
    When the virtual interface is ingested
    Then the virtual interface and "undefined" virtual machine with "undefined" site are created in the database

@smoke
@ingestion.virtual_interface
Scenario: Ingestion of existing virtual interface
    Given virtual interface "eth0" with "undefined" virtual machine and "undefined" site already exists in the database
    When the virtual interface is ingested
    Then the virtual interface remains the same

@smoke
@ingestion.virtual_interface
Scenario: Ingestion of virtual interface object to update the virtual machine, description and MTU
    Given virtual interface "eth0" with virtual machine "VM-1", description "some string" and MTU "1500"
    Then check if the virtual machine "VM-1" related interface exists in the database and remove it
    When the virtual interface object is ingested with the updates
    Then the virtual machine "VM-1" is created and the interface updated
