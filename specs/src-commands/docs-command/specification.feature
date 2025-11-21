@L2 @ov @deps:docker @env:isolated-test-project @env:mkdocs-docker
Feature: src-commands_docs-command

  As a developer of the eac platform
  I want to serve project documentation using MkDocs
  So that I can view and navigate the documentation locally

  Rule: Command starts MkDocs container and serves documentation

    @skip:broken # Container enters 'restarting' state - implementation issue, not timing
    Scenario: Start MkDocs documentation server
      Given docker service is available
      When I run the command "docs serve --no-browser"
      Then MkDocs container should start successfully
      And I should see success message with URL
      And documentation should be accessible at "http://localhost:8000"

    @skip:broken
    Scenario: Stop MkDocs documentation server
      Given docker service is available
      When I run the command "docs serve --no-browser"
      Then MkDocs container should start successfully
      When I run the command "docs serve --stop"
      Then MkDocs container should be stopped
      And I should see "stopped" message
