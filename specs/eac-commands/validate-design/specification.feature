@L2 @deps:docker @ov @env:isolated-test-project
Feature: eac-commands_validate-design

  # Note: Tests use real fixture modules (eac-core, docs, r2r-cli) instead of
  # dynamic test modules. This matches production workflow and avoids module
  # registration complexity. The mock validator is enabled via R2R_MOCK_STRUCTURIZR
  # environment variable, which is set automatically by the test context.

  Background:
    Given docker service is available

  Rule: Single module validation

    Scenario: Validate existing workspace
      Given module "eac-core" has a valid workspace at "specs/eac-core/.design/workspace.dsl"
      When I run "validate design eac-core"
      Then the exit code should be 0
      And the output should contain "Validating module:"
      And the output should contain "Summary:"
      And validation results should be written to "out/design/validation-results.json"

    Scenario: Validate invalid workspace
      Given module "eac-core" has an invalid workspace at "specs/eac-core/.design/workspace.dsl"
      When I run "validate design eac-core"
      Then the exit code should be 1
      And the output should contain "Errors:"

    Scenario: Module not found
      When I run "validate design nonexistent-module"
      Then the exit code should be 2
      And the output should contain "Module not found"

  Rule: Batch validation

    Scenario: Validate all modules
      Given multiple modules have workspace files
      When I run "validate design --all"
      Then the exit code should be 0 or 1
      And the output should contain "Total modules:"
      And aggregated results should be written to JSON file

  Rule: Debug output

    Scenario: Verbose output shows Docker commands
      Given module "eac-core" has a valid workspace
      When I run "validate design eac-core --verbose"
      Then the output should contain Docker command details
