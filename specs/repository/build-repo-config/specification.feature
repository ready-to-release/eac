@deps:go
Feature: repository_build-repo-config

  As a repository maintainer
  I want to ensure the repo-config module builds successfully
  So that repository configuration files are valid and can be used by the system

  Background:
    Given the repository root exists

  Rule: The repo-config module must build without errors

    @L2 @ov
    Scenario: repo-config module builds successfully
      When I run the command "build repo-config"
      Then the exit code is 0
      And I should not see any build errors
