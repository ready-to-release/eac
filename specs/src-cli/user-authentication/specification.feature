@skip:todo
@env:isolated-test-project
@deps:go @depm:src-cli
Feature: src-cli_user-authentication

  As a user
  I want to authenticate with credentials
  So that I can securely access the application

  Rule: User can authenticate with valid credentials

    @ov
    Scenario: Successful login with valid credentials
      Given a user with username and password registered
      When the user provides valid credentials
      Then authentication succeeds
      And the user session is created
      And access to protected resources is granted

    @ov
    Scenario: User receives feedback on failed authentication
      Given a user attempting to login
      When the user provides invalid credentials
      Then authentication fails
      And an error message is displayed
      And no session is created

  Rule: User passwords are securely handled

    @ov
    Scenario: Password is not displayed during input
      Given a login form is presented
      When the user enters a password
      Then the password characters are masked
      And the password is not visible on screen

    @ov
    Scenario: Session expires after inactivity
      Given a user with an active session
      When no activity occurs for the configured timeout period
      Then the session is terminated
      And the user must re-authenticate to continue

  Rule: Authentication credentials are validated

    @ov
    Scenario: Empty credentials are rejected
      Given a login form is presented
      When the user submits without entering credentials
      Then validation fails
      And an appropriate error message is shown

    @ov
    Scenario: Malformed credentials are rejected
      Given a login form is presented
      When the user submits credentials in incorrect format
      Then validation fails
      And the user is prompted to correct the input