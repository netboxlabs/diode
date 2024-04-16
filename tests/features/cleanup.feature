Feature: Cleanup tests objects


@cleanup
Scenario: Cleanup of site object
    Given the site object "Site A" is deleted
    Then the site object is removed from the database