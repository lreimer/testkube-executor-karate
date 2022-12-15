Feature: Testing the Chuck Norris Joke API

  Background:
    * url 'https://api.chucknorris.io/jokes/'

  Scenario: Testing random jokes GET endpoint
    Given url 'https://api.chucknorris.io/jokes/random/'
    When method GET
    Then status 500
