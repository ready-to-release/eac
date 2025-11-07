@src-commands
Feature: src-commands_docs-command

  As a developer
  I want to serve project documentation using MkDocs
  So that I can view and navigate the documentation locally

  Rule: Command starts MkDocs container and serves documentation

    @success @ac1
    Scenario: Start MkDocs documentation server
      Given Docker is running
      When I run "go run . docs serve"
      Then MkDocs container should start successfully
      And I should see success message with URL
      And documentation should be accessible at "http://localhost:8000"

    @success @ac1
    Scenario: Stop MkDocs documentation server
      Given Docker is running
      And MkDocs container is running
      When I run "go run . docs serve --stop"
      Then MkDocs container should be stopped
      And I should see "stopped" message
