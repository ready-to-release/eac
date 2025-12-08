@deps:go @control:cm-2 @control:sa-10
Feature: repository_build-repo-config

  As a repository maintainer
  I want to ensure the repository module builds successfully
  So that repository configuration files are valid and can be used by the system

  Background:
    Given the repository root exists

  Rule: The repository module must build without errors

    @L1 @ov
    Scenario: repository module builds successfully
      When I run the command "build repository --dry-run"
      Then the exit code is 0
      And I should not see any build errors
