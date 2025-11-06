# Feature ID: src-cli_cli-invocation
# Module: src-cli

@src-cli @critical
Feature: CLI Invocation

  As a user
  I want to invoke the CLI
  So that I can use the application

  Rule: CLI shows version when requested

    @success @ac1
    Scenario: Version flag displays version
      When I run "r2r --version"
      Then the exit code is 0
