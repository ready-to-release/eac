@dep:go
Feature: src-commands_build-module

  As a developer of the eac platform
  I want to build a specific module by its moniker
  So that I can compile/prepare the module

  Rule: Module must be identified by moniker

    Scenario: Build existing module
      When I run the command "build module src-commands"
      Then the exit code is 0
      And I should see "Building" or "Success" or "build"

    Scenario: Error on non-existent module
      When I run the command "build module non-existent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error" or "unknown"

  Rule: Command must be accessible

    Scenario: Command is registered
      When I run the command "build module"
      Then the exit code is 1
      And I should see "build" or "module" or "Error"

    Scenario: Invalid module shows error
      When I run the command "build module"
      Then the exit code is 1
      And I should see "build" or "module" or "usage"
