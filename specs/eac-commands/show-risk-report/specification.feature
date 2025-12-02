@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_show-risk-report

  As a security engineer
  I want to view aggregated risk assessment reports
  So that I can understand the overall risk posture across modules

  Background:
    Given I am in a git repository

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "show risk-report"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help show risk-report"
      Then the exit code is 0
      And stdout contains "aggregated"
      And stdout contains "--format"
      And stdout contains "--module"
      And stdout contains "--detail"

  Rule: Command displays aggregated risk report

    @L2 @ov
    Scenario: Show report for all modules
      Given assessment-results exist for modules "billing" and "auth"
      When I run "show risk-report"
      Then the exit code is 0
      And stdout contains "RISK ASSESSMENT REPORT"
      And stdout contains "Modules Assessed: 2"
      And stdout contains "billing"
      And stdout contains "auth"

    @L2 @ov
    Scenario: Show summary statistics
      Given assessment-results for "billing" with 5 satisfied and 2 not-satisfied controls
      When I run "show risk-report"
      Then stdout contains "Total Controls:"
      And stdout contains "Satisfied:"
      And stdout contains "Not Satisfied:"

    @L2 @ov
    Scenario: Show overall risk band
      Given assessment-results for multiple modules
      When I run "show risk-report"
      Then stdout contains "Overall Risk Band:"
      And the risk band is calculated from all modules

  Rule: Module filter supported

    @L2 @ov
    Scenario: Filter to specific module
      Given assessment-results exist for modules "billing" and "auth"
      When I run "show risk-report --module billing"
      Then the exit code is 0
      And stdout contains "billing"
      And stdout does not contain "auth"

    @L2 @ov
    Scenario: Short module flag works
      When I run "show risk-report -m billing"
      Then the exit code is 0
      And stdout contains "billing"

    @L2 @ov
    Scenario: Non-existent module shows error
      When I run "show risk-report --module nonexistent"
      Then the exit code is 1
      And stderr contains "No assessment results found for module"

  Rule: Detail flag shows findings breakdown

    @L2 @ov
    Scenario: Default shows summary only
      Given assessment-results for "billing" with findings
      When I run "show risk-report"
      Then stdout shows module summary
      And stdout does not show individual findings

    @L2 @ov
    Scenario: Detail flag shows findings
      Given assessment-results for "billing" with findings for "ac-2" and "ia-2"
      When I run "show risk-report --detail"
      Then stdout contains "ac-2"
      And stdout contains "ia-2"
      And stdout shows test coverage for each finding

  Rule: Multiple output formats supported

    @L2 @ov
    Scenario: Default text format
      Given assessment-results exist
      When I run "show risk-report"
      Then stdout contains human-readable text
      And stdout contains box characters or separators

    @L2 @ov
    Scenario: JSON format
      Given assessment-results exist for "billing"
      When I run "show risk-report --format json"
      Then the exit code is 0
      And stdout contains valid JSON
      And the JSON has "total_modules" field
      And the JSON has "modules" array

    @L2 @ov
    Scenario: Markdown format
      Given assessment-results exist for "billing"
      When I run "show risk-report --format markdown"
      Then the exit code is 0
      And stdout contains "# Risk Assessment Report"
      And stdout contains markdown table syntax

    @L2 @ov
    Scenario: Invalid format shows error
      When I run "show risk-report --format xml"
      Then the exit code is 1
      And stderr contains "invalid format"

  Rule: Report shows per-module breakdown

    @L2 @ov
    Scenario: Module breakdown in text format
      Given assessment-results exist for "billing" and "auth"
      When I run "show risk-report"
      Then stdout contains "MODULE BREAKDOWN"
      And stdout shows satisfied/total for each module
      And stdout shows risk score for each module

    @L2 @ov
    Scenario: Module breakdown in JSON format
      Given assessment-results exist for "billing" and "auth"
      When I run "show risk-report --format json"
      Then the JSON modules array has 2 entries
      And each module has "satisfied" and "total" fields
      And each module has "risk_score" field

  Rule: Risk scoring follows ISO 27005

    @L2 @ov
    Scenario: Calculate risk score per module
      Given assessment-results for "billing" with 3 not-satisfied controls
      When I run "show risk-report --format json"
      Then the module risk_score has "likelihood" field
      And the module risk_score has "impact" field
      And the module risk_score has "score" field
      And the module risk_score has "band" field

    @L2 @ov
    Scenario: Risk bands are Low, Medium, High, Critical
      Given assessment-results with varying risk levels
      When I run "show risk-report --format json"
      Then risk bands are one of "Low", "Medium", "High", "Critical"

  Rule: Report includes recommendations

    @L2 @ov
    Scenario: Show recommendations for not-satisfied controls
      Given assessment-results with not-satisfied control "ac-2"
      When I run "show risk-report"
      Then stdout contains "RECOMMENDATIONS"
      And stdout contains "ac-2"
      And stdout suggests adding tests with @control tags

    @L2 @ov
    Scenario: No recommendations when all satisfied
      Given assessment-results with all controls satisfied
      When I run "show risk-report"
      Then stdout contains "All controls are satisfied"

  Rule: Handle missing assessment-results gracefully

    @L2 @ov
    Scenario: No assessment-results found
      Given no assessment-results exist
      When I run "show risk-report"
      Then the exit code is 1
      And stderr contains "No assessment results found"
      And stderr suggests running "risk assess"

    @L2 @ov
    Scenario: Partial assessment-results with load errors
      Given valid assessment-results for "billing"
      And corrupted assessment-results for "auth"
      When I run "show risk-report"
      Then the exit code is 0
      And stdout shows report for "billing"
      And a warning is logged about "auth"

  Rule: Detail view shows test coverage

    @L2 @ov
    Scenario: Show test coverage per control
      Given assessment-results for "billing" with "ac-2" having "3/4 scenarios passing"
      When I run "show risk-report --detail"
      Then stdout shows "ac-2" with test coverage "3/4"

    @L2 @ov
    Scenario: Show no tests indicator
      Given assessment-results for "billing" with "ia-2" having no tests
      When I run "show risk-report --detail"
      Then stdout shows "ia-2" with "no tests"
