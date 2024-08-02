Feature: Tests for interface ingestion
    Validate the behavior of the ingestion of an interface

@smoke
@ingestion.interface
Scenario: Ingest a new interface
    Given a new interface "GigabitEthernet0/0/0"
    When the interface is ingested
      Then the interface is found
      And the interface is associated with the device "undefined"
      And the interface is enabled
      And the interface type is "other"

@smoke
@ingestion.interface
Scenario: Ingest an interface with updates
    Given an interface "GigabitEthernet0/0/0" with MTU "1500"
    When the interface with MTU is ingested
      Then the interface is found
      And the interface MTU is updated

@smoke
@ingestion.interface
Scenario: Ingestion of an existing interface disabled
    Given an interface "GigabitEthernet0/0/0" with enabled field set to "true"
    When the interface with enabled field is ingested
    Then the interface with enabled field is found
        And enabled field is "true"
    Given an interface "GigabitEthernet0/0/0" with enabled field set to "false"
    When the interface with enabled field is ingested
    Then the interface with enabled field is found
        And enabled field is "false"