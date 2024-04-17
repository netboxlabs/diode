Feature: Tests for ingestion of device role objects
    Validate the behavior of the ingestion of device role objects

@smoke
@ingestion.device_role
Scenario: Ingestion of new device role object
    Given a new device role "WAN Router" object
    When the device role object is ingested
    Then the device role object is created in the database

@smoke
@ingestion.device_role
Scenario: Ingestion of existing device role object
    Given device role "WAN Router" exists in the database
    When the device role object is ingested
    Then the device role object remains the same


@smoke
@ingestion.device_role
Scenario: Ingestion of device role object to change existing device role object
    Given device role "WAN Router" with color "509415" and description "some string"
    When the device role object is ingested with the updates
    Then the device role object is updated in the database

