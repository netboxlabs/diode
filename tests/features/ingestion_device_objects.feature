Feature: Tests for ingestion of device
    Validate the behavior of the ingestion of device

@smoke
@ingestion.device
Scenario: Ingestion of new device
  Given a new device "router01"
  When the device is ingested
  Then the device, device type "undefined" , role "undefined", and site "undefined" are created

@smoke
@ingestion.device
Scenario: Ingestion of existing device to update its device type, site, and role
  Given device type "ISR4321", role "WAN Router", and site "Site A" for device "router01"
  When the device is ingested with the updates
  Then device type, role and site are created
  Then the device is updated


@smoke
@ingestion.device
Scenario: Ingestion of existing device to update its platform
Given platform "Cisco IOS 15.6", device type "ISR4321", role "WAN Router", and site "Site A" for device "router01"
  When the device is ingested with the platform update
  Then the device is updated and platform is created
  Given the device "router01" is deleted
  Given the device type "ISR4321" is deleted
  Given the device role "WAN Router" is deleted
  Given the platform "Cisco IOS 15.6" is deleted
  Given the site "Site A" is deleted

  Given the site "undefined" is deleted
  Given the device type "undefined" is deleted
  Given the device role "undefined" is deleted
  Given the site "undefined" is deleted
  Given the platform "undefined" is deleted