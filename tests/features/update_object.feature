Feature: Update objects using Apply Change Set endpoint tests
  Validate the behaviors expected for the endpoint apply-change-set to update objects

  @smoke
  @update.object
  Scenario: Correct payload to update a site
    Given I provide a correct payload to update the slug of the site
    When I send a POST request to the endpoint
    Then I must get a response with status code 200 and a Json object with success message

  @smoke
  @update.object
  Scenario: Incorrect payload to update a site
    Given I provide payload with object_type missing for update
    When I send a POST request to the endpoint
    Then I must get a response with status code 400 and a Json object with error message