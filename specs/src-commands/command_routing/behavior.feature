# Feature ID: src_commands_command_routing
# Module: src-commands

@src-commands @critical @command_routing
Feature: Command Discovery and Routing

  As a CLI user
  I want commands to be auto-discovered and routed correctly
  So that I can execute any registered command without manual registration

  Background:
    Given the CLI application is initialized
    And multiple commands are registered

  Rule: Commands register themselves via init() and are discoverable

    @success @ac1
    Scenario: Command registers itself via init function
      Given a command file with init() calling Register()
      When the application starts
      Then the command is available in the commands map
      And the command can be invoked by name

  Rule: Nested commands match longest path first

    @success @ac2
    Scenario: Nested command matches longest path first
      Given commands "show" and "show modules" are registered
      When I run "go run . show modules"
      Then the "show modules" command executes
      And not the "show" command with "modules" as argument

    @success @ac2
    Scenario: Three-level nested command routing
      Given commands "show files" and "show files staged" are registered
      When I run "go run . show files staged"
      Then the "show files staged" command executes
      And not the "show files" command

  Rule: Parent commands without implementation show available subcommands

    @success @ac3
    Scenario: Parent command shows subcommands
      Given commands "show modules", "show files", "show moduletypes" are registered
      When I run "go run . show"
      Then I should see "Available subcommands for 'show':"
      And I should see "modules" in the list
      And I should see "files" in the list
      And I should see "moduletypes" in the list
      And the exit code is 0

    @success @ac3
    Scenario: Root level shows all commands
      Given multiple root-level commands are registered
      When I run "go run ." without arguments
      Then I should see "Available commands:"
      And I should see all registered commands listed

  Rule: Unknown commands show helpful error messages

    @error @ac4
    Scenario: Unknown command shows error
      Given command "invalid-cmd" does not exist
      When I run "go run . invalid-cmd"
      Then I should see "Error: Command not found: invalid-cmd"
      And I should see "Available commands:"
      And the exit code is 1

    @error @ac4
    Scenario: Unknown nested command shows error
      Given command "show invalid" does not exist
      When I run "go run . show invalid"
      Then I should see "Error: Command not found: show invalid"
      And the exit code is 1

  Rule: Commands return appropriate exit codes

    @success @ac5
    Scenario: Successful command returns exit code 0
      Given command "list commands" succeeds
      When I run "go run . list commands"
      Then the exit code is 0

    @error @ac5
    Scenario: Failed command returns exit code 1
      Given command "show modules" encounters an error
      When I run "go run . show modules" in invalid directory
      Then the exit code is 1
