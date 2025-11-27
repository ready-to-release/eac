@deps:docker @ov
Feature: src-commands_design-validate

  Background:
    Given docker service is available

  Rule: Single module validation

    Scenario: Validate existing workspace
      Given module "test-module" has a valid workspace at "specs/test-module/.design/workspace.dsl"
      When I run "design validate test-module"
      Then the exit code should be 0
      And the output should contain "Validating module:"
      And the output should contain "Summary:"
      And validation results should be written to "out/design-validation-results.json"

    Scenario: Validate invalid workspace
      Given module "test-module" has an invalid workspace at "specs/test-module/.design/workspace.dsl"
      When I run "design validate test-module"
      Then the exit code should be 1
      And the output should contain "Errors:"

    Scenario: Module not found
      When I run "design validate nonexistent-module"
      Then the exit code should be 2
      And the output should contain "workspace not found"

  Rule: Batch validation

    Scenario: Validate all modules
      Given multiple modules have workspace files
      When I run "design validate --all"
      Then the exit code should be 0 or 1
      And the output should contain "Total modules:"
      And aggregated results should be written to JSON file

  Rule: Debug output

    Scenario: Verbose output shows Docker commands
      Given module "test-module" has a valid workspace
      When I run "design validate test-module --verbose"
      Then the output should contain Docker command details
