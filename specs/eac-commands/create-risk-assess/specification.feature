@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_create-risk-assess

  As a security engineer
  I want to create OSCAL assessment-results from test and security evidence
  So that I can track control satisfaction status with evidence

  Background:
    Given I am in a git repository
    And module "billing" exists with a profile at "specs/risk-controls/billing.profile.json"

  Rule: Command must be registered and accessible

    Scenario: Command shows help
      When I run the command "show help create risk-assess"
      Then the exit code is 0
      And stdout contains "assessment-results"
      And stdout contains "profile"

  Rule: Command requires valid inputs

    @skip:wip
    Scenario: Missing profile flag shows error
      When I run "create risk-assess billing"
      Then the exit code is 1
      And stderr contains "--profile flag is required"

    @skip:wip
    Scenario: Non-existent profile file shows error
      When I run "create risk-assess billing --profile nonexistent.json"
      Then the exit code is 1
      And stderr contains "profile file not found"

  Rule: Command creates OSCAL assessment-results for single module

    @skip:wip
    Scenario: Create assessment-results for single module with evidence
      Given module "billing" has test results with @control tags
      And module "billing" has security scan results
      When I run "create risk-assess billing --profile specs/risk-controls/billing.profile.json"
      Then the exit code is 0
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And stdout contains "Assessing module: billing"

  Rule: Command supports multi-module assessment

    @skip:wip
    Scenario: Assess all modules when no module specified
      Given modules "billing", "api", and "auth" exist with profiles
      And module "billing" has test results with @control tags
      And module "api" has test results with @control tags
      And module "auth" has test results with @control tags
      When I run "create risk-assess --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And files matching "out/risk/api/assessment-results-*.json" exist
      And files matching "out/risk/auth/assessment-results-*.json" exist
      And stdout contains "Modules assessed: 3"

    @skip:wip
    Scenario: Assess multiple specific modules with space-separated names
      Given modules "billing" and "api" exist with profiles
      And module "billing" has test results with @control tags
      And module "api" has test results with @control tags
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And files matching "out/risk/api/assessment-results-*.json" exist
      And stdout contains "Modules assessed: 2"

  Rule: Parallel execution is default for multiple modules

    @skip:wip
    Scenario: Multiple modules run in parallel by default
      Given modules "billing", "api", and "auth" exist with profiles
      And module "billing" has test results with @control tags
      And module "api" has test results with @control tags
      And module "auth" has test results with @control tags
      When I run "create risk-assess billing api auth --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And stdout contains "Assessing 3 modules in parallel"
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And files matching "out/risk/api/assessment-results-*.json" exist
      And files matching "out/risk/auth/assessment-results-*.json" exist

    @skip:wip
    Scenario: Sequential flag disables parallel execution
      Given modules "billing" and "api" exist with profiles
      And module "billing" has test results with @control tags
      And module "api" has test results with @control tags
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json --sequential"
      Then the exit code is 0
      And stdout contains "Assessing 2 modules sequentially"
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And files matching "out/risk/api/assessment-results-*.json" exist

  Rule: Partial failures are handled gracefully

    @skip:wip
    Scenario: One module succeeds, another has no evidence
      Given module "billing" exists with a profile
      And module "api" exists with a profile
      And module "billing" has test results with @control tags
      And module "api" has no evidence
      When I run "create risk-assess billing api --profile specs/.risk-controls/risk-profile.json"
      Then the exit code is 0
      And files matching "out/risk/billing/assessment-results-*.json" exist
      And stdout contains "1 module(s) failed"
      And stdout contains "1 module(s) completed successfully"
      And stdout contains "Modules assessed: 1"
