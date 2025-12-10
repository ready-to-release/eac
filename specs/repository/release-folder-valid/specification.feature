@L1 @ov @control:cm-2
Feature: repository_release-folder-valid

  As a repository maintainer
  I want to ensure the release folder only contains valid module changelogs
  So that the release automation does not fail due to orphan directories

  Background:
    Given the repository root exists

  Rule: Release folder must only contain valid module directories

    The release/ directory is watched by the release-auto workflow.
    Each subdirectory must correspond to a registered module with a changelog.
    Orphan directories (not matching any module) will cause CI failures.

    @L1 @ov
    Scenario: All release directories correspond to registered modules
      Given I load all module contracts from the repository
      When I scan the release directory for subdirectories
      Then every release subdirectory should correspond to a registered module
      And if any orphan directories are found, I should see their names

    @L1 @ov
    Scenario: All release changelogs are valid
      Given I load all module contracts from the repository
      When I scan the release directory for changelog files
      Then every changelog should have valid format
      And no unexpected files should exist in release subdirectories
