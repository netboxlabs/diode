Feature: Cleanup tests


@cleanup
Scenario: Cleanup
    Given the device "router01" is deleted
    Given the site "Site A" is deleted
    Given the device "router01" is deleted
    Given the device type "ISR4321" is deleted
    Given the device role "WAN Router" is deleted
    Given the platform "Cisco IOS 15.6" is deleted
    Given the manufacturer "Cisco" is deleted

    Given the site "undefined" is deleted
    Given the device type "undefined" is deleted
    Given the device role "undefined" is deleted
    Given the site "undefined" is deleted
    Given the platform "undefined" is deleted
    Given the manufacturer "undefined" is deleted
