@src-commands @files @git @modules
Feature: src-commands_file-tracking

  As a developer
  I want to track files with their module ownership
  So that I can understand which modules are affected by file changes

  Background:
    Given the repository has module contracts defined
    And I am in the src/commands directory

  Rule: All tracked files must be shown with module ownership

    @success @ac1
    Scenario: Show all tracked files with modules
      When I run "go run . show files"
      Then I should see a table header with "File | Modules"
      And all tracked files are listed
      And each file shows module ownership or "NONE"
      And the exit code is 0

    @success @ac1
    Scenario: Show files reports use correct parameters
      When I run "go run . show files"
      Then GetFilesModulesReport is called with trackedOnly=true
      And includeIgnored=false
      And stagedOnly=false

  Rule: Changed files must be filtered and displayed

    @success @ac2
    Scenario: Show only changed (modified, unstaged) files
      Given I have modified files that are not staged
      When I run "go run . show files changed"
      Then I should see only modified files in output
      And staged files are excluded
      And unmodified files are excluded
      And the exit code is 0

    @success @ac2
    Scenario: Show files changed uses git diff
      When I run "go run . show files changed"
      Then command executes "git diff --name-only HEAD"
      And filters full file list to match changed files
      And module ownership is shown for each changed file

    @success @ac2
    Scenario: No changed files shows no output
      Given no files are modified
      When I run "go run . show files changed"
      Then I should see a table header with "File | Modules"
      And the exit code is 0

  Rule: Staged files must be filtered and displayed

    @success @ac3
    Scenario: Show only staged files
      Given I have staged files
      When I run "go run . show files staged"
      Then I should see only staged files in output
      And unstaged files are excluded
      And the exit code is 0

    @success @ac3
    Scenario: Show files staged reports use stagedOnly flag
      When I run "go run . show files staged"
      Then GetFilesModulesReport is called with stagedOnly=true
      And trackedOnly=true
      And includeIgnored=false

  Rule: Files with multiple modules must show comma-separated list

    @success @ac5
    Scenario: File with multiple modules shows comma-separated list
      Given file "containers/README.md" belongs to ["containers", "readme"]
      When I run "go run . show files"
      Then "containers/README.md" shows "containers, readme" in Modules column

    @success @ac5
    Scenario: Module list is comma-space separated
      Given file belongs to modules ["containers", "readme"]
      When I run "go run . show files"
      Then modules column shows "containers, readme"
      And format uses ", " separator (comma followed by space)

  Rule: Output must be formatted as markdown table

    @success @ac6
    Scenario: Show files outputs markdown table
      When I run "go run . show files"
      Then output is valid markdown table format
      And columns are separated by "|"
      And header row is present

    @success @ac6
    Scenario: Show files changed outputs markdown table
      Given there are changed files
      When I run "go run . show files changed"
      Then output is valid markdown table format

    @success @ac6
    Scenario: Show files staged outputs markdown table
      Given there are staged files
      When I run "go run . show files staged"
      Then output is valid markdown table format
