@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_create-risk-assess

  As a security engineer
  I want to create OSCAL assessment-results from existing test and security evidence
  So that I can track control satisfaction status with evidence

  Note: This command is read-only and does NOT run tests or scans.
  It only reads existing evidence and errors if evidence is missing or too old.

  Background:
    Given I am in a git repository
    And module "billing" exists with a profile at "specs/risk-controls/billing.profile.json"

  Rule: Command requires valid inputs

    Scenario: Missing profile flag shows error
      When I run "create risk-assess billing"
      Then the exit code is 1
      And stderr contains "required flag missing: --profile"

    Scenario: Non-existent profile file shows error
      When I run "create risk-assess billing --profile nonexistent.json"
      Then the exit code is 1
      And stderr contains "profile file not found"

  Rule: Command is read-only and errors when evidence is missing or too old

    Scenario: Command errors when no test evidence exists
      Given module "billing" has no test results
      When I run "create risk-assess billing --profile specs/risk-controls/billing.profile.json"
      Then the exit code is 1
      And stderr contains "no evidence found for module 'billing'"
      And stderr contains "out/test/"
      And stderr contains "Run tests and scans to generate evidence"

    Scenario: Command errors when test evidence is too old
      Given module "billing" has test results older than 24 hours
      When I run "create risk-assess billing --profile specs/risk-controls/billing.profile.json"
      Then the exit code is 1
      And stderr contains "test evidence for module 'billing' is too old"
      And stderr contains "out/test/"
      And stderr contains "Run tests to update evidence"

    Scenario: Command errors when security evidence is too old
      Given module "billing" has security scan results older than 24 hours
      When I run "create risk-assess billing --profile specs/risk-controls/billing.profile.json"
      Then the exit code is 1
      And stderr contains "security evidence for module 'billing' is too old"
      And stderr contains "out/scan/"
      And stderr contains "Run security scans to update evidence"

  Rule: Command creates OSCAL assessment-results from fresh evidence

    Scenario: Create assessment-results for single module with fresh evidence
      Given module "billing" has fresh test results with @control tags
      And module "billing" has fresh security scan results
      When I run "create risk-assess billing --profile specs/risk-controls/billing.profile.json"
      Then the exit code is 0
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And stdout contains "Assessing module: billing"

  Rule: Command supports multi-module assessment

    Scenario: Assess all modules when no module specified
      Given modules "billing", "api", and "auth" exist with profiles
      And all modules have fresh test and security evidence
      When I run "create risk-assess --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And files matching "out/risk/*/api/assessment-results.json" exist
      And files matching "out/risk/*/auth/assessment-results.json" exist
      And stdout contains "3 module(s) completed successfully"

    Scenario: Assess multiple specific modules with space-separated names
      Given modules "billing" and "api" exist with profiles
      And all modules have fresh test and security evidence
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And files matching "out/risk/*/api/assessment-results.json" exist
      And stdout contains "2 module(s) completed successfully"

  Rule: Parallel execution is default for multiple modules

    Scenario: Multiple modules run in parallel by default
      Given modules "billing", "api", and "auth" exist with profiles
      And all modules have fresh test and security evidence
      When I run "create risk-assess billing api auth --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And stdout contains "Assessing 3 modules in parallel"
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And files matching "out/risk/*/api/assessment-results.json" exist
      And files matching "out/risk/*/auth/assessment-results.json" exist

    Scenario: Sequential flag disables parallel execution
      Given modules "billing" and "api" exist with profiles
      And all modules have fresh test and security evidence
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json --sequential"
      Then the exit code is 0
      And stdout contains "Assessing 2 modules sequentially"
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And files matching "out/risk/*/api/assessment-results.json" exist

  Rule: Partial failures are handled gracefully

    Scenario: One module has fresh evidence, another has no evidence
      Given module "billing" exists with a profile
      And module "api" exists with a profile
      And module "billing" has fresh test results with @control tags
      And module "api" has no evidence
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And stderr contains "no evidence found for module 'api'"
      And stdout contains "1 module(s) failed"
      And stdout contains "1 module(s) completed successfully"

    Scenario: One module has fresh evidence, another has stale evidence
      Given module "billing" exists with a profile
      And module "api" exists with a profile
      And module "billing" has fresh test results with @control tags
      And module "api" has test results older than 24 hours
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/*/billing/assessment-results.json" exist
      And stderr contains "test evidence for module 'api' is too old"
      And stdout contains "1 module(s) failed"
      And stdout contains "1 module(s) completed successfully"
