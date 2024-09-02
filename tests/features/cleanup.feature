Feature: Cleanup tests


@cleanup
Scenario: Cleanup
    Given the device "router01" is deleted
    Given the site "Site A" is deleted
    Given the site "Site B" is deleted
    Given the device "router01" is deleted
    Given the device type "ISR4321" is deleted
    Given the device role "WAN Router" is deleted
    Given the platform "Cisco IOS 15.6" is deleted
    Given the manufacturer "Cisco" is deleted
    Given the cluster "aws-us-east-1" is deleted

    Given the site "undefined" is deleted
    Given the device type "undefined" is deleted
    Given the device role "undefined" is deleted
    Given the site "undefined" is deleted
    Given the platform "undefined" is deleted
    Given the manufacturer "undefined" is deleted

    Given the interface "GigabitEthernet0/0/0" is deleted

    Given the IP address "192.168.0.1/32" is deleted
    Given the prefix "192.168.0.0/32" is deleted

    Given the cluster type "VMWare" is deleted
    Given the cluster group "AWS" is deleted

    Given the tag "tag 100" is deleted
