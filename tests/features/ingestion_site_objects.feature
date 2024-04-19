Feature: Tests for ingestion of site
    Validate the behavior of the ingestion of site

@smoke
@ingestion.site
Scenario: Ingestion of new site
    Given a new site "Site A"
    When the site is ingested
    Then the site is created in the database

@smoke
@ingestion.site
Scenario: Ingestion of existing site
    Given site "Site A" already exists in the database
    When the site is ingested
    Then the site remains the same


@smoke
@ingestion.site
Scenario: Ingestion of site to update status and description
    Given site "Site A" with status "planned" and description "some string"
    When the site is ingested with the updates
    Then the site is updated in the database


