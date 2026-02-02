@L2 @ov @deps:go @control:sa-3 @env:isolated-test-project
Feature: eac-cli_show-test-results
  As a developer
  I want to see test execution results in human-readable format
  So that I can quickly understand test status and coverage

  Background:
    Given a repository with test manifests

  Rule: Uses template-based rendering with all key sections

    Scenario: Renders results with template
      Given test execution data is available
      When I run "show test-results"
      Then the output uses markdown template
      And includes "Test Execution Results" header
      And includes module overview table
      And includes specification coverage table
      And includes control summary section

    Scenario: Template includes all key sections
      Given multiple modules with test results
      When I run "show test-results"
      Then the output includes sections:
        | section                       |
        | Test Execution Results        |
        | Module Overview               |
        | Specification Coverage        |
        | Test Results by Module        |
        | Summary                       |

    Scenario: Shows test run information
      Given tests ran at "2026-01-07T15:42:56Z"
      And 6 modules were tested
      When I run "show test-results"
      Then the output includes "**Last Run:** 2026-01-07T15:42:56Z"
      And the output includes "**Modules Tested:** 6"

    Scenario: Module overview table
      Given module "eac-cli" has 701 tests: 701 passed, 0 failed
      And module has control tags: [ai-2, sa-3]
      When I run "show test-results"
      Then the module overview table includes row:
        | Module       | Tests | Passed | Failed | Controls  |
        | eac-cli      | 701   | 701    | 0      | ai-2, sa-3 |

  Rule: Shows specification coverage with summaries

    Scenario: Specification coverage table
      Given feature "create-design" has 3 scenarios: 3 passed, 0 failed
      And feature has control tags: [ai-2]
      When I run "show test-results"
      Then the spec coverage table includes row:
        | Feature        | Scenarios | Status | Controls |
        | create-design  | 3         | ✓ 3/3  | ai-2     |

    Scenario: Shows coverage summary
      Given 28 features with scenarios
      And 142 total scenarios: 142 passed, 0 failed
      When I run "show test-results"
      Then the spec coverage section shows summary:
        """
        Features: 28
        Total scenarios: 142
        Passed: 142 (100%)
        """

    Scenario: Test status formatting
      Given tests with different statuses:
        | name   | status  |
        | Test1  | passed  |
        | Test2  | failed  |
        | Test3  | skipped |
      When I run "show test-results"
      Then status is formatted with icons:
        | status  | icon |
        | passed  | ✓    |
        | failed  | ✗    |
        | skipped | ⊘    |

  Rule: Shows control summaries

    Scenario: Control summary section
      Given control "ai-2" has 47 tests across 2 modules
      And all 47 tests passed
      When I run "show test-results"
      Then the control summary includes:
        """
        - **ai-2**: 47 tests across 2 modules (all passed)
        """

  Rule: Shows per-module test breakdowns

    Scenario: Module breakdown includes type and suite
      Given module "eac-cli" has 747 tests
      When I run "show test-results"
      Then the module section includes "By Type"
      And includes "By Suite"

    Scenario: Shows test listing with metadata
      Given module "eac-cli" has tests with various statuses
      When I run "show test-results"
      Then the module section includes test listing table
      And shows test type, name, suite, status, and tags

    Scenario: Shows module duration
      Given module "eac-cli" has duration 33.1 seconds
      When I run "show test-results"
      Then the module overview shows "33.1s" in Duration column

  Rule: Handles missing manifests gracefully

    Scenario: No manifests exist
      Given no test manifests exist in out/test/
      When I run "show test-results"
      Then the command fails with error
      And suggests running "test <module>" first
