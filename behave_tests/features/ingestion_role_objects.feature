Feature: Tests for ingestion of device role objects
    Validate the behavior of the ingestion of device role objects

@smoke
@ingestion.device_role
Scenario: Ingestion of new device role
    Given a new device role "WAN Router"
    When the device role is ingested
    Then the device role is created in the database

@smoke
@ingestion.device_role
Scenario: Ingestion of existing device role
    Given device role "WAN Router" exists in the database
    When the device role is ingested
    Then the device role remains the same


@smoke
@ingestion.device_role
Scenario: Ingestion of device role to update color and description
    Given device role "WAN Router" with color "509415" and description "some string"
    When the device role is ingested with the updates
    Then the device role is updated in the database


