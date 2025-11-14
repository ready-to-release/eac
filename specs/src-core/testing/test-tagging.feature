@L2
Feature: src-core_test-tagging

  As a test framework
  I want to discover and infer test tags
  So that I can select and run appropriate test suites

  Background:
    Given the tag contract is loaded from "contracts/testing/0.1.0/tags.yml"

  # ============================================================================
  # Discovery
  # ============================================================================

  Rule: Must discover tags from Godog feature files

    @L2 @ov
    Scenario: Discover feature-level tag
      Given a Godog feature file:
        """
        @deps:go
        Feature: Build Module
          Scenario: Build command
            When I run build
        """
      When I discover tests from the feature
      Then I should find 1 test
      And test "Build command" should have tags:
        | @deps:go |
        | @L2     |

    @L2 @ov
    Scenario: Feature tags inherited by scenario
      Given a Godog feature file:
        """
        @deps:go @L3
        Feature: Deployment Tests
          Scenario: Deploy to PLTE
            When I deploy
        """
      When I discover tests from the feature
      Then test "Deploy to PLTE" should have tags:
        | @deps:go |
        | @L3     |

    @L2 @ov
    Scenario: Scenario tags override feature level tags
      Given a Godog feature file:
        """
        @L2
        Feature: Tests
          @L3
          Scenario: Deployment test
            When I test
        """
      When I discover tests from the feature
      Then test "Deployment test" should have tag "@L3"
      And test "Deployment test" should NOT have tag "@L2"

  Rule: Must discover tags from Go test files

    @L2 @ov
    Scenario: Go test without build tag defaults to L1
      Given a Go test file "service_test.go" without build tags:
        """
        package service

        func TestValidateEmail(t *testing.T) {
          // test implementation
        }
        """
      When I discover tests from the file
      Then I should find 1 test
      And test "TestValidateEmail" should have tag "@L1"

    @L2 @ov
    Scenario: Go test with L0 build tag
      Given a Go test file "service_fast_test.go" with build tag "L0":
        """
        //go:build L0
        // +build L0

        package service

        func TestQuickValidation(t *testing.T) {
          // very fast test
        }
        """
      When I discover tests from the file
      Then test "TestQuickValidation" should have tag "@L0"

  # ============================================================================
  # Inference
  # ============================================================================

  Rule: Godog features without level tag default to L2

    @L2 @ov
    Scenario: Godog feature without level gets L2
      Given a Godog feature file:
        """
        @deps:go
        Feature: Integration Tests
          Scenario: Test integration
        """
      When I apply inferences
      Then test "Test integration" should have tag "@L2"

  Rule: Go tests without level tag default to L1

    @L2 @ov
    Scenario: Go test without level gets L1
      Given a Go test file without build tags
      And the test has no level tags
      When I apply inferences
      Then the test should have tag "@L1"

  Rule: Verification tags infer taxonomy levels

    @L2 @ov
    Scenario: Installation verification infers L3
      Given a Godog feature file:
        """
        @iv
        Feature: Installation Verification
          Scenario: Check installation
        """
      When I apply inferences
      Then test "Check installation" should have tags:
        | @iv |
        | @L3 |

    @L2 @ov
    Scenario: Performance verification infers L3
      Given a Godog feature file:
        """
        @pv
        Feature: Performance Tests
          Scenario: Load test
        """
      When I apply inferences
      Then test "Load test" should have tags:
        | @pv |
        | @L3 |

    @L2 @ov
    Scenario: Production installation verification infers L4
      Given a Godog feature file:
        """
        @piv
        Feature: Production Smoke Tests
          Scenario: Health check
        """
      When I apply inferences
      Then test "Health check" should have tags:
        | @piv |
        | @L4  |

    @L2 @ov
    Scenario: Production performance verification infers L4
      Given a Godog feature file:
        """
        @ppv
        Feature: Production Performance
          Scenario: Monitor performance
        """
      When I apply inferences
      Then test "Monitor performance" should have tags:
        | @ppv |
        | @L4  |

  Rule: Operational verification is derived

    @L2 @ov
    Scenario: Test without verification tags gets @ov
      Given a Godog feature file:
        """
        Feature: Feature Tests
          Scenario: Test feature
        """
      When I apply inferences
      Then test "Test feature" should have tag "@ov"

    @L2 @ov
    Scenario: Test with @iv does not get @ov
      Given a Godog feature file:
        """
        @iv
        Feature: Installation Tests
          Scenario: Install check
        """
      When I apply inferences
      Then test "Install check" should have tag "@iv"
      And test "Install check" should NOT have tag "@ov"

  Rule: Explicit level tags override inferences

    @L2 @ov
    Scenario: Explicit L2 overrides Godog default
      Given a Godog feature file:
        """
        @L2
        Feature: Custom Level
          Scenario: Test
        """
      When I apply inferences
      Then test "Test" should have tag "@L2"
      And there should be only one level tag

    @L2 @ov
    Scenario: Explicit L2 with @pv keeps explicit level
      Given a Godog feature file:
        """
        @L2 @pv
        Feature: Performance at L2
          Scenario: Test
        """
      When I apply inferences
      Then test "Test" should have tags:
        | @L2 |
        | @pv |
      And test "Test" should NOT have tag "@L3"

  # ============================================================================
  # Suite Selection
  # ============================================================================

  Rule: Suite selectors filter tests by tags

    @L2 @ov
    Scenario: Pre-commit suite selects L0-L2 tests
      Given the following tests:
        | name    | tags          |
        | Test A  | @L0, @ov      |
        | Test B  | @L1, @ov      |
        | Test C  | @L2, @ov      |
        | Test D  | @L3, @iv      |
      When I select tests for suite "pre-commit"
      Then tests should be selected:
        | Test A |
        | Test B |
        | Test C |
      And tests should NOT be selected:
        | Test D |

    @L2 @ov
    Scenario: Acceptance suite selects by verification type
      Given the following tests:
        | name    | tags          |
        | Test A  | @L3, @iv      |
        | Test B  | @L3, @ov      |
        | Test C  | @L3, @pv      |
        | Test D  | @L4, @piv     |
      When I select tests for suite "acceptance"
      Then tests should be selected:
        | Test A |
        | Test B |
        | Test C |
      And tests should NOT be selected:
        | Test D |

    @L2 @ov
    Scenario: Production verification suite requires both L4 and @piv
      Given the following tests:
        | name    | tags          |
        | Test A  | @L4, @piv     |
        | Test B  | @L4, @ov      |
        | Test C  | @L3, @piv     |
      When I select tests for suite "production-verification"
      Then tests should be selected:
        | Test A |
      And tests should NOT be selected:
        | Test B |
        | Test C |

  # ============================================================================
  # System Dependencies
  # ============================================================================

  Rule: Must extract system dependencies from tests

    @L2 @ov
    Scenario: Extract dependencies from selected tests
      Given the following tests:
        | name    | tags                  |
        | Test A  | @L1, @deps:go          |
        | Test B  | @L2, @deps:docker      |
        | Test C  | @L3, @deps:git         |
      When I get system dependencies
      Then I should get dependencies:
        | @deps:go     |
        | @deps:docker |
        | @deps:git    |

    @L2 @ov
    Scenario: No duplicate dependencies
      Given the following tests:
        | name    | tags                          |
        | Test A  | @L1, @deps:go                  |
        | Test B  | @L1, @deps:go, @deps:docker     |
      When I get system dependencies
      Then I should get dependencies:
        | @deps:go     |
        | @deps:docker |
      And there should be 2 unique dependencies
