@L2 @deps:docker @ov @env:isolated-test-project
Feature: eac-commands_update-design

  Background:
    Given a repository with AI configs at ".r2r/eac/ai/design"
    And docker service is available

  Rule: Update requires existing workspace

    @skip:wip
    Scenario: Update existing workspace
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "update design test-module"
      Then the exit code should be 0
      And the workspace should be updated
      And the workspace should pass Structurizr validation

    @skip:wip
    Scenario: Fail when workspace does not exist
      Given module "test-module" exists in "src/test-module"
      And no workspace exists for "test-module"
      When I run "update design test-module"
      Then the exit code should be 1
      And the output should contain "workspace not found"

  Rule: Confirmation required by default

    @skip:wip
    Scenario: Prompt for confirmation
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      When I run "update design test-module" interactively
      Then the output should contain "Update workspace?"

    @skip:wip
    Scenario: Skip confirmation with force flag
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "update design test-module --force"
      Then the exit code should be 0
      And no confirmation prompt should appear

  Rule: Custom output path supported

    @skip:wip
    Scenario: Update to custom path
      Given module "test-module" has a workspace at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return an updated workspace
      When I run "update design test-module --output out/updated.dsl"
      Then the exit code should be 0
      And the file "out/updated.dsl" should exist
