@L2
Feature: src-core_execution-control-tags

  As a test framework
  I want to support execution control tags (@ignore, @Manual)
  So that I can exclude WIP tests and handle manual test execution

  Background:
    Given the tag contract is loaded from "contracts/testing/0.1.0/tags.yml"

  # ============================================================================
  # @ignore Tag
  # ============================================================================

  Rule: @ignore excludes tests from all suites

    @ov
    Scenario: Ignored test excluded from pre-commit suite
      Given a Godog feature file:
        """
        @ignore @L2 @ov
        Feature: WIP Feature
          Scenario: Test in progress
            When I test something
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      Then the ignored test should NOT be selected

    @ov
    Scenario: @ignore evaluated before other selectors
      Given a Godog feature file:
        """
        @ignore @L2 @ov
        Feature: WIP Feature
          Scenario: Test in progress
            When I test something
        """
      When I discover tests from the feature
      And I select tests for suite "acceptance"
      Then the ignored test should NOT be selected

    @ov
    Scenario: Feature-level @ignore excludes all scenarios
      Given a Godog feature file:
        """
        @ignore @L2
        Feature: WIP Feature

          @ov
          Scenario: Test 1
            When I test something

          @ov
          Scenario: Test 2
            When I test something else
        """
      When I discover tests from the feature
      Then both scenarios should have "@ignore" tag
      When I select tests for suite "pre-commit"
      Then no scenarios should be selected

    @ov
    Scenario: Scenario-level @ignore excludes only that scenario
      Given a Godog feature file:
        """
        @L2
        Feature: Mixed Feature

          @ov
          Scenario: Stable test
            When I test something

          @ignore @ov
          Scenario: WIP test
            When I test something else
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      Then only "Stable test" should be selected

  Rule: @ignore detection sets IsIgnored field

    @ov
    Scenario: Godog scenario with @ignore sets IsIgnored true
      Given a Godog feature file:
        """
        @ignore @L2 @ov
        Feature: WIP Feature
          Scenario: Test in progress
            When I test something
        """
      When I discover tests from the feature
      Then test "Test in progress" should have IsIgnored = true

    @ov
    Scenario: Godog scenario without @ignore sets IsIgnored false
      Given a Godog feature file:
        """
        @L2 @ov
        Feature: Active Feature
          Scenario: Working test
            When I test something
        """
      When I discover tests from the feature
      Then test "Working test" should have IsIgnored = false

  # ============================================================================
  # @Manual Tag
  # ============================================================================

  Rule: @Manual detection sets IsManual field

    @ov
    Scenario: Godog scenario with @Manual sets IsManual true
      Given a Godog feature file:
        """
        @L2
        Feature: Manual Testing
          @Manual @ov
          Scenario: Manual verification test
            When I test manually
        """
      When I discover tests from the feature
      Then test "Manual verification test" should have IsManual = true

    @ov
    Scenario: Godog scenario without @Manual sets IsManual false
      Given a Godog feature file:
        """
        @L2 @ov
        Feature: Automated Testing
          Scenario: Automated test
            When I test automatically
        """
      When I discover tests from the feature
      Then test "Automated test" should have IsManual = false

  Rule: @Manual tests are included in suite selection

    @ov
    Scenario: Manual tests included in pre-commit suite
      Given a Godog feature file:
        """
        @L2
        Feature: Manual Testing
          @Manual @ov
          Scenario: Manual verification test
            When I test manually
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      Then test "Manual verification test" should be selected

    @ov
    Scenario: Manual tests can be filtered separately
      Given a Godog feature file:
        """
        @L2
        Feature: Mixed Testing

          @ov
          Scenario: Automated test
            When I test automatically

          @Manual @ov
          Scenario: Manual test
            When I test manually
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      And I filter for manual tests only
      Then only "Manual test" should be selected

  # ============================================================================
  # Tag Validation
  # ============================================================================

  Rule: @ignore and @Manual are valid tags

    @ov
    Scenario: @ignore passes validation
      Given a test with tags ["@ignore", "@L2", "@ov"]
      When I validate the tags
      Then validation should pass

    @ov
    Scenario: @Manual passes validation
      Given a test with tags ["@Manual", "@L2", "@ov"]
      When I validate the tags
      Then validation should pass

  # ============================================================================
  # Combined Usage
  # ============================================================================

  Rule: @ignore and @Manual can be combined

    @ov
    Scenario: Test can be both ignored and manual
      Given a Godog feature file:
        """
        @L2
        Feature: WIP Manual Testing
          @ignore @Manual @ov
          Scenario: Manual test in progress
            When I test manually
        """
      When I discover tests from the feature
      Then test "Manual test in progress" should have IsIgnored = true
      And test "Manual test in progress" should have IsManual = true
      When I select tests for suite "pre-commit"
      Then test "Manual test in progress" should NOT be selected
