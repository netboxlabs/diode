Feature: Tests for ingestion of device
    Validate the behavior of the ingestion of device

@smoke
@ingestion.device
Scenario: Ingestion of a new device (site not provided)
    Given device "router01" with site not provided
        And device "router01" with site "undefined" does not exist
    When the device without site is ingested
    Then the device is found
        And device type is "undefined"
        And role is "undefined"

@smoke
@ingestion.device
Scenario: Ingestion of existing device (site not provided)
    Given device "router01" with site not provided
        And device "router01" with site "undefined" exists
    When the device without site is ingested
    Then the device is found
        And device type is "undefined"
        And role is "undefined"

@smoke
@ingestion.device
Scenario: Ingestion of a new device (site provided)
    Given a new device "router01" with site "Site A"
        And device "router01" with site "Site A" does not exist
    When the device with site is ingested
    Then the device is found
        And device type is "undefined"
        And role is "undefined"

@smoke
@ingestion.device
Scenario: Ingestion of existing device (site provided) with different device type and role
    Given device "router01" with site "Site A", device type "ISR4321" and role "WAN Router"
        And device "router01" with site "Site A" exists
    When the device with site, device type and role is ingested
    Then the device is found
        And device type is "ISR4321"
        And role is "WAN Router"
