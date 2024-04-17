Feature: Tests for ingestion of device objects
    Validate the behavior of the ingestion of device objects

@smoke
@ingestion.device
Scenario: Ingestion of new device object
  Given a new device "router01" object
  When the device object is ingested
  Then the device, device_type "undefined" , role "undefined", and site "undefined" are created

