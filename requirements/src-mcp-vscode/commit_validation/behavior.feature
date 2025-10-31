# Feature ID: src_mcp_vscode_commit_validation
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @validation @critical
Feature: Commit Validation

  Background:
    Given the validation system is initialized
    And I have a GitContext with staged changes

  @success @ac1
  Scenario: Pass validation with top-level heading
    Given I have a commit message starting with "# feat: Add new feature"
    When I validate the commit message
    Then no MD041 error is returned
    And the heading validation passes

  @error @ac1
  Scenario: Fail validation without top-level heading
    Given I have a commit message starting with "This is a commit"
    When I validate the commit message
    Then an MD041 error is returned
    And the error indicates "First line in a file should be a top-level heading"

  @error @ac1
  Scenario: Fail validation with wrong heading level
    Given I have a commit message starting with "## feat: Add feature"
    When I validate the commit message
    Then an MD041 error is returned
    And the error indicates the heading must use single #

  @success @ac2
  Scenario: Pass validation with correct semantic format
    Given I have a module section "## cli: feat: Add new command"
    When I validate the commit message
    Then no semantic format error is returned
    And the format validation passes

  @error @ac2
  Scenario: Fail validation with missing type
    Given I have a module section "## cli: Add new command"
    When I validate the commit message
    Then a semantic format error is returned
    And the error indicates missing type field

  @error @ac2
  Scenario: Fail validation with invalid type
    Given I have a module section "## cli: newfeature: Add command"
    When I validate the commit message
    Then a semantic format error is returned
    And the error indicates "newfeature" is not a valid type

  @success @ac2
  Scenario: Pass validation with all valid types
    Given I have module sections with types: feat, fix, docs, refactor, test, chore
    When I validate the commit message
    Then no semantic format errors are returned
    And all types are recognized as valid

  @error @ac2
  Scenario: Fail validation with missing colons
    Given I have a module section "## cli feat Add command"
    When I validate the commit message
    Then a semantic format error is returned
    And the error indicates missing colon separators

  @success @ac3
  Scenario: Pass validation with subject within 72 chars
    Given I have a top-level heading "# feat: Add user authentication system"
    When I validate the commit message
    Then no subject length error is returned
    And the subject validation passes

  @error @ac3
  Scenario: Fail validation with subject exceeding 72 chars
    Given I have a top-level heading with 85 characters
    When I validate the commit message
    Then a subject length error is returned
    And the error indicates the subject is too long
    And the error includes the character count

  @success @ac3
  Scenario: Subject length excludes the # character
    Given I have a heading "# " followed by exactly 72 characters
    When I validate the commit message
    Then no subject length error is returned

  @success @ac4
  Scenario: Pass validation with all body lines within 72 chars
    Given I have a commit body with lines of 50, 60, and 70 characters
    When I validate the commit message
    Then no line wrapping errors are returned
    And the body validation passes

  @error @ac4
  Scenario: Fail validation with body line exceeding 72 chars
    Given I have a commit body with a line of 85 characters
    When I validate the commit message
    Then a line wrapping error is returned
    And the error includes the line number
    And the error indicates the line is too long

  @success @ac4
  Scenario: Exclude code blocks from wrapping check
    Given I have a code block with lines exceeding 72 characters
    When I validate the commit message
    Then no line wrapping errors are returned for code block lines

  @success @ac4
  Scenario: Exclude table rows from wrapping check
    Given I have a markdown table with rows exceeding 72 characters
    When I validate the commit message
    Then no line wrapping errors are returned for table rows

  @error @ac4
  Scenario: Multiple line wrapping errors
    Given I have 3 body lines exceeding 72 characters
    When I validate the commit message
    Then 3 line wrapping errors are returned
    And each error includes the specific line number

  @success @ac5
  Scenario: Pass validation with complete file list
    Given GitContext has 5 changed files
    And the commit message lists all 5 files in tables
    When I validate the commit message
    Then no file completeness errors are returned
    And the file validation passes

  @error @ac5
  Scenario: Fail validation with missing files
    Given GitContext has 5 changed files
    And the commit message lists only 3 files
    When I validate the commit message
    Then a file completeness error is returned
    And the error lists the 2 missing files

  @error @ac5
  Scenario: Fail validation with extra files
    Given GitContext has 3 changed files
    And the commit message lists 5 files
    When I validate the commit message
    Then a file completeness error is returned
    And the error lists the 2 extra files

  @error @ac5
  Scenario: Fail validation with both missing and extra files
    Given GitContext has files [A, B, C]
    And the commit message lists files [B, D, E]
    When I validate the commit message
    Then file completeness errors are returned
    And the errors indicate A and C are missing
    And the errors indicate D and E are extra

  @success @ac6
  Scenario: Pass validation with all modules present
    Given GitContext has files from modules [cli, docs, src-mcp-vscode]
    And the commit has sections for [cli, docs, src-mcp-vscode]
    When I validate the commit message
    Then no module completeness errors are returned
    And the module validation passes

  @error @ac6
  Scenario: Fail validation with missing module section
    Given GitContext has files from modules [cli, docs, src-mcp-vscode]
    And the commit has sections for [cli, docs]
    When I validate the commit message
    Then a module completeness error is returned
    And the error indicates src-mcp-vscode is missing

  @error @ac6
  Scenario: Fail validation with multiple missing modules
    Given GitContext has files from 4 modules
    And the commit has sections for 2 modules
    When I validate the commit message
    Then module completeness errors are returned
    And the errors list all missing modules

  @success @ac7
  Scenario: Pass validation with closed YAML blocks
    Given I have a commit with properly closed ```yaml blocks
    When I validate the commit message
    Then no YAML block errors are returned
    And the YAML validation passes

  @error @ac7
  Scenario: Fail validation with unclosed YAML block
    Given I have a commit with ```yaml but no closing ```
    When I validate the commit message
    Then a YAML block error is returned
    And the error indicates an unclosed code block

  @success @ac7
  Scenario: Multiple YAML blocks all closed
    Given I have a commit with 3 YAML blocks all properly closed
    When I validate the commit message
    Then no YAML block errors are returned

  @error @ac7
  Scenario: Multiple YAML blocks with one unclosed
    Given I have a commit with 3 YAML blocks
    And the second block is not closed
    When I validate the commit message
    Then a YAML block error is returned
    And the error indicates which block is unclosed

  @success @ac8
  Scenario: Pass validation following module contract
    Given I have a contract for "src-mcp-vscode" module
    And the commit adheres to the contract requirements
    When I validate the commit message
    Then no contract violation errors are returned
    And the contract validation passes

  @error @ac8
  Scenario: Fail validation violating module contract
    Given I have a contract for "cli" module
    And the commit violates a contract rule
    When I validate the commit message
    Then a contract violation error is returned
    And the error describes the specific violation

  @success @ac8
  Scenario: Validate multiple modules against their contracts
    Given I have contracts for 3 modules
    And the commit follows all 3 contracts
    When I validate the commit message
    Then no contract violation errors are returned

  @error @ac8
  Scenario: Contract violation in one of multiple modules
    Given I have contracts for 3 modules
    And the commit violates the contract for module 2
    When I validate the commit message
    Then a contract violation error is returned for module 2
    And no errors for modules 1 and 3

  @success @ac9
  Scenario: Collect multiple validation errors
    Given the commit has MD041 error, semantic format error, and file completeness error
    When I validate the commit message
    Then 3 CommitValidationError objects are returned
    And each error has code, message, and line number
    And the errors are returned as an array

  @success @ac9
  Scenario: Error objects include all required fields
    Given the commit has a validation error
    When I validate the commit message
    Then the error object includes error code
    And the error object includes descriptive message
    And the error object includes line number if applicable
    And the error object is properly structured

  @success @ac9
  Scenario: Errors are ordered by occurrence
    Given the commit has errors on lines 1, 5, and 10
    When I validate the commit message
    Then the errors array preserves the line order
    And error 1 corresponds to line 1
    And error 2 corresponds to line 5
    And error 3 corresponds to line 10

  @success @ac10
  Scenario: Return empty array for fully valid commit
    Given I have a properly formatted commit message
    And all files are present
    And all modules are represented
    And semantic format is correct
    And line lengths are within limits
    When I validate the commit message
    Then an empty errors array is returned
    And the validation result indicates success

  @success @ac10
  Scenario: Valid single-module commit
    Given I have a commit with one module
    And the module section is properly formatted
    And all files are listed correctly
    When I validate the commit message
    Then an empty errors array is returned

  @success @ac10
  Scenario: Valid multi-module commit
    Given I have a commit with 4 modules
    And all modules are properly formatted
    And all files are distributed correctly
    When I validate the commit message
    Then an empty errors array is returned

  @success @ac1 @ac2 @ac3 @ac4 @ac5 @ac6 @ac7
  Scenario: Comprehensive validation of fully valid commit
    Given I have a commit message with:
      | Property              | Value                                       |
      | Top-level heading     | # feat: Add comprehensive testing suite    |
      | Heading length        | 50 characters                               |
      | Module sections       | cli, docs, src-mcp-vscode                   |
      | Semantic format       | All sections follow <module>: <type>: desc  |
      | Body line lengths     | All ≤72 characters                          |
      | File completeness     | All git files present in tables             |
      | Module completeness   | All git modules have sections               |
      | YAML blocks           | All properly closed                         |
      | Contract compliance   | All modules follow contracts                |
    When I validate the commit message
    Then no validation errors are returned
    And the commit is approved for use

  @error @ac1 @ac2 @ac3 @ac4 @ac5 @ac6
  Scenario: Comprehensive validation with multiple errors
    Given I have a commit message with:
      | Error Type            | Description                        |
      | MD041                 | No top-level heading               |
      | Semantic format       | Invalid type in cli section        |
      | Subject length        | 90 characters (exceeds 72)         |
      | Line wrapping         | 2 lines exceed 72 characters       |
      | File completeness     | 2 files missing from tables        |
      | Module completeness   | 1 module section missing           |
    When I validate the commit message
    Then 8+ validation errors are returned
    And each error type is properly identified
    And the errors array includes all violations
