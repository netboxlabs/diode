Feature: Tests for prefixes ingestion
    Validate the behavior of the ingestion of a prefix

@smoke
@ingestion.prefix
Scenario: Ingest a new prefix
  Given a new prefix "192.168.0.0/32"
  When the prefix is ingested
    Then the prefix is found
    And the prefix is associated with the site "undefined"
    And the prefix is active

@smoke
@ingestion.prefix
Scenario: Ingest a prefix with updates
  Given a prefix "192.168.0.0/32" with description "lorem ipsum"
  When the prefix with description is ingested
    Then the prefix is found
    And the prefix description is updated
