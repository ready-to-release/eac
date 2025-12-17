@L2 @deps:go @ov
Feature: eac-commands_show-help

  As a developer using the eac platform
  I want to see help information for commands
  So that I can learn how to use the available functionality

  Rule: Lists all commands without arguments

    Scenario: Show all available commands
      When I run "show help"
      Then the exit code is 0
      And I see a list of commands with descriptions
      And commands are grouped by category

    Scenario: Commands are sorted alphabetically within groups
      When I run "show help"
      Then the exit code is 0
      And commands within each group are sorted

  Rule: Shows detailed help for specific command

    Scenario: Show help for a single command
      When I run "show help build"
      Then the exit code is 0
      And I see the command description
      And I see available flags
      And I see usage examples

    Scenario: Show help for multi-word command
      When I run "show help show modules"
      Then the exit code is 0
      And I see the description for "show modules"

    Scenario: Handle unknown command
      When I run "show help nonexistent-command"
      Then the exit code is 1
      And stderr contains "unknown command" or "not found"

  Rule: Verbose mode shows additional details

    @skip:wip
    Scenario: Verbose flag shows subcommands
      When I run "show help --verbose"
      Then the exit code is 0
      And I see all subcommands listed
      And I see advanced options

    @skip:wip
    Scenario: Shorthand -v for verbose
      When I run "show help -v"
      Then the exit code is 0
      And I see all subcommands listed

  Rule: Help for parent commands shows subcommands

    Scenario: Show help for parent command lists children
      When I run "show help show"
      Then the exit code is 0
      And I see available subcommands under "show"
      And each subcommand has a brief description

    Scenario: Show help for create lists AI commands
      When I run "show help create"
      Then the exit code is 0
      And I see "create spec"
      And I see "create design"
      And I see "create pr"
