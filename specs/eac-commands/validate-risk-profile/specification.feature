@deps:go @L2 @ov @env:isolated-test-project @skip:broken
Feature: eac-commands_validate-risk-profile

  As a security engineer
  I want to validate OSCAL profile documents
  So that I can ensure they conform to the OSCAL 1.1.2 schema before use

  Background:
    Given I am in a git repository

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "validate risk-profile"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help validate risk-profile"
      Then the exit code is 0
      And stdout contains "OSCAL profile"
      And stdout contains "1.1.2"

  Rule: Command requires file argument

    @L2 @ov
    Scenario: Missing file shows error
      When I run "validate risk-profile"
      Then the exit code is 1
      And stderr contains "file path required"

    @L2 @ov
    Scenario: Non-existent file shows error
      When I run "validate risk-profile nonexistent.json"
      Then the exit code is 1
      And stderr contains "file not found"

  Rule: Command validates OSCAL profile documents using go-oscal

    @L2 @ov
    Scenario: Valid profile passes validation
      Given a valid OSCAL 1.1.2 profile at "specs/risk-controls/billing.profile.json"
      When I run "validate risk-profile specs/risk-controls/billing.profile.json"
      Then the exit code is 0
      And stdout contains "✓ Profile is valid OSCAL 1.1.2 document"

    @L2 @ov
    Scenario: Profile parsed with go-oscal types
      Given a valid OSCAL profile at "specs/risk-controls/billing.profile.json"
      When I run "validate risk-profile specs/risk-controls/billing.profile.json"
      Then the profile is parsed using go-oscal types
      And go-oscal validation is used

  Rule: Required fields are validated

    @L2 @ov
    Scenario: Missing UUID fails validation
      Given an OSCAL profile without UUID
      When I run "validate risk-profile profile.json"
      Then the exit code is 1
      And stderr contains "missing required field: uuid"

    @L2 @ov
    Scenario: Missing metadata title fails validation
      Given an OSCAL profile without metadata title
      When I run "validate risk-profile profile.json"
      Then the exit code is 1
      And stderr contains "missing required field: title"

    @L2 @ov
    Scenario: Missing last-modified fails validation
      Given an OSCAL profile without last-modified
      When I run "validate risk-profile profile.json"
      Then the exit code is 1
      And stderr contains "missing required field: last-modified"

    @L2 @ov
    Scenario: Missing imports fails validation
      Given an OSCAL profile without imports
      When I run "validate risk-profile profile.json"
      Then the exit code is 1
      And stderr contains "missing required field: imports"

    @L2 @ov
    Scenario: Import without href fails validation
      Given an OSCAL profile with import missing href
      When I run "validate risk-profile profile.json"
      Then the exit code is 1
      And stderr contains "missing required field: href"

  Rule: Control ID format is validated

    @L2 @ov
    Scenario: Valid control ID formats pass
      Given an OSCAL profile with controls "ac-2", "ia-5", "si-10"
      When I run "validate risk-profile profile.json"
      Then the exit code is 0
      And control IDs are validated

    @L2 @ov
    Scenario: Invalid control ID format shows warning
      Given an OSCAL profile with malformed control ID "invalid-id"
      When I run "validate risk-profile profile.json"
      Then the exit code is 0
      And stdout contains "Warnings:"
      And stdout contains "may not be valid NIST 800-53 format"

  Rule: OSCAL version compatibility

    @L2 @ov
    Scenario: OSCAL 1.1.2 version matches expected
      Given an OSCAL profile with oscal-version "1.1.2"
      When I run "validate risk-profile profile.json"
      Then the exit code is 0
      And no version warnings

    @L2 @ov
    Scenario: Different OSCAL version shows warning
      Given an OSCAL profile with oscal-version "1.0.0"
      When I run "validate risk-profile profile.json"
      Then the exit code is 0
      And stdout contains "Warnings:"
      And stdout contains "differs from expected 1.1.2"

  Rule: Invalid JSON handling

    @L2 @ov
    Scenario: Malformed JSON fails validation
      Given a file with invalid JSON
      When I run "validate risk-profile malformed.json"
      Then the exit code is 1
      And stderr contains "invalid OSCAL profile JSON"

    @L2 @ov
    Scenario: Non-profile OSCAL document fails
      Given an OSCAL catalog document
      When I run "validate risk-profile catalog.json"
      Then the exit code is 1
      And stderr contains "invalid OSCAL profile JSON"
