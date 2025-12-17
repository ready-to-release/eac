@L2 @deps:go @ov
Feature: eac-commands_get-commands

  As a developer integrating with the eac platform
  I want to get structured command information
  So that I can build shell integrations and tooling

  Rule: Outputs structured JSON by default

    Scenario: Output valid JSON
      When I run "get commands"
      Then the exit code is 0
      And stdout is valid JSON

    Scenario: JSON contains command list
      When I run "get commands"
      Then the exit code is 0
      And the JSON contains a "commands" array
      And each command has "name" and "description"

    Scenario: JSON contains command tree
      When I run "get commands"
      Then the exit code is 0
      And the JSON contains a "tree" object
      And the tree maps parent commands to children

  Rule: Command information is complete

    Scenario: Commands have required fields
      When I run "get commands"
      Then the exit code is 0
      And each command has:
        | field       |
        | name        |
        | parts       |
        | description |
        | parent      |
        | is_leaf     |

    Scenario: Leaf commands are marked correctly
      When I run "get commands"
      Then the exit code is 0
      And "show modules" has "is_leaf" set to true
      And "show" has "is_leaf" set to false

  Rule: Module completions are included

    @skip:wip
    Scenario: Modules list included for completion
      Given modules exist in the repository
      When I run "get commands"
      Then the exit code is 0
      And the JSON contains a "modules" array
      And the array contains module monikers

  Rule: Argument types are specified

    @skip:wip
    Scenario: Commands with module arguments are marked
      When I run "get commands"
      Then the exit code is 0
      And "build" has "args" set to "modules"
      And "show files" has "args" empty or unset

  Rule: Supports YAML output format

    @skip:wip
    Scenario: Output as YAML with --format flag
      When I run "get commands --format yaml"
      Then the exit code is 0
      And stdout is valid YAML
      And the YAML contains commands list
