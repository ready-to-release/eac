@L2 @deps:ai @deps:docker @ov @skip:wip
Feature: src-commands_design-create

  Background:
    Given a repository with contracts at "contracts/ai/design/0.1.0"
    And docker service is available

  Rule: Module must exist in source directory

    Scenario: Create design for existing module
      Given module "test-module" exists in "src/test-module"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module"
      Then the exit code should be 0
      And the file "specs/test-module/.design/workspace.dsl" should exist
      And the workspace should pass Structurizr validation

    Scenario: Fail when module does not exist
      When I run "design create nonexistent-module"
      Then the exit code should be 1
      And the output should contain "source code not found"

  Rule: Output path can be customized

    Scenario: Use custom output path
      Given module "test-module" exists in "src/test-module"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module --output out/custom.dsl"
      Then the exit code should be 0
      And the file "out/custom.dsl" should exist

  Rule: Existing files require force flag

    Scenario: Fail when workspace exists without force
      Given module "test-module" exists in "src/test-module"
      And a workspace file exists at "specs/test-module/.design/workspace.dsl"
      When I run "design create test-module"
      Then the exit code should be 1
      And the output should contain "Use --force to overwrite"

    Scenario: Overwrite with force flag
      Given module "test-module" exists in "src/test-module"
      And a workspace file exists at "specs/test-module/.design/workspace.dsl"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module --force"
      Then the exit code should be 0

  Rule: Custom prompts can be provided

    Scenario: Use custom prompt file
      Given module "test-module" exists in "src/test-module"
      And a custom prompt file at "prompts/design.md"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module --prompt prompts/design.md"
      Then the exit code should be 0
      And debug logs should show the custom prompt was used

  Rule: Validation can be skipped

    Scenario: Skip Docker validation
      Given module "test-module" exists in "src/test-module"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module --skip-validation"
      Then the exit code should be 0
      And the output should contain "Validation skipped"

  Rule: Debug mode saves intermediate files

    Scenario: Debug mode outputs to logs
      Given module "test-module" exists in "src/test-module"
      And the mock AI is configured to return a valid workspace
      When I run "design create test-module --debug"
      Then the exit code should be 0
      And debug files should exist in "out/logs/design/"
