@deps:python @L1 @env:isolated-test-project
Feature: eac-cli_build-python

  As a developer with Python modules
  I want to build Python modules by moniker
  So that I can verify my Python code compiles and packages correctly

  Rule: Python modules must build successfully

    Scenario: Build python-with-tests template
      Given I use the "python-with-tests" test layout
      When I run the command "build --dry-run calculator"
      Then the exit code is 0
      And I should see "Building" or "dry-run" or "build"

    Scenario: Build python-with-bdd template
      Given I use the "python-with-bdd" test layout
      When I run the command "build --dry-run data-store"
      Then the exit code is 0
      And I should see "Building" or "dry-run" or "build"

    Scenario: Build python-full-stack template
      Given I use the "python-full-stack" test layout
      When I run the command "build --dry-run mathlib"
      Then the exit code is 0
      And I should see "Building" or "dry-run" or "build"

  Rule: Build must fail for non-existent modules

    Scenario: Error on non-existent module
      Given I use the "python-with-tests" test layout
      When I run the command "build --dry-run non-existent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error" or "unknown"
