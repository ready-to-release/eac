@deps:go @ov @skip:todo
Feature: src-commands_test-debug

  As a developer of the eac platform
  I want to parse test output files and list all failures
  So that I can quickly identify and locate bugs in the codebase

  Rule: Command must parse test log files in out/test directory

    @ov
    Scenario: Parse test logs with failures
      Given test log files exist in "out/test" with failure information
      When I run the command "test debug"
      Then the exit code is 0
      And I should see a table with bug descriptions and file locations

    @ov
    Scenario: Handle empty test output directory
      Given the "out/test" directory is empty
      When I run the command "test debug"
      Then the exit code is 0
      And I should see "No test failures found" or "No bugs detected"

    @ov
    Scenario: Handle missing test output directory
      Given the "out/test" directory does not exist
      When I run the command "test debug"
      Then the exit code is 0
      And I should see "No test output directory found" or "out/test not found"

  Rule: Command must extract failure details from Go test output

    @ov
    Scenario: Extract file and line number from failure
      Given a test log contains "--- FAIL: TestExample (0.00s)"
      And the log contains "example_test.go:42: assertion failed"
      When I run the command "test debug"
      Then the output should include "example_test.go:42"
      And the output should include "TestExample"

    @ov
    Scenario: Extract panic information
      Given a test log contains "panic: runtime error"
      And the log contains "parser.go:123"
      When I run the command "test debug"
      Then the output should include "parser.go:123"
      And the output should include "panic"

  Rule: Command must be accessible and provide usage help

    @ov
    Scenario: Command is registered
      When I run the command "test debug --help"
      Then the exit code is 0
      And I should see "test debug" or "Parse test output"

    @ov
    Scenario: Command runs without arguments
      When I run the command "test debug"
      Then the exit code is 0
