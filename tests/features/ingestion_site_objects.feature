Feature: Tests for ingestion of site
    Validate the behavior of the ingestion of site

@smoke
@ingestion.site
Scenario: Ingestion of a new site
    Given a site "Site A"
        And site "Site A" does not exist
    When the site is ingested
    Then the site is found
        And the site status is "active"
        And the site description is empty

@smoke
@ingestion.site
Scenario: Ingestion of existing site - nothing to update
    Given a site "Site A"
    When the site is ingested
    Then the site is found
        And the site status is "active"
        And the site description is empty


@smoke
@ingestion.site
Scenario: Ingestion of existing site with updated status and description
    Given a site "Site A"
        And the site status "planned"
        And the site description "some string"
    When the site with status and description is ingested
    Then the site is found
        And the site status is "planned"
        And the site description is "some string"


