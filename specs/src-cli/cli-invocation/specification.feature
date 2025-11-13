@ov
Feature: src-cli_cli-invocation

  As a user
  I want to invoke the CLI
  So that I can use the application

  Rule: CLI shows version when requested

    Scenario: Version flag displays version
      When I run "r2r --version"
      Then the exit code is 0
