@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_create-risk-profile

  As a security engineer
  I want to create OSCAL profiles from risk assessment documents
  So that I can define security controls for my solution

  Background:
    Given I am in a git repository
    And AI provider is configured

  Rule: Command must be registered and accessible

    Scenario: Command shows help
      When I run the command "show help create risk-profile"
      Then the exit code is 0
      And stdout contains "OSCAL profile"
      And stdout contains "risk assessment"

  Rule: Command requires valid inputs

    Scenario: Missing assessment file shows error
      When I run "create risk-profile"
      Then the exit code is 1
      And stderr contains "assessment file path required"

    Scenario: Non-existent file shows error
      When I run "create risk-profile nonexistent.md"
      Then the exit code is 1
      And stderr contains "assessment file not found"

  Rule: Command generates OSCAL profile from assessment

    Scenario: Create profile from assessment document
      Given a risk assessment file at "assessment.md"
      When I run "create risk-profile assessment.md"
      Then the exit code is 0
      And a file exists at "specs/.risk-controls/risk-profile.json"
      And stdout contains "Created OSCAL profile"

  Rule: Force flag controls overwrite behavior

    Scenario: Existing profile not overwritten by default
      Given a risk assessment file at "assessment.md"
      And a profile exists at "specs/.risk-controls/risk-profile.json"
      When I run "create risk-profile assessment.md"
      Then the exit code is 1
      And stderr contains "already exists"

    Scenario: Force flag overwrites existing profile
      Given a risk assessment file at "assessment.md"
      And a profile exists at "specs/.risk-controls/risk-profile.json"
      When I run "create risk-profile assessment.md --force"
      Then the exit code is 0
      And a file exists at "specs/.risk-controls/risk-profile.json"
