Feature: Testing the Chuck Norris Joke API - part 2

  Scenario: Testing random career jokes GET endpoint
    Given url 'https://api.chucknorris.io/jokes/random/'
    And param category = 'career'
    When method GET
    Then status 200
    And match response contains { categories: ["career"] }