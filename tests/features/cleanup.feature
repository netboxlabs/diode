Feature: Cleanup tests objects


@cleanup
Scenario: Cleanup
    Given the device "router01" is deleted
    Given the site object "Site A" is deleted
    Given the device "router01" is deleted
    Given the device type object "ISR4321" is deleted
    Given the device role object "WAN Router" is deleted
    Given the platform object "Cisco IOS 15.6" is deleted
    Given the manufacturer object "Cisco" is deleted

    Given the site object "undefined" is deleted
    Given the device type object "undefined" is deleted
    Given the device role object "undefined" is deleted
    Given the site object "undefined" is deleted
    Given the platform object "undefined" is deleted
    Given the manufacturer object "undefined" is deleted

#
#@cleanup
#Scenario: Cleanup of device role object
#    Given the device role object "WAN Router" is deleted
#    Then the device role object is removed from the database
#
#@cleanup
#Scenario: Cleanup of manufacturer object
#    Given the manufacturer object "Cisco" is deleted
#    Then the manufacturer object is removed from the database
#
#
#@cleanup
#Scenario: Cleanup of device type object
#    Given the device type object "ISR4321" is deleted
#    Given the manufacturer object "Cisco" is deleted
#    Given the manufacturer object "undefined" is deleted
#    Then the device type object is removed from the database
#    Then the manufacturer object is removed from the database