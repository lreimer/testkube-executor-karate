Feature: Testing the Chuck Norris Joke API - part 1

  Scenario: Testing random jokes GET endpoint
    Given url 'https://api.chucknorris.io/jokes/random/'
    When method GET
    Then status 200

  @ignore
  Scenario: This should be ignored
      Given url 'https://api.chucknorris.io/jokes/random/'
      When method GET
      Then status 404