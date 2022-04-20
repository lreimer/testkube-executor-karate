Feature: Testing the Chuck Norris Joke API

  Background:
    * url 'http://api.icndb.com/jokes/'

  Scenario: Testing random jokes GET endpoint
    Given url 'http://api.icndb.com/jokes/random/'
    When method GET
    Then status 500
