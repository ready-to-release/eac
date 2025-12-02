@deps:go @L2 @ov @env:isolated-test-project @skip:broken
Feature: eac-commands_validate-risk

  As a security engineer
  I want to validate OSCAL risk documents
  So that I can ensure they conform to the schema before use

  Background:
    Given I am in a git repository

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "validate risk"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help validate risk"
      Then the exit code is 0
      And stdout contains "OSCAL"
      And stdout contains "profile"
      And stdout contains "assessment"
      And stdout contains "--type"
      And stdout contains "--check-evidence"

  Rule: Command requires file argument

    @L2 @ov
    Scenario: Missing file shows error
      When I run "validate risk"
      Then the exit code is 1
      And stderr contains "file path required"

    @L2 @ov
    Scenario: Non-existent file shows error
      When I run "validate risk nonexistent.json"
      Then the exit code is 1
      And stderr contains "file not found"

  Rule: Command auto-detects document type

    @L2 @ov
    Scenario: Auto-detect profile document
      Given a valid OSCAL profile at "specs/risk-controls/billing.profile.json"
      When I run "validate risk specs/risk-controls/billing.profile.json"
      Then the exit code is 0
      And stdout contains "Auto-detected document type: profile"
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: Auto-detect assessment-results document
      Given a valid OSCAL assessment-results at "out/risk/billing/assessment-results.json"
      When I run "validate risk out/risk/billing/assessment-results.json"
      Then the exit code is 0
      And stdout contains "Auto-detected document type: assessment"
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: Unknown document type
      Given a JSON file that is not OSCAL
      When I run "validate risk unknown.json"
      Then the exit code is 1
      And stderr contains "Could not auto-detect document type"

  Rule: Explicit type flag supported

    @L2 @ov
    Scenario: Explicit profile type
      Given a valid OSCAL profile at "document.json"
      When I run "validate risk document.json --type profile"
      Then the exit code is 0
      And validation uses profile schema

    @L2 @ov
    Scenario: Explicit assessment type
      Given a valid OSCAL assessment-results at "document.json"
      When I run "validate risk document.json --type assessment"
      Then the exit code is 0
      And validation uses assessment-results schema

  Rule: Profile validation checks required fields

    @L2 @ov
    Scenario: Valid profile passes validation
      Given a valid OSCAL profile with all required fields
      When I run "validate risk profile.json --type profile"
      Then the exit code is 0
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: Missing UUID fails validation
      Given a profile missing the uuid field
      When I run "validate risk profile.json --type profile"
      Then the exit code is 1
      And stdout contains "Validation failed"
      And stdout contains "missing required field: uuid"

    @L2 @ov
    Scenario: Missing title fails validation
      Given a profile missing the metadata.title field
      When I run "validate risk profile.json --type profile"
      Then the exit code is 1
      And stdout contains "missing required field: title"

    @L2 @ov
    Scenario: Missing imports fails validation
      Given a profile with no imports
      When I run "validate risk profile.json --type profile"
      Then the exit code is 1
      And stdout contains "missing required field: imports"

    @L2 @ov
    Scenario: Invalid control ID format shows warning
      Given a profile with control ID "invalid-format"
      When I run "validate risk profile.json --type profile"
      Then stdout contains warning about control ID format

  Rule: Assessment-results validation checks required fields

    @L2 @ov
    Scenario: Valid assessment-results passes validation
      Given a valid OSCAL assessment-results with all required fields
      When I run "validate risk assessment.json --type assessment"
      Then the exit code is 0
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: Missing UUID fails validation
      Given assessment-results missing the uuid field
      When I run "validate risk assessment.json --type assessment"
      Then the exit code is 1
      And stdout contains "missing required field: uuid"

    @L2 @ov
    Scenario: Missing results fails validation
      Given assessment-results with no results array
      When I run "validate risk assessment.json --type assessment"
      Then the exit code is 1
      And stdout contains "missing required field: results"

    @L2 @ov
    Scenario: Duplicate UUIDs fail validation
      Given assessment-results with duplicate observation UUIDs
      When I run "validate risk assessment.json --type assessment"
      Then the exit code is 1
      And stdout contains "duplicate UUID"

    @L2 @ov
    Scenario: Invalid observation reference fails validation
      Given assessment-results with finding referencing non-existent observation
      When I run "validate risk assessment.json --type assessment"
      Then the exit code is 1
      And stdout contains "references non-existent observation"

  Rule: Evidence link verification optional

    @L2 @ov
    Scenario: Skip evidence check by default
      Given assessment-results with evidence links to missing files
      When I run "validate risk assessment.json"
      Then the exit code is 0
      And no warnings about missing evidence

    @L2 @ov
    Scenario: Check evidence links with flag
      Given assessment-results with evidence link "out/test/results.json"
      And file "out/test/results.json" does not exist
      When I run "validate risk assessment.json --check-evidence"
      Then stdout contains warning "evidence file not found"

    @L2 @ov
    Scenario: Evidence check passes when files exist
      Given assessment-results with evidence link "out/test/results.json"
      And file "out/test/results.json" exists
      When I run "validate risk assessment.json --check-evidence"
      Then the exit code is 0
      And no warnings about missing evidence

  Rule: Validation output is informative

    @L2 @ov
    Scenario: Show error count on failure
      Given a profile with 3 validation errors
      When I run "validate risk profile.json"
      Then the exit code is 1
      And stdout contains "Errors: 3"

    @L2 @ov
    Scenario: Show warning count
      Given a profile with 2 warnings
      When I run "validate risk profile.json"
      Then stdout contains "Warnings: 2"

    @L2 @ov
    Scenario: Show field path for each error
      Given a profile missing nested field "metadata.title"
      When I run "validate risk profile.json"
      Then stdout contains "profile.metadata.title"

  Rule: OSCAL version compatibility

    @L2 @ov
    Scenario: Warn on version mismatch
      Given a profile with oscal-version "1.0.0"
      When I run "validate risk profile.json"
      Then stdout contains warning about OSCAL version
      And stdout contains "differs from expected 1.1.2"

  Rule: JSON parsing errors are clear

    @L2 @ov
    Scenario: Invalid JSON shows clear error
      Given a file with invalid JSON syntax
      When I run "validate risk invalid.json"
      Then the exit code is 1
      And stdout contains "invalid JSON"
