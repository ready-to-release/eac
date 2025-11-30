@L2 @ov @deps:docker
Feature: src-commands_serve-docs

  As a developer of the eac platform
  I want to serve project documentation using MkDocs
  So that I can view and navigate the documentation locally

  Rule: Command routing and help functionality

    Scenario: Show help with --help flag
      When I run the command "serve docs --help"
      Then the exit code is 0
      And I should see "serve docs"

  Rule: MkDocs server lifecycle management

    Scenario: Start MkDocs server with --no-browser flag
      Given docker service is available
      When I run the command "serve docs --no-browser"
      Then the exit code is 0
      And MkDocs container should start successfully
      And I should see "Starting MkDocs documentation server" or "MkDocs is already running"
      And I should see "Documentation:"
      And documentation should be accessible

    Scenario: Reject request for different port when container running
      Given docker service is available
      And MkDocs container is running
      When I run the command "serve docs --no-browser --port 9999"
      Then the exit code is 1
      And I should see "MkDocs is already running on port"
      And I should see "To use a different port"

    Scenario: Handle already running container
      Given docker service is available
      And MkDocs container is running
      When I run the command "serve docs --no-browser"
      Then the exit code is 0
      And I should see "MkDocs is already running"
      And I should see "Documentation:"

    Scenario: Stop MkDocs server
      Given docker service is available
      And MkDocs container is running
      When I run the command "serve docs --stop"
      Then the exit code is 0
      And MkDocs container should be stopped
      And I should see "MkDocs documentation server stopped"

    Scenario: Stop when container not running
      Given docker service is available
      And MkDocs container is not running
      When I run the command "serve docs --stop"
      Then the exit code is 0
      And I should see "MkDocs documentation server stopped"

  Rule: Error handling and validation

    Scenario: Handle invalid port number
      When I run the command "serve docs --port invalid"
      Then the exit code is 1
      And I should see "Error: invalid port number"

    Scenario: Handle missing port value
      When I run the command "serve docs --port"
      Then the exit code is 1
      And I should see "Error: --port requires a value"

    Scenario: Handle unknown flag
      When I run the command "serve docs --unknown-flag"
      Then the exit code is 1
      And I should see "Error: unknown flag: --unknown-flag"
