@src-commands
Feature: src-commands_docs-command

  As a developer of the eac platform
  I want to serve project documentation using MkDocs
  So that I can view and navigate the documentation locally

  Rule: Command starts MkDocs container and serves documentation

    @success @ac1 @docker
    Scenario: Start MkDocs documentation server
      Given docker service is available
      When I run the command "docs serve --no-browser"
      Then MkDocs container should start successfully
      And I should see success message with URL
      And documentation should be accessible at "http://localhost:8000"

    @success @ac1 @docker
    Scenario: Stop MkDocs documentation server
      Given docker service is available
      And MkDocs container is running
      When I run the command "docs serve --stop"
      Then MkDocs container should be stopped
      And I should see "stopped" message
