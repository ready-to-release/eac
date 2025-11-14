@L2 @ov
Feature: src-commands_test-suite
  As a developer
  I want to run tests by suite
  So that I can execute appropriate tests for different pipeline stages

  Background:
    Given the EAC repository is available
    And test suites are defined (pre-commit, acceptance, production-verification)

  Rule: Must list available test suites

    @skip # Meta-test: Testing test-suite command itself requires CLI execution infrastructure not yet implemented
    Scenario: User lists available suites
      When I run "test list-suites"
      Then the command succeeds
      And the output contains "pre-commit"
      And the output contains "acceptance"
      And the output contains "production-verification"

  Rule: Must verify system dependencies before running tests

    @skip # Meta-test: Requires mocking system dependency state and verifying internal command behavior
    Scenario: All dependencies available
      Given all system dependencies are installed
      When I run "test pre-commit"
      Then the command checks system dependencies
      And the command runs the selected tests

    @skip # Meta-test: Requires simulating missing dependencies and capturing interactive prompt behavior
    Scenario: Missing dependencies
      Given Docker is not installed
      And a test requires "@deps:docker"
      When I run "test pre-commit"
      Then the command checks system dependencies
      And the output warns "Missing dependencies: @deps:docker"
      And the command asks if I want to continue
      When I answer "no"
      Then the command exits with code 1

    @skip # Meta-test: Verifies --skip-deps flag behavior requires CLI execution with flag parsing
    Scenario: Skip dependency check
      Given Docker is not installed
      When I run "test pre-commit --skip-deps"
      Then the command does not check system dependencies
      And the command runs the selected tests

  Rule: Must select and run tests matching suite criteria

    @skip # Meta-test: Verifying discovery/inference/selection phases requires parsing multi-phase command output
    Scenario: Run pre-commit suite
      When I run "test pre-commit"
      Then the command discovers all tests
      And the command applies inference rules
      And the command selects tests with tags "@L0", "@L1", or "@L2"
      And the command runs Go unit tests
      And the command runs Godog features

    @skip # Meta-test: Testing verification tag selection logic requires analyzing test selection phase output
    Scenario: Run acceptance suite
      When I run "test acceptance"
      Then the command selects tests with tags "@iv", "@ov", or "@pv"
      And the command runs the selected tests

    @skip # Meta-test: Testing compound tag selection (AND logic) requires verifying selection phase filtering
    Scenario: Run production-verification suite
      When I run "test production-verification"
      Then the command selects tests with tags "@L4" AND "@piv"
      And the command runs the selected tests

  Rule: Must generate test reports

    @skip # Meta-test: Verifying test report generation requires running full test execution and parsing report structure
    Scenario: Generate test summary
      When I run "test pre-commit"
      And tests complete successfully
      Then a test run directory is created in "out/test-results/<timestamp>"
      And the summary includes discovered tests count
      And the summary includes selected tests count
      And the summary includes dependency verification results
      And the summary includes test execution results

  Rule: Must handle errors gracefully

    @skip # Meta-test: Testing error handling requires CLI execution and exit code verification
    Scenario: Unknown suite
      When I run "test unknown-suite"
      Then the command exits with code 1
      And the error message is "suite not found: unknown-suite"
      And the output lists available suites
