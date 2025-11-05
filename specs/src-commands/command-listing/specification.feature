@src-commands @introspection @command_listing
Feature: src-commands_command-listing

  As a CLI user
  I want to list and describe available commands
  So that I can discover what the CLI can do

  Background:
    Given the CLI application has registered commands

  Rule: Commands must be listed alphabetically

    @success @ac1
    Scenario: List all commands alphabetically
      When I run "go run . list commands"
      Then I should see "Available Commands"
      And commands are listed in alphabetical order
      And "commit-ai" appears before "describe commands"
      And "show modules" appears after "show files"
      And the exit code is 0

    @success @ac1
    Scenario: List commands with compact formatting
      When I run "go run . list commands"
      Then output uses compact list format
      And each command is on its own line
      And no extra whitespace padding exists

  Rule: Command descriptions must be available as JSON

    @success @ac2
    Scenario: Describe commands outputs valid JSON
      When I run "go run . describe commands"
      Then the output is valid JSON
      And JSON contains "commands" array
      And JSON contains "tree" object
      And the exit code is 0

    @success @ac2
    Scenario: Describe commands includes command metadata
      When I run "go run . describe commands"
      Then each command in JSON has "name" field
      And each command has "parts" array field
      And each command has "description" field
      And each command has "parent" field
      And each command has "is_leaf" boolean field

  Rule: Command descriptions must come from file comments

    @success @ac3
    Scenario: Command descriptions from file comments
      When I run "go run . describe commands"
      Then "list commands" has description "List all available commands"
      And "show modules" has description "Show all module contracts in the repository"
      And "commit-ai" has description matching AI commit generation

    @skip @success @ac3
    Scenario: Missing descriptions return empty string
      # SKIPPED: Needs investigation - may already work
      Given a command without description comment
      When I run "go run . describe commands"
      Then that command has empty description field
      And no error is raised

  Rule: Command hierarchy must be represented correctly

    @success @ac4
    Scenario: Root commands have no parent
      When I run "go run . describe commands"
      Then "commit-ai" command has parent field ""
      And "commit-ai" is marked as is_leaf true

    @success @ac4
    Scenario: Nested commands show parent relationship
      When I run "go run . describe commands"
      Then "show modules" command has parent "show"
      And "show files staged" command has parent "show files"
      And all nested commands are marked as is_leaf true

    @success @ac4
    Scenario: Tree structure maps parent to children
      When I run "go run . describe commands"
      Then tree object contains "show" key
      And "show" maps to array including "modules", "files", "moduletypes"
      And "show files" maps to array including "staged", "changed"

    @success @ac4
    Scenario: Command parts array splits by whitespace
      When I run "go run . describe commands"
      Then "show files staged" has parts ["show", "files", "staged"]
      And "list commands" has parts ["list", "commands"]
      And "commit-ai" has parts ["commit-ai"]
