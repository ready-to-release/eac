@deps:go @L1 @ov @control:sa-11
Feature: repository_test-sanity

  As a developer
  I want test discovery to match raw filesystem scan
  So that I can trust all tests are being discovered

  Background:
    Given I am in the repository root

  Rule: Feature file discovery must be complete

    @L1 @ov
    Scenario: All feature files in specs/ are discovered
      When I scan specs/ for *.feature files
      And I run test discovery
      Then the discovered godog file count matches the raw scan count

  Rule: Go test discovery must be complete

    @L1 @ov
    Scenario: All Go unit test files are discovered
      When I scan for *_test.go files excluding godog_test.go
      And I run test discovery
      Then the discovered gotest file count matches the raw scan count

    @L1 @ov
    Scenario: Godog runner files are NOT counted as tests
      When I scan for godog_test.go files
      Then the count is greater than zero
      And none of those files appear as gotest type in discovery

  Rule: TypeScript test discovery must be complete

    @L1 @ov
    Scenario: All TypeScript test files are discovered
      When I scan for *.test.ts files
      And I run test discovery
      Then the discovered mocha file count matches the raw scan count

    @L1 @ov
    Scenario: TypeScript Cucumber feature files are discovered as tscucumber
      When I run test discovery
      Then at least one test should have type tscucumber
      And all tscucumber tests should reference existing feature files

  Rule: Discovery totals must be verifiable

    @L1 @ov
    Scenario: Total discovered tests equals sum of all types
      When I run test discovery
      Then total count equals godog + gotest + tscucumber + mocha counts

    @L1 @ov
    Scenario: No duplicate test entries exist
      When I run test discovery
      Then each test moniker is unique
