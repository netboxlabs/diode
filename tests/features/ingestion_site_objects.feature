Feature: Tests for ingestion of site objects
    Validate the behavior of the ingestion of site objects

@smoke
@ingestion.site
Scenario: Ingestion of new site object
    Given a new site "Site A" object
    When the site object is ingested
    Then the site object is created in the database

@smoke
@ingestion.site
Scenario: Ingestion of existing site object
    Given site "Site A" already exists in the database
    When the site object is ingested
    Then the site object remains the same


@smoke
@ingestion.site
Scenario: Ingestion of site object to change existing site object
    Given site "Site A" with status "planned" and description "some string"
    When the site object is ingested with the updates
    Then the site object is updated in the database


