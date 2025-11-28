@L2 @deps:ai @deps:docker @ov @env:isolated-test-project
Feature: src-commands_design-update

  Background:
    Given a repository with contracts at "contracts/ai/design/0.1.0"
    And docker service is available

  Rule: Update requires existing workspace

    Scenario: Update existing workspace
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "design update test-module"
      Then the exit code should be 0
      And the workspace should be updated
      And the workspace should pass Structurizr validation

    Scenario: Fail when workspace does not exist
      Given module "test-module" exists in "src/test-module"
      And no workspace exists for "test-module"
      When I run "design update test-module"
      Then the exit code should be 1
      And the output should contain "workspace not found"

  Rule: Confirmation required by default

    Scenario: Prompt for confirmation
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      When I run "design update test-module" interactively
      Then the output should contain "Update workspace?"

    Scenario: Skip confirmation with force flag
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "design update test-module --force"
      Then the exit code should be 0
      And no confirmation prompt should appear

  Rule: Custom output path supported

    Scenario: Update to custom path
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "design update test-module --output out/updated.dsl"
      Then the exit code should be 0
      And the file "out/updated.dsl" should exist
