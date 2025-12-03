@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_validate-risk-catalog

  As a security engineer
  I want to validate OSCAL catalogs against the official schema
  So that I can ensure control definitions conform to OSCAL 1.1.3 specification

  Background:
    Given I am in a git repository

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "validate risk-catalog"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help validate risk-catalog"
      Then the exit code is 0
      And stdout contains "OSCAL"
      And stdout contains "catalog"
      And stdout contains "schema"

  Rule: Command requires file argument

    @L2 @ov
    Scenario: Missing file shows error
      When I run "validate risk-catalog"
      Then the exit code is 1
      And stderr contains "file path required"

    @L2 @ov
    Scenario: Non-existent file shows error
      When I run "validate risk-catalog nonexistent.json"
      Then the exit code is 1
      And stderr contains "file not found"

  Rule: Schema validation against official OSCAL schema

    @L2 @ov
    Scenario: Valid minimal catalog passes validation
      Given a file "test-catalog.json" with content:
        """
        {
          "catalog": {
            "uuid": "12345678-1234-4234-8234-123456789abc",
            "metadata": {
              "title": "Test Catalog",
              "last-modified": "2025-01-01T00:00:00Z",
              "version": "1.0.0",
              "oscal-version": "1.1.3"
            },
            "groups": [
              {
                "id": "ac",
                "title": "Access Control",
                "controls": [
                  {
                    "id": "ac-1",
                    "title": "Policy and Procedures"
                  }
                ]
              }
            ]
          }
        }
        """
      When I run "validate risk-catalog test-catalog.json"
      Then the exit code is 0
      And stdout contains "Validation passed"
      And stdout contains "Type: catalog"
      And stdout contains "Schema: OSCAL 1.1.3"

    @L2 @ov
    Scenario: Invalid JSON shows clear error
      Given a file "invalid.json" with content:
        """
        { invalid json
        """
      When I run "validate risk-catalog invalid.json"
      Then the exit code is 1
      And stdout contains "Validation failed"
      And stdout contains "invalid JSON"

    @L2 @ov
    Scenario: Missing required catalog fields fails validation
      Given a file "incomplete-catalog.json" with content:
        """
        {
          "catalog": {
            "metadata": {
              "title": "Test Catalog",
              "last-modified": "2025-01-01T00:00:00Z",
              "version": "1.0.0"
            }
          }
        }
        """
      When I run "validate risk-catalog incomplete-catalog.json"
      Then the exit code is 1
      And stdout contains "Validation failed"

    @L2 @ov
    Scenario: Catalog with invalid structure fails schema validation
      Given a file "bad-catalog.json" with content:
        """
        {
          "catalog": {
            "uuid": "not-a-valid-uuid",
            "metadata": {
              "title": "Test Catalog"
            }
          }
        }
        """
      When I run "validate risk-catalog bad-catalog.json"
      Then the exit code is 1
      And stdout contains "Validation failed"
