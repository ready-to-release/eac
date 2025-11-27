@skip:todo
@env:isolated-test-project
@deps:go @depm:src-cli
Feature: src-cli_user-login

  As a user
  I want to login with my credentials
  So that I can access the system securely

  Rule: User successfully logs in with valid credentials

    @L2 @ov
    Scenario: User logs in with correct username and password
      Given the login interface is available
      When the user provides a valid username and password
      Then access is granted
      And the user is authenticated

    @L2 @ov
    Scenario: User is presented with login failure on incorrect credentials
      Given the login interface is available
      When the user provides an incorrect password
      Then access is denied
      And an error message is displayed

  Rule: Login process handles invalid input gracefully

    @L2 @ov
    Scenario: User attempts login with empty username
      Given the login interface is available
      When the user submits the form with an empty username
      Then the login fails
      And a validation error is shown

    @L2 @ov
    Scenario: User attempts login with empty password
      Given the login interface is available
      When the user submits the form with an empty password
      Then the login fails
      And a validation error is shown

  Rule: Login session is established upon successful authentication

    @L2 @ov
    Scenario: User session is created after login
      Given the user has provided valid credentials
      When authentication succeeds
      Then a session token is issued
      And subsequent requests use this session

    @iv @L3
    Scenario: Login deploys and runs successfully
      Given the login service is deployed
      When a user authenticates successfully
      Then the service responds with a session
      And the health check confirms the service is running