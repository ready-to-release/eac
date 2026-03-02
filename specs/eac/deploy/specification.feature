# Intent: Enable developers to deploy modules to target environments using configured deployer tools
# Architecture: Affects eac-cli deploy command; resolves module + environment from registry and environments.yml; uses deploy bridge for tool dispatch

@deps:go @L1 @ov @sequential
Feature: eac-cli_deploy

  As a developer of the eac platform
  I want to deploy modules to target environments
  So that I can provision infrastructure and applications

  Rule: Deploy requires a module moniker and environment

    Scenario: Deploy help is available
      When I run the command "deploy --help"
      Then the exit code is 0
      And I should see "deploy" or "module" or "environment"

    Scenario: Error when no arguments provided
      When I run the command "deploy"
      Then the exit code is 1
      And I should see "requires" or "argument" or "Error"

    Scenario: Error on non-existent module
      When I run the command "deploy non-existent-module-xyz development --no-tui"
      Then the exit code is 1
      And I should see "not found" or "Error" or "unknown"

    Scenario: Error on unknown environment
      When I run the command "deploy infra non-existent-env-xyz --no-tui"
      Then the exit code is 1
      And I should see "unknown environment" or "Error"

  Rule: Dry-run mode previews changes without applying

    Scenario: Dry-run flag is accepted
      When I run the command "deploy --help"
      Then the exit code is 0
      And I should see "dry-run" or "what-if" or "Preview"

  Rule: Component filter restricts deployment scope

    Scenario: Component flag is accepted
      When I run the command "deploy --help"
      Then the exit code is 0
      And I should see "component"
