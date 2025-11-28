@L2 @deps:docker @ov @env:isolated-test-project
Feature: src-commands_design-serve

  Background:
    Given docker service is available

  Rule: Module workspace required

    Scenario: Serve existing workspace
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      When I run "design serve test-module"
      Then Structurizr container should start successfully
      And I should see success message with URL
      And documentation should be accessible at the URL

    Scenario: Fail when workspace missing
      Given no workspace exists for "test-module"
      When I run "design serve test-module"
      Then the exit code should be 1
      And the output should contain "Workspace not found"
      And the output should contain "Create one first"

  Rule: Multi-instance support

    Scenario: Each module gets unique container
      Given module "module-a" has a workspace
      And module "module-b" has a workspace
      When I run "design serve module-a"
      And I run "design serve module-b"
      Then two separate containers should be running
      And each should have a unique port
