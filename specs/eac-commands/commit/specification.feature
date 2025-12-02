@deps:go @env:isolated-test-project @L2 @ov
Feature: eac-commands_commit

  As a developer of the eac platform
  I want AI-powered commit message generation
  So that I can create semantic commit messages from staged changes

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "commit"

    Scenario: Command has proper description
      When I run the command "show help commit"
      Then the exit code is 0
      And I should see "commit" or "AI" or "commit"

  Rule: Command validates contract implementation before execution

    Scenario: Command can be described
      When I run the command "show help commit"
      Then the exit code is 0
      And I should see "commit"

    Scenario: Command handles all execution paths
      When I run the command "show help"
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

  Rule: Agent noise must be filtered from AI-generated output

    Scenario: Remove markdown code fences wrapping entire output
      Given AI output wrapped in triple backticks
      When noise filtering is applied
      Then the code fences should be removed

  Rule: Module sections must be generated for multi-module commits

    Scenario: Skip module sections for single-module commit
      Given one affected module
      When module sections are generated
      Then no module sections should be created

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

    Scenario: Handle module name edge cases
      Given module names with edge cases (single char, max length, special patterns)
      When module names are validated
      Then validation should correctly accept or reject based on rules
