# Feature ID: src_mcp_vscode_auto_correction
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @correction @formatting
Feature: Auto Correction

  Background:
    Given the auto-correction system is initialized

  @success @ac1
  Scenario: Add blank line before second-level heading
    Given I have a commit message:
      """
      # Top heading
      ## Module section
      """
    When I apply auto-correction
    Then the corrected message has a blank line before "## Module section"
    And the corrected message is:
      """
      # Top heading

      ## Module section
      """

  @success @ac1
  Scenario: Add blank lines before multiple headings
    Given I have a commit message with 3 headings without blank lines
    When I apply auto-correction
    Then blank lines are added before the 2nd and 3rd headings
    And the first heading has no preceding blank line

  @success @ac1
  Scenario: Preserve existing blank lines before headings
    Given I have a commit message with proper blank lines before headings
    When I apply auto-correction
    Then no additional blank lines are added
    And the existing blank lines are preserved

  @success @ac1
  Scenario: Skip blank line before first heading
    Given I have a commit message starting with "# First heading"
    When I apply auto-correction
    Then no blank line is added before the first heading
    And the message starts with "#"

  @success @ac2
  Scenario: Add missing trailing newline
    Given I have a commit message without a trailing newline
    When I apply auto-correction
    Then a single newline is added at the end
    And the file ends with exactly one newline character

  @success @ac2
  Scenario: Reduce multiple trailing newlines to one
    Given I have a commit message with 3 trailing newlines
    When I apply auto-correction
    Then the trailing newlines are reduced to 1
    And the file ends with exactly one newline character

  @success @ac2
  Scenario: Preserve single trailing newline
    Given I have a commit message with exactly one trailing newline
    When I apply auto-correction
    Then the trailing newline is preserved
    And no modification is made to the end

  @success @ac3
  Scenario: Remove trailing spaces from lines
    Given I have lines with trailing spaces:
      """
      This line has trailing spaces
      This line also has spaces
      This line is clean
      """
    When I apply auto-correction
    Then all trailing spaces are removed
    And the corrected message is:
      """
      This line has trailing spaces
      This line also has spaces
      This line is clean
      """

  @success @ac3
  Scenario: Remove trailing tabs from lines
    Given I have lines with trailing tabs
    When I apply auto-correction
    Then all trailing tabs are removed
    And line content is preserved

  @success @ac3
  Scenario: Preserve leading indentation
    Given I have indented lines with trailing spaces
    When I apply auto-correction
    Then trailing spaces are removed
    And leading indentation is preserved

  @success @ac4
  Scenario: Reduce multiple consecutive blank lines to one
    Given I have a commit message with 3 consecutive blank lines
    When I apply auto-correction
    Then the 3 blank lines are reduced to 1 blank line
    And the surrounding content is preserved

  @success @ac4
  Scenario: Preserve single blank lines
    Given I have a commit message with single blank lines for separation
    When I apply auto-correction
    Then the single blank lines are preserved
    And no blank lines are removed

  @success @ac4
  Scenario: Fix multiple groups of consecutive blank lines
    Given I have 3 groups of multiple blank lines in the message
    When I apply auto-correction
    Then each group is reduced to a single blank line
    And the structure is preserved

  @success @ac5
  Scenario: Preserve trailing whitespace in code blocks
    Given I have a code block with trailing spaces:
      """
      ```yaml
      key: value
      another: line
      ```
      """
    When I apply auto-correction
    Then the trailing spaces inside the code block are preserved
    And the code block content is unchanged

  @success @ac5
  Scenario: Preserve multiple blank lines in code blocks
    Given I have a code block with multiple blank lines
    When I apply auto-correction
    Then the blank lines inside the code block are preserved
    And no blank lines are reduced inside the block

  @success @ac5
  Scenario: Apply corrections outside code blocks
    Given I have trailing spaces before and after a code block
    When I apply auto-correction
    Then trailing spaces outside the block are removed
    And trailing spaces inside the block are preserved

  @success @ac5
  Scenario: Handle multiple code blocks
    Given I have 3 code blocks in the message
    When I apply auto-correction
    Then all 3 code blocks preserve their internal formatting
    And corrections are applied between blocks

  @success @ac5
  Scenario: Preserve exact formatting in YAML code blocks
    Given I have a YAML code block with specific indentation
    When I apply auto-correction
    Then the YAML indentation is preserved exactly
    And no whitespace inside the block is modified

  @success @ac6
  Scenario: Preserve table structure
    Given I have a markdown table:
      """
      | File | Status |
      |------|--------|
      | main.go | Modified |
      """
    When I apply auto-correction
    Then the table structure is preserved
    And table spacing is maintained

  @success @ac6
  Scenario: Allow long lines in tables
    Given I have a table row exceeding 72 characters
    When I apply auto-correction
    Then the table row is not wrapped
    And the table remains intact

  @success @ac6
  Scenario: Apply corrections outside tables
    Given I have trailing spaces before and after a table
    When I apply auto-correction
    Then trailing spaces outside the table are removed
    And the table formatting is preserved

  @success @ac7
  Scenario: Return corrected commit message
    Given I have a commit message with multiple formatting issues
    When I apply auto-correction
    Then a corrected message string is returned
    And the message is valid markdown
    And all formatting issues are fixed

  @success @ac7
  Scenario: Corrected message is immediately valid
    Given I have a commit with formatting errors
    When I apply auto-correction
    And I validate the corrected message
    Then no formatting validation errors are returned

  @success @ac7
  Scenario: Idempotent correction
    Given I have a commit message
    When I apply auto-correction once
    And I apply auto-correction again to the result
    Then the second correction produces the same output
    And no further changes are made

  @success @ac8
  Scenario: Preserve text content
    Given I have a commit message with text "Add user authentication"
    When I apply auto-correction
    Then the text "Add user authentication" is unchanged
    And only formatting is modified

  @success @ac8
  Scenario: Preserve heading text
    Given I have a heading "## cli: feat: Add new command"
    When I apply auto-correction
    Then the heading text remains "## cli: feat: Add new command"
    And only surrounding blank lines may be added

  @success @ac8
  Scenario: Preserve file lists
    Given I have a file table with 5 files
    When I apply auto-correction
    Then all 5 files remain in the table
    And file names are unchanged
    And file paths are unchanged

  @success @ac8
  Scenario: Preserve module names
    Given I have module sections for "cli", "docs", and "src-mcp-vscode"
    When I apply auto-correction
    Then all 3 module names are preserved
    And module structure is unchanged

  @success @ac8
  Scenario: Preserve semantic meaning
    Given I have a commit describing feature additions and bug fixes
    When I apply auto-correction
    Then the semantic meaning is completely preserved
    And the commit still describes the same changes

  @success @ac1 @ac2 @ac3 @ac4
  Scenario: Comprehensive correction of multiple issues
    Given I have a commit message with:
      | Issue                          | Description                      |
      | Missing blank before heading   | 3 headings missing blank lines   |
      | Multiple trailing newlines     | 4 newlines at end                |
      | Trailing spaces                | 5 lines with trailing spaces     |
      | Multiple consecutive blanks    | 2 groups of 3+ blank lines       |
    When I apply auto-correction
    Then all blank lines before headings are added
    And trailing newlines are reduced to 1
    And all trailing spaces are removed
    And all consecutive blank line groups are reduced
    And the message is properly formatted

  @success @ac5 @ac6 @ac8
  Scenario: Correct while preserving protected content
    Given I have a commit with:
      | Content Type    | Status                           |
      | Text sections   | Has trailing spaces              |
      | Code block      | Has trailing spaces and blanks   |
      | Table           | Has long lines                   |
      | Headings        | Missing blank lines              |
    When I apply auto-correction
    Then text section trailing spaces are removed
    And code block content is preserved exactly
    And table structure is preserved
    And blank lines are added before headings
    And semantic content is unchanged
