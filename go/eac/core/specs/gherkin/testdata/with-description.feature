Feature: test-module_user-authentication

  Rule: User login authentication
    @Manual @ov @control:ac-2
    Scenario: User login with valid credentials
      This scenario tests the happy path for user authentication
      with valid username and password.

      Given a user with username "testuser" and password "password123"
      When the user attempts to login
      Then the user should be successfully authenticated
      And a session token should be created
