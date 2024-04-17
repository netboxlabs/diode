Feature: Cleanup tests objects


@cleanup
Scenario: Cleanup of site object
    Given the site object "Site A" is deleted
    Then the site object is removed from the database

@cleanup
Scenario: Cleanup of device role object
    Given the device role object "WAN Router" is deleted
    Then the device role object is removed from the database

@cleanup
Scenario: Cleanup of manufacturer object
    Given the manufacturer object "Cisco" is deleted
    Then the manufacturer object is removed from the database


@cleanup
Scenario: Cleanup of device type object
    Given the device type object "ISR4321" is deleted
    Given the manufacturer object "Cisco" is deleted
    Given the manufacturer object "undefined" is deleted
    Then the device type object is removed from the database
    Then the manufacturer object is removed from the database