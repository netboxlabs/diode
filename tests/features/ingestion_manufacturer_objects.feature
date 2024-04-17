Feature: Tests for ingestion of manufacturer objects
    Validate the behavior of the ingestion of manufacturer objects

@smoke
@ingestion.manufacturer
Scenario: Ingestion of new manufacturer object
    Given a new manufacturer "Cisco" object
    When the manufacturer object is ingested
    Then the manufacturer object is created in the database

@smoke
@ingestion.manufacturer
Scenario: Ingestion of existing manufacturer object
    Given manufacturer "Cisco" already exists in the database
    When the manufacturer object is ingested
    Then the manufacturer object remains the same


@smoke
@ingestion.manufacturer
Scenario: Ingestion of manufacturer object to change existing manufacturer object
    Given manufacturer "Cisco" with description "some string"
    When the manufacturer object is ingested with the updates
    Then the manufacturer object is updated in the database
