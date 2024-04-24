Feature: Tests for ingestion of device type
    Validate the behavior of the ingestion of device type

@smoke
@ingestion.device_type
Scenario: Ingestion of new device type
    Given a new device type "ISR4321"
    When the device type is ingested
    Then the device type and "undefined" manufacturer are created in the database

@smoke
@ingestion.device_type
Scenario: Ingestion of existing device type
    Given device type "ISR4321" already exists in the database
    When the device type is ingested
    Then the device type remains the same


@smoke
@ingestion.device_type
Scenario: Ingestion of device type object to update the manufacturer, description, and part number
    Given device type "ISR4321" with manufacturer "Cisco", description "some string", and part number "xyz123"
    Then check if the manufacturer "Cisco" exists in the database and remove it
    When the device type object is ingested with the updates
    Then the device type object is updated and the manufacturer "Cisco" is created
