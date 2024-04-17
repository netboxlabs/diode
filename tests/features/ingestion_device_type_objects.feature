Feature: Tests for ingestion of device type objects
    Validate the behavior of the ingestion of device type objects

@smoke
@ingestion.device_type
Scenario: Ingestion of new device type object
    Given a new device type "ISR4321" object
    When the device type object is ingested
    Then the device type object and "undefined" manufacturer are created in the database

@smoke
@ingestion.device_type
Scenario: Ingestion of existing device type object
    Given device type "ISR4321" already exists in the database
    When the device type object is ingested
    Then the device type object remains the same


@smoke
@ingestion.device_type
Scenario: Ingestion of device type object to change existing device type object
    Given device type "ISR4321" with manufacturer "Cisco", description "some string", and part number "xyz123"
    Then check if the manufacturer "Cisco" exists in the database and remove it
    When the device type object is ingested with the updates
    Then the device type object is updated and the manufacturer "Cisco" is created
