@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-cli_validate-risk-catalog

  As a security engineer
  I want to validate OSCAL catalogs against the official schema
  So that I can ensure control definitions conform to OSCAL 1.1.3 specification

  Background:
    Given I am in a git repository

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
      And stdout contains "Invalid JSON format"

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
  Rule: Catalog validated using go-oscal types

    @L2 @ov
    Scenario: Catalog parsed with go-oscal types
      Given a valid OSCAL catalog
      When I run "validate risk-catalog catalog.json"
      Then the catalog is parsed using go-oscal types
      And go-oscal validation is used

    @L2 @ov
    Scenario: Required fields validated with go-oscal
      Given a catalog missing UUID
      When I run "validate risk-catalog catalog.json"
      Then the exit code is 1
      And stderr contains "Missing required field: uuid"

    @L2 @ov
    Scenario: Metadata validated with go-oscal
      Given a catalog with missing metadata title
      When I run "validate risk-catalog catalog.json"
      Then the exit code is 1
      And stderr contains "Missing required field: metadata.title"

    @L2 @ov
    Scenario: Catalog must have controls or groups
      Given a catalog without controls or groups
      When I run "validate risk-catalog catalog.json"
      Then the exit code is 1
      And stderr contains "must have at least one control or group"

    @L2 @ov
    Scenario: Valid catalog output format
      Given a valid OSCAL 1.1.3 catalog
      When I run "validate risk-catalog catalog.json"
      Then the exit code is 0
      And stdout contains "✓ Catalog is valid OSCAL 1.1.3 document"
