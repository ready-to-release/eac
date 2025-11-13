@L1
Feature: src-core_risk-control-tags

  As a test framework
  I want to support risk control tags
  So that I can link tests to risk control specifications and generate traceability

  Background:
    Given the tag contract is loaded from "contracts/testing/0.1.0/tags.yml"

  # ============================================================================
  # Risk Control Tag Detection
  # ============================================================================

  Rule: Risk control tags are detected and extracted

    @ov
    Scenario: Single risk control tag detected
      Given a Godog feature file:
        """
        @L2
        Feature: Authentication
          @ov @risk-control:auth-mfa-01
          Scenario: Multi-factor authentication required
            When I login with MFA
        """
      When I discover tests from the feature
      Then test "Multi-factor authentication required" should have 1 risk control
      And risk control should be "@risk-control:auth-mfa-01"

    @ov
    Scenario: Multiple risk control tags detected
      Given a Godog feature file:
        """
        @L2
        Feature: Authentication
          @ov @risk-control:auth-mfa-01 @risk-control:audit-log-02
          Scenario: Authentication with audit
            When I login with MFA
        """
      When I discover tests from the feature
      Then test "Authentication with audit" should have 2 risk controls
      And risk controls should include "@risk-control:auth-mfa-01"
      And risk controls should include "@risk-control:audit-log-02"

    @ov
    Scenario: GxP risk control tag detected
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ov @gxp @risk-control:gxp-batch-release
          Scenario: Quality manager approves batch
            When batch is approved
        """
      When I discover tests from the feature
      Then test "Quality manager approves batch" should have 1 risk control
      And risk control should be "@risk-control:gxp-batch-release"

    @ov
    Scenario: Test without risk control has empty list
      Given a Godog feature file:
        """
        @L2 @ov
        Feature: Simple Feature
          Scenario: Simple test
            When I test something
        """
      When I discover tests from the feature
      Then test "Simple test" should have 0 risk controls

  # ============================================================================
  # Risk Control Tag Validation
  # ============================================================================

  Rule: Risk control tags must follow correct format

    @ov
    Scenario: Valid standard risk control tag
      Given a test with tag "@risk-control:auth-mfa-01"
      When I validate the risk control tag
      Then validation should pass

    @ov
    Scenario: Valid GxP risk control tag
      Given a test with tag "@risk-control:gxp-account-lockout"
      When I validate the risk control tag
      Then validation should pass

    @ov
    Scenario: Invalid risk control tag without dash
      Given a test with tag "@risk-control:authmfa01"
      When I validate the risk control tag
      Then validation should fail
      And error should mention "Invalid risk control tag format"

    @ov
    Scenario: Invalid risk control tag with missing ID
      Given a test with tag "@risk-control:auth-mfa-"
      When I validate the risk control tag
      Then validation should fail

    @ov
    Scenario: Invalid risk control tag with missing name
      Given a test with tag "@risk-control:-01"
      When I validate the risk control tag
      Then validation should fail

  # ============================================================================
  # Risk Control Parsing
  # ============================================================================

  Rule: Risk control tags can be parsed into components

    @ov
    Scenario: Parse standard risk control tag
      Given a risk control tag "@risk-control:auth-mfa-01"
      When I parse the risk control reference
      Then control name should be "auth-mfa"
      And scenario ID should be "01"
      And IsGxP should be false

    @ov
    Scenario: Parse GxP risk control tag
      Given a risk control tag "@risk-control:gxp-account-lockout"
      When I parse the risk control reference
      Then control name should be "gxp-account-lockout"
      And scenario ID should be ""
      And IsGxP should be true

  # ============================================================================
  # Traceability Matrix Generation
  # ============================================================================

  Rule: Traceability matrix links tests to risk controls

    @ov
    Scenario: Generate matrix for tests with risk controls
      Given the following tests:
        | Test Name            | Risk Controls                    | Type      |
        | Auth test            | @risk-control:auth-mfa-01        | Automated |
        | Batch release        | @risk-control:gxp-batch-release  | Manual    |
        | Security test        | @risk-control:auth-mfa-01        | Automated |
      When I generate traceability matrix
      Then matrix should contain 3 rows
      And matrix should link "Auth test" to "@risk-control:auth-mfa-01"
      And matrix should link "Batch release" to "@risk-control:gxp-batch-release"
      And matrix should show "Batch release" as "Manual"

    @ov
    Scenario: Generate matrix with multiple risk controls per test
      Given the following tests:
        | Test Name     | Risk Controls                                              |
        | Complex test  | @risk-control:auth-mfa-01, @risk-control:audit-log-02      |
      When I generate traceability matrix
      Then matrix should contain 2 rows
      And both rows should reference "Complex test"

  # ============================================================================
  # Suite Selection with Risk Controls
  # ============================================================================

  Rule: Risk control tags do not affect suite selection

    @ov
    Scenario: Tests with risk controls selected normally
      Given a Godog feature file:
        """
        @L2
        Feature: Authentication
          @ov @risk-control:auth-mfa-01
          Scenario: MFA test
            When I login with MFA
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      Then test "MFA test" should be selected

    @ov
    Scenario: Risk control tags preserved through selection
      Given a Godog feature file:
        """
        @L2
        Feature: Authentication
          @ov @risk-control:auth-mfa-01
          Scenario: MFA test
            When I login with MFA
        """
      When I discover tests from the feature
      And I select tests for suite "pre-commit"
      Then selected test should still have risk control "@risk-control:auth-mfa-01"
