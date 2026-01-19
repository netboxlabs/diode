Feature: Tests for ingestion of virtual machine
    Validate the behavior of the ingestion of virtual machine

@smoke
@ingestion.virtual_machine
Scenario: Ingestion of a new virtual machine (site not provided)
    Given virtual machine "vm01" with site not provided
        And virtual machine "vm01" with site "undefined" does not exist
    When the virtual machine without site is ingested
    Then the virtual machine is found
        And device role is "undefined"

@smoke
@ingestion.virtual_machine
Scenario: Ingestion of existing virtual machine (site not provided)
    Given virtual machine "vm01" with site not provided
        And virtual machine "vm01" with site "undefined" exists
    When the virtual machine without site is ingested
    Then the virtual machine is found
        And device role is "undefined"

@smoke
@ingestion.virtual_machine
Scenario: Ingestion of a new virtual machine (site provided)
    Given a new virtual machine "vm02" with site "Site B"
        And virtual machine "vm02" with site "Site B" does not exist
    When the virtual machine with site is ingested
    Then the virtual machine is found
        And device role is "undefined"

@smoke
@ingestion.virtual_machine
Scenario: Ingestion of existing virtual machine (site provided) with different device role
    Given virtual machine "vm02" with site "Site B" and role "Server"
        And virtual machine "vm02" with site "Site B" exists
    When the virtual machine with site and device role is ingested
    Then the virtual machine is found
        And device role is "Server"

@smoke
@ingestion.virtual_machine
Scenario: Ingestion of existing virtual machine with description set to empty
    Given virtual machine "vm01" with "lorem ipsum" description
    When the virtual machine with description is ingested
    Then the virtual machine with ingested "description" field is found
        And vm description is "lorem ipsum"
    Given virtual machine "vm01" with "empty" description
    When the virtual machine with description is ingested
    Then the virtual machine with ingested "description" field is found
        And vm description is "empty"
