@L1 @verification @unisolated
Feature: repository_one-module-per-file

  As a repository maintainer
  I want to ensure each file belongs to exactly one module
  So that ownership is clear and modules don't overlap

  Background:
    Given the repository root exists

  Rule: Each file must belong to exactly one module

    @L2 @ov
    Scenario: No files have multi-module ownership
      Given the module contracts are loaded
      When I check for files with multi-module ownership
      Then no files should belong to multiple modules
      And if any files have multi-ownership, I should see their paths and conflicting modules
