# Intent: Allow developers to see test status, timing, and coverage by retrieving structured test execution results from manifests
# Architecture: Affects eac-cli get test-results command; reads test manifests from out/test/; aggregates results across modules; includes specification coverage, control tag summaries, and timing data; outputs YAML or JSON

@L2 @ov @deps:go @control:sa-3 @env:isolated-test-project
Feature: eac-cli_get-test-results
  As a developer
  I want to get test execution results from manifests
  So that I can see test status, timing, and coverage

  Background:
    Given a repository with test manifests

  Rule: Returns test execution data from manifests

    Scenario: Get test results from single module
      Given module "core" has test manifest with 5 passed tests
      When I run "get test-results"
      Then the output contains "modules_tested: 1"
      And the output contains "total_tests: 5"
      And the output contains test entries with status "passed"

    Scenario: Get test results from multiple modules
      Given module "core" has test manifest with 5 passed tests
      And module "eac" has test manifest with 10 passed tests
      When I run "get test-results"
      Then the output contains "modules_tested: 2"
      And the output contains "total_tests: 15"

  Rule: Includes specification coverage for godog tests

    Scenario: Shows specification coverage
      Given module "eac" has godog test for feature "create-design"
      And the feature has 3 scenarios, all passed
      When I run "get test-results"
      Then the output contains spec_coverage entry for "create-design"
      And the entry shows 3 scenarios, 3 passed, 0 failed

    Scenario: Groups tests by feature file
      Given module "eac" has godog tests for features:
        | feature           | scenarios |
        | create-design     | 3         |
        | create-commit-msg | 2         |
      When I run "get test-results"
      Then the output contains 2 spec_coverage entries

  Rule: Includes control tag summaries

    Scenario: Shows control summaries
      Given godog tests have "@control:ai-2" tag
      And 10 tests with ai-2 passed in module "eac"
      And 5 tests with ai-2 passed in module "vscode-commit"
      When I run "get test-results"
      Then the output contains control_summary entry for "ai-2"
      And the entry shows test_count: 15
      And the entry shows modules: [eac, vscode-commit]

    Scenario: Extracts control tags from test tags
      Given test has tags ["L2", "control:au-2", "deps:go"]
      When I run "get test-results"
      Then the test has control_tags: ["au-2"]

  Rule: Handles missing or invalid manifests

    Scenario: No manifests exist
      Given no test manifests exist in out/test/
      When I run "get test-results"
      Then the command fails with error "no test manifests found"
      And the error message suggests "run tests first"

    Scenario: Invalid manifest JSON
      Given module "core" has corrupted manifest file
      When I run "get test-results"
      Then the command skips the corrupted manifest
      And logs warning about skipped manifest
      And processes other valid manifests

  Rule: Supports output format flags

    Scenario Outline: Output format flags
      Given test manifests exist
      When I run "get test-results <flag>"
      Then the output is valid <format>

      Examples:
        | flag      | format |
        |           | YAML   |
        | --as-yaml | YAML   |
        | --as-json | JSON   |

  Rule: Includes timing and suite information

    Scenario: Shows complete test metadata
      Given test "Generate message from commits" in manifest has:
        | field        | value         |
        | status       | passed        |
        | suite        | integration   |
        | duration_ms  | 1234          |
      When I run "get test-results"
      Then the test entry includes all metadata fields
