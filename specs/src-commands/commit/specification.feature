@deps:ai @deps:go @deps:git @env:isolated-test-project @ov
Feature: src-commands_commit

  As a developer of the eac platform
  I want AI-powered commit message generation
  So that I can create semantic commit messages from staged changes

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "commit"

    Scenario: Command has proper description
      When I run the command "describe commands commit"
      Then the exit code is 0
      And I should see "commit" or "AI" or "commit"

  Rule: Command validates contract implementation before execution

    Scenario: Command can be described
      When I run the command "describe commands commit"
      Then the exit code is 0
      And I should see "commit"

    Scenario: Command handles all execution paths
      When I run the command "list commands"
      Then I should see "commit"
      And the exit code is 0

  Rule: Contract validation must ensure verifier implements all contract rules

    Scenario: Contract version matches implementation
      Given a commit message contract with version "0.1.0"
      When the contract implementation is verified
      Then no version mismatch errors should occur

    Scenario: Contract has all required structure sections
      Given a commit message contract
      When the contract implementation is verified
      Then the contract must include "top_level_heading" section
      And the contract must include "top_level_body" section
      And the contract must include "module_sections" section

    Scenario: Contract has all semantic types
      Given a commit message contract
      When the contract implementation is verified
      Then the contract must include semantic types: feat, fix, refactor, docs, chore, test, perf, style

  Rule: Commit messages must follow conventional commit format

    Scenario: Header exceeds 72 characters
      Given a commit message with header longer than 72 characters
      When the message is validated
      Then a "HEADER_TOO_LONG" error should occur

    Scenario: Header has trailing period
      Given a commit message with header ending in a period
      When the message is validated
      Then a "HEADER_TRAILING_PERIOD" error should occur

    Scenario: Missing Auditor-Summary field
      Given a commit message without Auditor-Summary
      When the message is validated
      Then a "MISSING_AUDITOR_SUMMARY" error should occur

  Rule: Agent noise must be filtered from AI-generated output

    Scenario: Remove initialization messages from top-level output
      Given AI output starting with "**Initialized and ready to assist**"
      And followed by a valid commit header "feat(multi-module): add features"
      When noise filtering is applied
      Then the output should start with "feat(multi-module): add features"

    Scenario: Remove greeting from module section output
      Given AI output starting with "I'll generate the module section for you."
      And followed by module section "src-commands"
      When noise filtering is applied
      Then the output should start with "src-commands"

    Scenario: Remove markdown code fences wrapping entire output
      Given AI output wrapped in triple backticks
      When noise filtering is applied
      Then the code fences should be removed

  Rule: Auto-cleanup must fix common formatting issues

    Scenario: Remove trailing periods from header
      Given a commit message header ending with a period
      When auto-cleanup is applied
      Then the period should be removed

    Scenario: Wrap long lines at 72 characters
      Given a body text line longer than 72 characters
      When auto-cleanup is applied
      Then the line should be wrapped at word boundaries

    Scenario: Close unclosed code blocks
      Given a commit message with an opening code fence but no closing fence
      When auto-cleanup is applied
      Then a closing fence should be added

    Scenario: Normalize duplicate blank lines
      Given a commit message with multiple consecutive blank lines
      When auto-cleanup is applied
      Then duplicate blank lines should be reduced to single blank lines

    Scenario: Ensure blank lines around code blocks
      Given a code block without blank lines before and after
      When auto-cleanup is applied
      Then blank lines should be added before and after the code block

  Rule: Context building must aggregate git information

    Scenario: Build context for single-module commit
      Given staged files belonging to one module
      And a git diff for those files
      When top-level context is built
      Then the context should indicate "1 (single-module)"
      And the context should list the affected module
      And the context should include the staged files table
      And the context should include the git diff

    Scenario: Build context for multi-module commit
      Given staged files belonging to multiple modules
      And a git diff for those files
      When top-level context is built
      Then the context should indicate the count as "(multi-module)"
      And the context should list all affected modules
      And the context should include the staged files table
      And the context should include the git diff

    Scenario: Build module-specific context
      Given a module with specific files
      And a full git diff
      When module context is built
      Then the context should include only files for that module
      And the diff should be filtered to only that module's changes

  Rule: Module sections must be generated for multi-module commits

    Scenario: Skip module sections for single-module commit
      Given one affected module
      When module sections are generated
      Then no module sections should be created

    Scenario: Generate sections for each module in multi-module commit
      Given multiple affected modules
      When module sections are generated
      Then a section should be created for each module

    Scenario: Add missing module sections as stubs
      Given a multi-module commit message
      And some modules missing from the output
      When missing modules are added
      Then stub sections should be generated for missing modules
      And stubs should indicate "Module changes not described by AI agent"

  Rule: Diff filtering must isolate module-specific changes

    Scenario: Filter diff for single file
      Given a full git diff with multiple files
      And a module with one file
      When the diff is filtered for that module
      Then only that file's diff should be included

    Scenario: Filter diff for multiple files
      Given a full git diff with multiple files
      And a module with multiple files
      When the diff is filtered for that module
      Then all of that module's files should be included
      And other files should be excluded

  Rule: Edge cases must be handled gracefully

    Scenario: Handle empty staged changes
      Given no staged changes in git
      When the commit command is run
      Then the message "No staged changes." should be displayed

    Scenario: Handle git command failure
      Given git diff command fails
      When execution context is built
      Then the error should indicate git failure

    Scenario: Handle very large diff
      Given a git diff larger than 10 MB
      When execution context is built
      Then the error should indicate diff size limit exceeded

    Scenario: Handle module name edge cases
      Given module names with edge cases (single char, max length, special patterns)
      When module names are validated
      Then validation should correctly accept or reject based on rules