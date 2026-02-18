# Intent: Ensure OSCAL profile documents conform to the OSCAL 1.1.2 schema before use in security control definitions
# Architecture: Affects eac-cli validate risk-profile command; reads OSCAL profile JSON file; validates against OSCAL 1.1.2 schema; checks for valid JSON structure and required profile fields

@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-cli_validate-risk-profile

  As a security engineer
  I want to validate OSCAL profile documents
  So that I can ensure they conform to the OSCAL 1.1.2 schema before use

  Background:
    Given I am in a git repository

  Rule: Command requires valid inputs

    Scenario: Missing file shows error
      When I run "validate risk-profile"
      Then the exit code is 1
      And stderr contains "file path required"

    Scenario: Non-existent file shows error
      When I run "validate risk-profile nonexistent.json"
      Then the exit code is 1
      And stderr contains "file not found"

  Rule: Command validates OSCAL profile documents

    Scenario: Valid profile passes validation
      Given a valid OSCAL profile at "profile.json"
      When I run "validate risk-profile profile.json"
      Then the exit code is 0
      And stdout contains "✓ Profile is valid OSCAL 1.1.2 document"

    Scenario: Invalid JSON fails validation
      Given a file "invalid.json" with invalid JSON
      When I run "validate risk-profile invalid.json"
      Then the exit code is 1
      And stderr contains "Invalid OSCAL profile JSON"
