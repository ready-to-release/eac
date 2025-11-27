@skip:todo @deps:go @env:isolated-test-project @L2 @ov
Feature: src-commands_specs-validate

  As a developer of the eac platform
  I want to validate existing Gherkin specifications against contracts
  So that I can ensure specifications follow the required structure and conventions

  Background:
    Given I am in a git repository
    And specification contracts are available

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "specs validate"

    @L3 @iv
    Scenario: Command has proper description
      When I run the command "describe commands specs validate"
      Then the exit code is 0
      And I should see "validate" or "specification" or "Gherkin"

  Rule: Command validates input and configuration

    @L2 @ov
    Scenario: Command requires file or directory path argument
      When I run "specs validate" without arguments
      Then the command exits with code 1
      And stderr contains "file path or directory is required"

    @L2 @ov
    Scenario: Command validates file path exists
      When I run "specs validate specs/nonexistent/file.feature"
      Then the command exits with code 1
      And stderr contains "file or directory not found"

    @L2 @ov
    Scenario: Invalid format option is rejected
      When I run "specs validate specs/test/valid.feature --format xml"
      Then the command exits with code 1
      And stderr contains "invalid format"

  Rule: Single file validation uses contract-based validator

    @L2 @ov
    Scenario: Validate valid specification file
      Given a valid specification file at "specs/test/valid.feature"
      When I run "specs validate specs/test/valid.feature"
      Then the command exits with code 0
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: Validate specification with errors
      Given a specification file with missing Rule at "specs/test/invalid.feature"
      When I run "specs validate specs/test/invalid.feature"
      Then the command exits with code 1
      And stdout contains "Validation failed"

    @L2 @ov
    Scenario: Validate specification with multiple errors
      Given a specification file with multiple errors at "specs/test/multi-errors.feature"
      When I run "specs validate specs/test/multi-errors.feature"
      Then the command exits with code 1
      And stdout contains "Validation failed"

  Rule: Directory validation processes all feature files

    @L2 @ov
    Scenario: Validate all specifications in directory
      Given multiple specification files in "specs/test/"
      When I run "specs validate specs/test/"
      Then the command exits with code 0

    @L2 @ov
    Scenario: Validate directory with mixed valid and invalid files
      Given 3 valid specifications and 2 invalid specifications in "specs/test/"
      When I run "specs validate specs/test/"
      Then the command exits with code 1

    @L2 @ov
    Scenario: Validate nested directories recursively
      Given specification files in nested directories under "specs/test/"
      When I run "specs validate specs/test/"
      Then the command processes all .feature files recursively

  Rule: Output format can be configured

    @L2 @ov
    Scenario: Default text format shows human-readable output
      Given a valid specification file at "specs/test/valid.feature"
      When I run "specs validate specs/test/valid.feature"
      Then the command exits with code 0
      And stdout contains "Validation passed"

    @L2 @ov
    Scenario: JSON format provides machine-readable output
      Given a valid specification file at "specs/test/valid.feature"
      When I run "specs validate specs/test/valid.feature --format json"
      Then the command exits with code 0
      And stdout contains valid JSON

    @L2 @ov
    Scenario: JSON output includes summary
      Given multiple specification files in "specs/test/"
      When I run "specs validate specs/test/ --format json"
      Then stdout contains valid JSON

  Rule: Quiet mode suppresses success messages

    @L2 @ov
    Scenario: Quiet mode shows only errors
      Given 3 valid specifications and 2 invalid specifications in "specs/test/"
      When I run "specs validate specs/test/ --quiet"
      Then the command exits with code 1
      And stdout only contains files with errors

    @L2 @ov
    Scenario: Quiet mode with all valid files
      Given multiple specification files in "specs/test/"
      When I run "specs validate specs/test/ -q"
      Then the command exits with code 0

  Rule: Verbose mode provides detailed output

    @L2 @ov
    Scenario: Verbose mode shows detailed validation
      Given a valid specification file at "specs/test/valid.feature"
      When I run "specs validate specs/test/valid.feature --verbose"
      Then the command exits with code 0

  Rule: Validation checks structure and tags

    @L2 @ov
    Scenario: Check verification tags on scenarios
      Given a specification with scenarios missing verification tags
      When I run "specs validate specs/test/spec.feature"
      Then the command exits with code 1
      And stdout contains "MISSING_VERIFICATION_TAG"

    @L2 @ov
    Scenario: Check Rule requirements
      Given a specification without Rule blocks
      When I run "specs validate specs/test/spec.feature"
      Then the command exits with code 1

  Rule: Edge cases must be handled gracefully

    @L2 @ov
    Scenario: Handle empty specification file
      Given an empty file at "specs/test/empty.feature"
      When I run "specs validate specs/test/empty.feature"
      Then the command exits with code 1

    @L2 @ov
    Scenario: Handle empty directory
      Given an empty directory at "specs/empty/"
      When I run "specs validate specs/empty/"
      Then the command exits with code 0
      And stdout contains "No specification files found"

    @L2 @ov
    Scenario: Handle directory with no feature files
      Given a directory with only .md files at "specs/docs/"
      When I run "specs validate specs/docs/"
      Then the command exits with code 0
      And stdout contains "No specification files found"

  Rule: Security prevents path traversal

    @L2 @ov
    Scenario: Path traversal attempts are rejected
      When I run "specs validate ../../../etc/passwd"
      Then the command exits with code 1
      And stderr contains "path must be within repository"
