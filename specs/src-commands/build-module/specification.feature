# Feature ID: src-commands_build-module
# Module: src-commands

@src-commands @critical @build-module
Feature: Build Module Command

  As a developer
  I want to build a specific module by its moniker
  So that I can compile/prepare the module based on its type

  Background:
    Given the repository contains module contracts
    And each module has a defined type

  Rule: Module must be identified by moniker

    @success @ac1
    Scenario: Build module by valid moniker
      Given a module with moniker "src-commands" exists
      And the module type is "go-library"
      When I run "build module src-commands"
      Then the build should dispatch to the go-library build function
      And the build should execute successfully
      And the output should indicate build success

    @error @ac1
    Scenario: Build non-existent module
      Given no module with moniker "non-existent" exists
      When I run "build module non-existent"
      Then the command should fail
      And the error message should indicate "module not found: non-existent"

  Rule: Build function must be dispatched based on module type

    @success @ac2
    Scenario: Dispatch go-library build
      Given a module with type "go-library" exists
      When I run "build module <moniker>"
      Then the command should execute "go build" in the module root
      And the build artifacts should be created

    @success @ac2
    Scenario: Dispatch vscode-ext build
      Given a module with type "vscode-ext" exists
      When I run "build module <moniker>"
      Then the command should execute "npm install" in the module root
      And the command should execute "npm run compile" in the module root
      And the build artifacts should be created

    @success @ac2
    Scenario: Dispatch mkdocs-site build
      Given a module with type "mkdocs-site" exists
      When I run "build module <moniker>"
      Then the command should execute the MkDocs build process
      And the static site should be generated

    @error @ac2
    Scenario: Unsupported module type
      Given a module with type "no-module-type" exists
      When I run "build module <moniker>"
      Then the command should fail
      And the error message should indicate "no build function for type: no-module-type"

  Rule: Build function must execute in module root directory

    @success @ac3
    Scenario: Execute build in correct directory
      Given a module with root "src/commands" exists
      When I run "build module src-commands"
      Then the build command should execute in directory "src/commands"
      And the working directory should be the module root

  Rule: Build errors must be reported clearly

    @error @ac4
    Scenario: Build command fails
      Given a module with type "go-library" exists
      And the go build command will fail
      When I run "build module <moniker>"
      Then the command should fail
      And the error output from the build should be displayed
      And the exit code should be non-zero

    @success @ac4
    Scenario: Build succeeds with warnings
      Given a module with type "go-library" exists
      And the go build command will succeed with warnings
      When I run "build module <moniker>"
      Then the command should succeed
      And the warnings should be displayed
      And the exit code should be zero
