@deps:go @env:isolated-test-project @control:ai-2 @control:si-10
Feature: eac-commands_create-spec

  As a developer of the eac platform
  I want AI-powered specification generation
  So that I can create Gherkin feature files from natural language descriptions

  Background:
    Given I am in a git repository

  Rule: Command validates input

    @L2 @ov
    Scenario: Command requires description argument
      When I run "create spec" without arguments
      Then the command exits with code 1
      And stderr contains "description is required"

  Rule: Existing files are protected by default

    @skip:broken @L2 @ov
    Scenario: Command refuses to overwrite existing files
      Given a specification file exists at "specs/eac-commands/auth/specification.feature"
      And the mock AI generates a feature that would create the same path
      When I run the command "create spec 'Add authentication'"
      Then the exit code is 1
      And stderr contains "File already exists"
      And stderr contains "--force"

  Rule: Security prevents path traversal

    @skip:broken @L2 @ov
    Scenario: Path traversal attempts are rejected
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec -o '../../../etc/passwd' 'Test security'"
      Then the exit code is 1
      And stderr contains "security error"
