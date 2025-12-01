@L2 @deps:docker @ov @env:isolated-test-project
Feature: src-commands_validate-design

  Background:
    Given docker service is available

  Rule: Single module validation

    @skip:wip
    Scenario: Validate existing workspace
      Given module "test-module" has a valid workspace at "specs/test-module/.design/workspace.dsl"
      When I run "validate design test-module"
      Then the exit code should be 0
      And the output should contain "Validating module:"
      And the output should contain "Summary:"
      And validation results should be written to "out/design-validation-results.json"

    @skip:wip
    Scenario: Validate invalid workspace
      Given module "test-module" has an invalid workspace at "specs/test-module/.design/workspace.dsl"
      When I run "validate design test-module"
      Then the exit code should be 1
      And the output should contain "Errors:"

    @skip:wip
    Scenario: Module not found
      When I run "validate design nonexistent-module"
      Then the exit code should be 2
      And the output should contain "workspace not found"

  Rule: Batch validation

    @skip:wip
    Scenario: Validate all modules
      Given multiple modules have workspace files
      When I run "validate design --all"
      Then the exit code should be 0 or 1
      And the output should contain "Total modules:"
      And aggregated results should be written to JSON file

  Rule: Debug output

    @skip:wip
    Scenario: Verbose output shows Docker commands
      Given module "test-module" has a valid workspace
      When I run "validate design test-module --verbose"
      Then the output should contain Docker command details
