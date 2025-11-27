@L2 @ov @deps:docker @env:isolated-test-project
Feature: src-commands_docs-command

  As a developer of the eac platform
  I want to serve project documentation using MkDocs
  So that I can view and navigate the documentation locally

  Rule: Command routing and help functionality

    Scenario: Show docs command usage when no subcommand provided
      When I run the command "docs"
      Then the exit code is 1
      And I should see "Usage: go run . docs <subcommand>"
      And I should see "serve"

    Scenario: Show help with --help flag
      When I run the command "docs --help"
      Then the exit code is 0
      And I should see "Usage: go run . docs <subcommand>"
      And I should see "serve"

    Scenario: Show error for unknown subcommand
      When I run the command "docs unknown"
      Then the exit code is 1
      And I should see "Error: unknown subcommand: unknown"

  Rule: MkDocs server lifecycle management

    @skip:broken-after-merge
    Scenario: Start MkDocs server with --no-browser flag
      Given docker service is available
      When I run the command "docs serve --no-browser"
      Then the exit code is 0
      And MkDocs container should start successfully
      And I should see "Starting MkDocs documentation server"
      And I should see "MkDocs documentation server is running"
      And I should see "Documentation:"
      And documentation should be accessible

    @skip:broken-after-merge
    Scenario: Start MkDocs server with custom port
      Given docker service is available
      When I run the command "docs serve --no-browser --port 9001"
      Then the exit code is 0
      And MkDocs container should start successfully
      And I should see "Documentation: http://localhost:9001"

    Scenario: Handle already running container
      Given docker service is available
      And MkDocs container is running
      When I run the command "docs serve --no-browser"
      Then the exit code is 0
      And I should see "MkDocs is already running"
      And I should see "Documentation:"

    Scenario: Stop MkDocs server
      Given docker service is available
      And MkDocs container is running
      When I run the command "docs serve --stop"
      Then the exit code is 0
      And MkDocs container should be stopped
      And I should see "MkDocs documentation server stopped"

    Scenario: Stop when container not running
      Given docker service is available
      And MkDocs container is not running
      When I run the command "docs serve --stop"
      Then the exit code is 0
      And I should see "MkDocs documentation server stopped"

  Rule: Error handling and validation

    Scenario: Handle invalid port number
      When I run the command "docs serve --port invalid"
      Then the exit code is 1
      And I should see "Error: invalid port number"

    Scenario: Handle missing port value
      When I run the command "docs serve --port"
      Then the exit code is 1
      And I should see "Error: --port requires a value"

    Scenario: Handle unknown flag
      When I run the command "docs serve --unknown-flag"
      Then the exit code is 1
      And I should see "Error: unknown flag: --unknown-flag"
