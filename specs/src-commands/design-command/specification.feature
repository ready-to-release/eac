@src-commands
Feature: src-commands_design-command

  As a developer
  I want to view architecture documentation using Structurizr Lite
  So that I can understand the system design through interactive C4 diagrams

  Rule: Command starts Structurizr container and displays documentation

    @success @ac1
    Scenario: Start Structurizr for a module
      Given Docker is running
      And module "src-cli" has workspace.dsl file
      When I run "go run . design serve src-cli"
      Then Structurizr container should start successfully
      And I should see success message with URL
      And documentation should be accessible at the URL

    @success @ac1
    Scenario: List available modules
      When I run "go run . design list"
      Then I should see a list of available modules
      And "src-cli" module should be in the list

  Rule: Command validates workspace files using Structurizr CLI

    @success @ac2
    Scenario: Validate one module
      Given Docker is running
      And module "src-cli" has workspace.dsl file
      When I run "go run . design validate src-cli"
      Then the workspace should be validated using Structurizr CLI
      And validation results should be displayed in console
      And validation results should be written to JSON file
      And I should see validation summary with errors and warnings

    @success @ac2
    Scenario: Validate all modules
      Given Docker is running
      And multiple modules have workspace.dsl files
      When I run "go run . design validate --all"
      Then all workspaces should be validated using Structurizr CLI
      And validation results for each module should be displayed in console
      And aggregated validation results should be written to JSON file
      And I should see overall summary with total errors and warnings
