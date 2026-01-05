Feature: Tests for ingestion of manufacturer
    Validate the behavior of the ingestion of manufacturer

@smoke
@ingestion.manufacturer
Scenario: Ingestion of new manufacturer
    Given a new manufacturer "Cisco"
    When the manufacturer is ingested
    Then the manufacturer is created in the database

@smoke
@ingestion.manufacturer
Scenario: Ingestion of existing manufacturer
    Given manufacturer "Cisco" already exists in the database
    When the manufacturer is ingested
    Then the manufacturer remains the same


@smoke
@ingestion.manufacturer
Scenario: Ingestion of manufacturer to update its description
    Given manufacturer "Cisco" with description "some string"
    When the manufacturer is ingested with the updates
    Then the manufacturer is updated in the database
