Feature: Tests for ingestion of device objects
    Validate the behavior of the ingestion of device objects

@smoke
@ingestion.device
Scenario: Ingestion of new device object
  Given a new device "router01" object
  When the device object is ingested
  Then the device, device type "undefined" , role "undefined", and site "undefined" are created

@smoke
@ingestion.device
Scenario: Change the device object
  Given device type "ISR4321", role "WAN Router", and site "Site A" for device "router01"
  When the device object is ingested with the updates
  Then device type, role and site are created
  Then the device is updated


@smoke
@ingestion.device
Scenario: Change device platform
Given platform "Cisco IOS 15.6", device type "ISR4321", role "WAN Router", and site "Site A" for device "router01"
  When the device object is ingested with the platform update
  Then the device is updated and platform is created
#  Given the device "router01" is deleted
#  Given the device type object "ISR4321" is deleted
#  Given the device role object "WAN Router" is deleted
#  Given the platform object "Cisco IOS 15.6" is deleted
#  Given the site object "Site A" is deleted
#
#  Given the site object "undefined" is deleted
#  Given the device type object "undefined" is deleted
#  Given the device role object "undefined" is deleted
#  Given the site object "undefined" is deleted
#  Given the platform object "undefined" is deleted