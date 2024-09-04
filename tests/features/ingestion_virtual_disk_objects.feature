Feature: Tests for ingestion of virtual disk
    Validate the behavior of the ingestion of virtual disk

@smoke
@ingestion.virtual_disk
Scenario: Ingestion of new virtual disk
    Given a new virtual disk "Disk-1" with size "1234"
    When the virtual disk is ingested
    Then the virtual disk and "undefined" virtual machine with "undefined" site are created in the database

@smoke
@ingestion.virtual_disk
Scenario: Ingestion of existing virtual disk
    Given virtual disk "Disk-1" with "undefined" virtual machine and "undefined" site already exists in the database
    When the virtual disk is ingested
    Then the virtual disk remains the same

@smoke
@ingestion.virtual_disk
Scenario: Ingestion of virtual disk object to update the virtual machine, description and size
    Given virtual disk "Disk-1" with virtual machine "VM-2", description "some string" and size "15232"
    Then check if the virtual machine "VM-2" related to disk exists in the database and remove it
    When the virtual disk object is ingested with the updates
    Then the virtual machine "VM-2" is created and the disk updated
