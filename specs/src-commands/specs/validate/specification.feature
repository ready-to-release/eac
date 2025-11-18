@deps:go @ov
Feature: src-commands_specs-validate

  As a developer of the eac platform
  I want to validate existing Gherkin specifications against contracts
  So that I can ensure specifications follow the required structure and conventions

  Background:
    Given I am in a git repository
    And specification contracts are available

  Rule: Command must be registered and accessible

    @L2 @iv
    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "specs-validate"

    @L2 @iv
    Scenario: Command has proper description
      When I run the command "describe commands specs-validate"
      Then the exit code is 0
      And I should see "validate" or "specification" or "Gherkin"

  Rule: Command validates input and configuration

    @L2 @ov
    Scenario: Command requires file or directory path argument
      When I run "specs validate" without arguments
      Then the command exits with code 1
      And stderr contains "Error: file path or directory is required"
      And stderr contains "Usage: specs validate <path>"

    @L2 @ov
    Scenario: Command validates file path exists
      When I run "specs validate specs/nonexistent/file.feature"
      Then the command exits with code 1
      And stderr contains "file or directory not found"

    @L2 @ov
    Scenario: Command validates contracts are available
      Given contract files are missing from "contracts/specifications"
      When I run "specs validate specs/src-commands/specs/create/specification.feature"
      Then the command exits with code 1
      And stderr contains "failed to load contract"

  Rule: Single file validation uses contract-based validator

    @L2 @ov
    Scenario: Validate valid specification file
      Given a valid specification file at "specs/test/valid.feature"
      When I run "specs validate specs/test/valid.feature"
      Then the command exits with code 0
      And stdout contains "✅ Validation passed"
      And stdout contains "specs/test/valid.feature"

    @L2 @ov
    Scenario: Validate specification with errors
      Given a specification file with missing Rule at "specs/test/invalid.feature"
      When I run "specs validate specs/test/invalid.feature"
      Then the command exits with code 1
      And stdout contains "❌ Validation failed"
      And stdout contains "MISSING_RULE"
      And stdout contains "Missing Rule: declaration (required for ATDD acceptance criteria)"

    @L2 @ov
    Scenario: Validate specification with multiple errors
      Given a specification file with multiple errors at "specs/test/multi-errors.feature"
      When I run "specs validate specs/test/multi-errors.feature"
      Then the command exits with code 1
      And stdout contains "❌ Validation failed"
      And stdout contains error count "3 error(s)"

    @L2 @ov
    Scenario: Validate specification with warnings only
      Given a specification file with naming warnings at "specs/test/warnings.feature"
      When I run "specs validate specs/test/warnings.feature"
      Then the command exits with code 0
      And stdout contains "✅ Validation passed"
      And stdout contains "1 warning(s)"

  Rule: Directory validation processes all feature files

    @L2 @ov
    Scenario: Validate all specifications in directory
      Given multiple specification files in "specs/test/"
      When I run "specs validate specs/test/"
      Then the command exits with code 0
      And stdout contains "Processing" and file count
      And stdout contains validation summary

    @L2 @ov
    Scenario: Validate directory with mixed valid and invalid files
      Given 3 valid specifications and 2 invalid specifications in "specs/test/"
      When I run "specs validate specs/test/"
      Then the command exits with code 1
      And stdout contains "3 passed, 2 failed"
      And stdout lists all invalid files with their errors

    @L2 @ov
    Scenario: Validate directory recursively
      Given specification files in nested directories under "specs/"
      When I run "specs validate specs/"
      Then the command processes all .feature files recursively
      And stdout contains validation results for all files

    @L2 @ov
    Scenario: Skip non-feature files in directory
      Given a directory with .feature, .md, and .txt files
      When I run "specs validate specs/test/"
      Then only .feature files are validated
      And non-feature files are ignored silently

  Rule: Validation uses same logic as create command

    @L2 @ov
    Scenario: Validate against contract structure
      Given a specification file at "specs/test/spec.feature"
      When I run "specs validate specs/test/spec.feature"
      Then the GherkinValidator is used
      And contract rules from "contracts/specifications/0.1.0/structure.yml" are applied

    @L2 @ov
    Scenario: Check feature naming convention
      Given a specification with feature name "InvalidName"
      When I run "specs validate specs/test/spec.feature"
      Then validation fails with "INVALID_FEATURE_NAMING"
      And stderr contains naming convention explanation

    @L2 @ov
    Scenario: Check verification tags on scenarios
      Given a specification with scenarios missing verification tags
      When I run "specs validate specs/test/spec.feature"
      Then validation fails with "MISSING_VERIFICATION_TAG"
      And stderr contains "required: @ov, @iv, @pv, @piv, or @ppv"

    @L2 @ov
    Scenario: Check keyword ordering
      Given a specification with Rule before Feature
      When I run "specs validate specs/test/spec.feature"
      Then validation fails with "RULE_BEFORE_FEATURE"

    @L2 @ov
    Scenario: Check ATDD requirements
      Given a specification without Rule blocks
      When I run "specs validate specs/test/spec.feature"
      Then validation fails with "MISSING_RULE"
      And stderr contains "required for ATDD acceptance criteria"

    @L2 @ov
    Scenario: Check BDD requirements
      Given a specification without Scenario blocks
      When I run "specs validate specs/test/spec.feature"
      Then validation fails with "MISSING_SCENARIO"
      And stderr contains "required for BDD behavior examples"

  Rule: Output format provides actionable feedback

    @L2 @ov
    Scenario: Display validation errors with line numbers
      Given a specification with errors on lines 10, 25, and 40
      When I run "specs validate specs/test/spec.feature"
      Then stdout contains "Line 10:" with error details
      And stdout contains "Line 25:" with error details
      And stdout contains "Line 40:" with error details

    @L2 @ov
    Scenario: Display error codes for programmatic use
      Given a specification with validation errors
      When I run "specs validate specs/test/spec.feature"
      Then stdout contains error codes like "MISSING_RULE", "INVALID_FEATURE_NAMING"

    @L2 @ov
    Scenario: Display severity levels
      Given a specification with errors and warnings
      When I run "specs validate specs/test/spec.feature"
      Then errors are displayed with "[ERROR]" prefix
      And warnings are displayed with "[WARNING]" prefix

    @L2 @ov
    Scenario: Summary shows total error and warning counts
      Given a specification with 3 errors and 2 warnings
      When I run "specs validate specs/test/spec.feature"
      Then stdout contains "3 error(s), 2 warning(s)"

  Rule: Command supports optional flags

    @L2 @ov
    Scenario: Quiet mode shows only errors
      Given multiple specification files in "specs/test/"
      When I run "specs validate specs/test/ --quiet"
      Then stdout only contains files with errors
      And successful validations are not displayed

    @L2 @ov
    Scenario: Verbose mode shows detailed validation
      When I run "specs validate specs/test/spec.feature --verbose"
      Then stdout contains detailed validation steps
      And stdout shows which contract rules are checked

    @L2 @ov
    Scenario: JSON output for CI integration
      When I run "specs validate specs/test/spec.feature --format json"
      Then stdout contains valid JSON
      And JSON includes file path, errors array, and summary

  Rule: Security validation prevents path traversal

    @L2 @ov
    Scenario: Reject path traversal attempts
      When I run "specs validate ../../etc/passwd"
      Then the command exits with code 1
      And stderr contains "invalid path" or "path must be within repository"

    @L2 @ov
    Scenario: Reject absolute paths outside repository
      When I run "specs validate /etc/passwd"
      Then the command exits with code 1
      And stderr contains "path must be within repository"

    @L2 @ov
    Scenario: Accept relative paths within repository
      When I run "specs validate specs/src-commands/specs/create/specification.feature"
      Then the path is accepted
      And validation proceeds normally

  Rule: Edge cases must be handled gracefully

    @L2 @ov
    Scenario: Handle empty specification file
      Given an empty file at "specs/test/empty.feature"
      When I run "specs validate specs/test/empty.feature"
      Then the command exits with code 1
      And stdout contains "EMPTY_OUTPUT"

    @L2 @ov
    Scenario: Handle malformed specification file
      Given a file with invalid UTF-8 encoding at "specs/test/malformed.feature"
      When I run "specs validate specs/test/malformed.feature"
      Then the command exits with code 1
      And stderr contains "failed to read file"

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

    @L2 @ov
    Scenario: Handle file system permissions error
      Given a specification file with no read permissions
      When I run "specs validate specs/test/no-read.feature"
      Then the command exits with code 1
      And stderr contains "permission denied" or "failed to read file"
