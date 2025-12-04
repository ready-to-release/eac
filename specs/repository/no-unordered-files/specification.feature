@control:sa-3
Feature: repository_no-unordered-files

  As a repository maintainer
  I want to ensure no files are left without module ownership
  So that the codebase remains organized and maintainable

  Background:
    Given the repository root exists

  Rule: All files must be organized into proper modules

    @L1 @ov
    Scenario: No files are orphaned without module ownership
      Given the module contracts are loaded
      When I check for orphan files
      Then no files should be orphaned
      And if any orphan files are found, I should see their paths with counts
