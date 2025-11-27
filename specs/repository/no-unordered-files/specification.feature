Feature: repository_no-unordered-files

  As a repository maintainer
  I want to ensure no files are left in unordered state
  So that the codebase remains organized and maintainable

  Background:
    Given the repository root exists

  Rule: All files must be organized into proper modules

    @L2 @ov
    Scenario: No files belong to the unordered catch-all module
      Given the module contracts are loaded
      When I lookup files belonging to the "unordered" module
      Then the file list should be empty
      And if any files are found, I should see their paths with counts
