@fixture.site.cleanup
@fixture.site.update
Feature: Apply Change Set endpoint tests
  Validate the behaviors expected for the endpoint apply-change-set

  @smoke
  Scenario: Correct payload to create a new site
    Given I provide a correct payload to create a new site
    When I send a POST request to the endpoint
    Then I must get a response with status code 200 and a Json object with success message

  @smoke
  Scenario: Incorrect payload sent to create a new site
    Given I provide payload with object_type missing
    When I send a POST request to the endpoint
    Then I must get a response with status code 400 and a Json object with error message

  @smoke
  Scenario: Correct payload to update a site
    Given I provide a correct payload to update the slug of the site
    When I send a POST request to the endpoint
    Then I must get a response with status code 200 and a Json object with success message

  @smoke
  Scenario: Incorrect payload to update a site
    Given I provide payload with object_type missing for update
    When I send a POST request to the endpoint
    Then I must get a response with status code 400 and a Json object with error message