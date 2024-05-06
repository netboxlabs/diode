Feature: Tests for IP address ingestion
    Validate the behavior of the ingestion of the IP address

@smoke
@ingestion.ipaddress
Scenario: Ingest a new IP address
  Given an IP address "192.168.0.1/32"
    And interface "GigabitEthernet0/0/0"
  When the IP address is ingested
    Then the IP address is found
    And the IP address is associated with the interface
    And the IP address status is "active"

@smoke
@ingestion.ipaddress
Scenario: Ingest an IP address with updates
  Given an IP address "192.168.0.1/32"
      And interface "GigabitEthernet0/0/0"
      And description "lorem ipsum"
  When the IP address with description is ingested
    Then the IP address is found
    And the IP address is associated with the interface
    And the IP address description is updated
