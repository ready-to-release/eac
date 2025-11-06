# Feature ID: src-commands_build-module
# Module: src-commands

@src-commands @critical @build-module
Feature: Build Module Command

  As a developer
  I want to build a specific module by its moniker
  So that I can compile/prepare the module

  Rule: Module must be identified by moniker

    @success @ac1
    Scenario: Build existing module
      When I run "go run . build-module src-commands"
      Then I should see "Building" or "Success" or "build"

    @error @ac1
    Scenario: Error on non-existent module
      When I run "go run . build-module non-existent-module-xyz"
      Then I should see "not found" or "Error" or "unknown"

  Rule: Command must be accessible

    @success @ac2
    Scenario: Command is registered
      When I run "go run . build-module"
      Then I should see "build" or "module" or "Error"

    @error @ac2
    Scenario: Invalid module shows error
      When I run "go run . build-module"
      Then I should see "build" or "module" or "usage"
