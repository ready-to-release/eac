@ov
Feature: src-commands_design-command

  As a developer of the eac platform
  I want to view architecture documentation using Structurizr Lite
  So that I can understand the system design through interactive C4 diagrams

  Rule: Command starts Structurizr container and displays documentation

    @L2 @deps:docker
    Scenario: Start Structurizr for a module
      Given docker service is available
      And module "src-cli" has workspace.dsl file
      When I run the command "design serve src-cli --no-browser"
      Then Structurizr container should start successfully
      And I should see success message with URL
      And documentation should be accessible at the URL

    Scenario: List available modules
      When I run the command "design list"
      Then the exit code is 0
      And I should see a list of available modules
      And "src-cli" module should be in the list

  Rule: Command validates workspace files using Structurizr CLI

    @L2 @deps:docker
    Scenario: Validate one module
      Given docker service is available
      And module "src-cli" has workspace.dsl file
      When I run the command "design validate src-cli"
      Then the workspace should be validated using Structurizr CLI
      And validation results should be displayed in console
      And validation results should be written to JSON file
      And I should see validation summary with errors and warnings

    @L2 @deps:docker
    Scenario: Validate all modules
      Given docker service is available
      And multiple modules have workspace.dsl files
      When I run the command "design validate --all"
      Then all workspaces should be validated using Structurizr CLI
      And validation results for each module should be displayed in console
      And aggregated validation results should be written to JSON file
      And I should see overall summary with total errors and warnings
