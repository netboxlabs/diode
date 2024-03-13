Feature: auth

Scenario: Unauthenticated user can't access the page
Given I am not authenticated
When I access the page
Then Status code is 302


#Scenario: Authenticated user can access the page
#Given I am authenticated
#When I access the page
#Then Status code is 200