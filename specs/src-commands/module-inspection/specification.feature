@src-commands @modules @inspection
Feature: src-commands-module-inspection

  As a developer
  I want to inspect module contracts in the repository
  So that I can understand the module structure and organization

  Background:
    Given the repository has module contracts defined
    And I am in the src/commands directory

  Rule: All modules must be shown with details

    @success @ac1
    Scenario: Show all modules with details
      When I run "go run . show modules"
      Then I should see a table header with "Moniker | Type | Root Path"
      And each module row contains moniker, type, and root path
      And the exit code is 0

    @success @ac1
    Scenario: Show modules includes all contract modules
      Given contracts define modules for cli, docs, automation
      When I run "go run . show modules"
      Then I should see "src-cli" in the moniker column
      And I should see "docs" in the moniker column
      And I should see "automation-gauge" in the moniker column

    @success @ac1
    Scenario: Show modules displays correct module types
      When I run "go run . show modules"
      Then module types include "source", "documentation", "automation"
      And each module shows its assigned type

  Rule: Module types must be shown with counts

    @success @ac2
    Scenario: Show module types with counts
      When I run "go run . show moduletypes"
      Then I should see a table header with "Module Type | Count"
      And each row shows a module type and its count
      And the exit code is 0

    @success @ac2
    Scenario: Module types are sorted alphabetically
      When I run "go run . show moduletypes"
      Then "automation" appears before "documentation"
      And "documentation" appears before "source"
      And types are in strict alphabetical order

    @success @ac2
    Scenario: Module types table includes footer
      When I run "go run . show moduletypes"
      Then I should see footer row "Total Types | <count>"
      And total count matches unique module types

  Rule: Module data must be sourced from contracts v0.1.0

    @success @ac3
    Scenario: Module data sourced from contracts v0.1.0
      When I run "go run . show modules"
      Then module data comes from contracts/modules/0.1.0/
      And reports.GetModuleContracts is called with version "0.1.0"

  Rule: Output must be formatted as markdown table

    @success @ac4
    Scenario: Modules output formatted as markdown table
      When I run "go run . show modules"
      Then output is valid markdown table format
      And columns are separated by "|"
      And header row is followed by separator row

    @success @ac4
    Scenario: Module types output formatted as markdown table
      When I run "go run . show moduletypes"
      Then output is valid markdown table format
      And footer row is distinguished from data rows

  Rule: Contract loading errors must be handled gracefully

    @error @ac5
    Scenario: Show modules handles contract loading error
      Given module contracts cannot be loaded
      When I run "go run . show modules"
      Then I should see "Error:" on stderr
      And the exit code is 1

    @error @ac5
    Scenario: Show moduletypes handles contract loading error
      Given module contracts cannot be loaded
      When I run "go run . show moduletypes"
      Then I should see "Error:" on stderr
      And the exit code is 1

    @error @ac5
    Scenario: Invalid repository path handled gracefully
      Given I am in a directory without module contracts
      When I run "go run . show modules"
      Then I should see descriptive error message
      And the exit code is 1
