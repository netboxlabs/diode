Feature: Object State endpoint tests
  Validate the behaviors expected for the endpoint object-state

  @smoke
  @object.state
  Scenario: Return object state using id
    Given the site id "Site Z" and object_type "dcim.site"
    Then the object state "Site Z" is returned successfully

  @smoke
  @object.state
  Scenario: Return object state using q parameter
    Given the site name "Site X" and object_type "dcim.site"
    Then the object state "Site X" is returned successfully

  @smoke
  @object.state
  Scenario: Return none object state using q parameter
    Given the site name "Site Does Not Exist" and object_type "dcim.site"
    Then endpoint return 200 and empty response

  @smoke
  @object.state
  Scenario: Return 400 when object type is not provided
    Given the site name "Site X" and not object_type
    Then endpoint return 400
