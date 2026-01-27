@deps:go @L1 @ov @sequential
Feature: eac-commands_build

  As a developer of the eac platform
  I want to build one or more modules by moniker
  So that I can compile/prepare modules

  Rule: Module must be identified by moniker

    Scenario: Build existing module
      When I run the command "build --dry-run eac-commands"
      Then the exit code is 0
      And I should see "Building" or "dry-run" or "build"

    Scenario: Error on non-existent module
      When I run the command "build --dry-run non-existent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error" or "unknown"

  Rule: Command must be accessible

    Scenario: Command help is available
      When I run the command "build --help"
      Then the exit code is 0
      And I should see "build" or "module" or "Usage"
