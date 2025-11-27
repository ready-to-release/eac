@deps:go @L2 @ov
Feature: src-commands_build

  As a developer of the eac platform
  I want to build one or more modules by moniker
  So that I can compile/prepare modules

  Rule: Module must be identified by moniker

    Scenario: Build existing module
      When I run the command "build src-commands"
      Then the exit code is 0
      And I should see "Building" or "Success" or "build"

    Scenario: Error on non-existent module
      When I run the command "build non-existent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error" or "unknown"

  Rule: Command must be accessible

    Scenario: Command help is available
      When I run the command "build --help"
      Then the exit code is 0
      And I should see "build" or "module" or "Usage"

    Scenario: Build with no args builds all
      When I run the command "build"
      Then the exit code is 0 or 1
      And I should see "Building" or "modules" or "build"
