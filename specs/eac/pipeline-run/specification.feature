@L2 @deps:go @deps:git @ov @env:isolated-test-project
Feature: eac-cli_pipeline-run

  As a developer using the eac platform
  I want to execute module pipelines respecting dependencies
  So that I can build and test modules in the correct order

  Background:
    Given I am in a git repository with EAC configuration

  Rule: Pipeline executes modules in dependency order

    Scenario: Run pipeline for all modules
      Given modules exist with dependencies:
        | moniker | depends_on |
        | core    |            |
        | eac     | core       |
      When I run "pipeline run"
      Then the exit code is 0
      And "core" is processed before "eac"

    Scenario: Run pipeline for specific modules
      Given modules "core" and "clie" exist
      When I run "pipeline run clie"
      Then the exit code is 0
      And only "clie" and its dependencies are processed

  Rule: Changed-only mode filters to modified modules

    Scenario: Run only changed modules
      Given module "core" has uncommitted changes
      And module "clie" has no changes
      When I run "pipeline run --changed-only"
      Then the exit code is 0
      And only "core" is processed

    Scenario: Compare against specific ref
      Given module "core" was changed since "main"
      When I run "pipeline run --changed-only --ref main"
      Then the exit code is 0
      And "core" is included in the pipeline

  Rule: Wait flag polls for completion

    Scenario: Wait for pipeline completion
      Given a pipeline is running
      When I run "pipeline run --wait"
      Then the command waits for completion
      And reports final status

    Scenario: Wait with timeout
      Given a pipeline is running
      When I run "pipeline run --wait --timeout 60"
      Then the command waits up to 60 seconds
      And exits with error if timeout exceeded

  Rule: Error handling

    Scenario: Handle module not found
      When I run "pipeline run nonexistent-module"
      Then the exit code is 1
      And stderr contains "module not found"

    Scenario: Handle circular dependencies
      Given modules with circular dependencies
      When I run "pipeline run"
      Then the exit code is 1
      And stderr contains "circular dependency"
