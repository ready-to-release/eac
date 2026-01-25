@control:sa-3
Feature: repository_command-docs-coverage

  As a documentation maintainer
  I want to ensure every CLI command has reference documentation
  So that users can discover and learn about all available commands

  Background:
    Given the repository root exists

  Rule: Every valid command must have a corresponding documentation file

    @L1 @ov
    Scenario: All commands have reference documentation
      Given I load all valid commands from the CLI
      And I scan docs/reference/eac/commands/ for command documentation files
      When I check each command for a corresponding documentation file
      Then every command should have a documentation file
      And if any commands are missing documentation, I should see their names
