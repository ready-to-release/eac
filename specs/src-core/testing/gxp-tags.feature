@L2
Feature: src-core_gxp-regulatory-tags

  As a test framework
  I want to support GxP regulatory tags
  So that I can validate regulatory requirements and generate compliance reports

  Background:
    Given the tag contract is loaded from "contracts/testing/0.1.0/tags.yml"

  # ============================================================================
  # @gxp Tag Detection
  # ============================================================================

  Rule: @gxp tag is detected and sets IsGxP field

    @ov
    Scenario: Feature with @gxp tag
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ov @gxp @risk-control:gxp-batch-release
          Scenario: Quality manager approves batch
            When batch is approved
        """
      When I discover tests from the feature
      Then test "Quality manager approves batch" should have IsGxP = true

    @ov
    Scenario: Scenario with @gxp tag
      Given a Godog feature file:
        """
        @L2
        Feature: Mixed Feature
          @ov @gxp @risk-control:gxp-auth
          Scenario: GxP scenario
            When I test GxP
        """
      When I discover tests from the feature
      Then test "GxP scenario" should have IsGxP = true

    @ov
    Scenario: Test without @gxp tag
      Given a Godog feature file:
        """
        @L2 @ov
        Feature: Regular Feature
          Scenario: Regular test
            When I test something
        """
      When I discover tests from the feature
      Then test "Regular test" should have IsGxP = false

  # ============================================================================
  # @critical-aspect Tag Detection
  # ============================================================================

  Rule: @critical-aspect tag is detected and sets IsCriticalAspect field

    @ov
    Scenario: Scenario with @critical-aspect tag
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ov @gxp @critical-aspect @risk-control:gxp-batch-release
          Scenario: Critical batch approval
            When batch is approved
        """
      When I discover tests from the feature
      Then test "Critical batch approval" should have IsCriticalAspect = true

    @ov
    Scenario: Test without @critical-aspect tag
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Audit Trail
          @ov @gxp @risk-control:gxp-audit
          Scenario: Audit logging
            When action is logged
        """
      When I discover tests from the feature
      Then test "Audit logging" should have IsCriticalAspect = false

  # ============================================================================
  # GxP Validation Rules
  # ============================================================================

  Rule: @gxp requires @risk-control:gxp-<name> tag

    @ov
    Scenario: GxP test without risk control fails validation
      Given a Godog feature file:
        """
        @gxp @L2 @ov
        Feature: Batch Release
          Scenario: Batch approved
            When batch is approved
        """
      When I discover tests from the feature
      And I validate test "Batch approved"
      Then validation should fail
      And error should mention "GxP requirement must have @risk-control:gxp-<name> tag"

    @ov
    Scenario: GxP test with GxP risk control passes validation
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ov @gxp @risk-control:gxp-batch-release
          Scenario: Batch approved
            When batch is approved
        """
      When I discover tests from the feature
      And I validate test "Batch approved"
      Then validation should pass

    @ov
    Scenario: GxP test with non-GxP risk control fails validation
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ov @gxp @risk-control:auth-mfa-01
          Scenario: Batch approved
            When batch is approved
        """
      When I discover tests from the feature
      And I validate test "Batch approved"
      Then validation should fail
      And error should mention "GxP requirement must have @risk-control:gxp-<name> tag"

  Rule: @critical-aspect requires @gxp tag

    @ov
    Scenario: Critical aspect without @gxp fails validation
      Given a Godog feature file:
        """
        @L2
        Feature: Regular Feature
          @ov @critical-aspect @risk-control:gxp-batch
          Scenario: Test with critical aspect
            When I test something
        """
      When I discover tests from the feature
      And I validate test "Test with critical aspect"
      Then validation should fail
      And error should mention "@critical-aspect must be used with @gxp tag"

    @ov
    Scenario: Critical aspect with @gxp passes validation
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: GxP Feature
          @ov @gxp @critical-aspect @risk-control:gxp-batch
          Scenario: Critical test
            When I test something
        """
      When I discover tests from the feature
      And I validate test "Critical test"
      Then validation should pass

  # ============================================================================
  # GxP Report Generation
  # ============================================================================

  Rule: GxP report summarizes regulatory compliance

    @ov
    Scenario: Generate GxP report with test summary
      Given the following tests:
        | Test Name         | Tags                                                |
        | Auth test         | @L2, @ov                                            |
        | Batch release     | @gxp, @L2, @ov, @risk-control:gxp-batch            |
        | Critical batch    | @gxp, @critical-aspect, @L2, @ov, @risk-control:gxp-batch |
        | Manual test       | @gxp, @Manual, @L2, @ov, @risk-control:gxp-manual  |
      When I generate GxP report
      Then report should show total tests: 4
      And report should show GxP tests: 3
      And report should show critical aspects: 1
      And report should show manual tests: 1

    @ov
    Scenario: GxP report includes traceability matrix
      Given the following tests:
        | Test Name     | Tags                                       |
        | Batch release | @gxp, @L2, @ov, @risk-control:gxp-batch   |
        | Audit trail   | @gxp, @L2, @ov, @risk-control:gxp-audit   |
      When I generate GxP report
      Then report should include traceability matrix
      And matrix should link "Batch release" to "@risk-control:gxp-batch"
      And matrix should link "Audit trail" to "@risk-control:gxp-audit"

  # ============================================================================
  # Combined GxP Tags
  # ============================================================================

  Rule: GxP tags can be combined with other tags

    @ov
    Scenario: GxP manual critical aspect test
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @Manual @ov @gxp @critical-aspect @risk-control:gxp-batch-release
          Scenario: Manual quality verification
            When QA inspects physical batch
        """
      When I discover tests from the feature
      Then test "Manual quality verification" should have IsGxP = true
      And test "Manual quality verification" should have IsCriticalAspect = true
      And test "Manual quality verification" should have IsManual = true
      When I validate test "Manual quality verification"
      Then validation should pass

    @ov
    Scenario: Ignored GxP test
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: Batch Release
          @ignore @ov @gxp @risk-control:gxp-batch-release
          Scenario: WIP batch test
            When batch is processed
        """
      When I discover tests from the feature
      Then test "WIP batch test" should have IsGxP = true
      And test "WIP batch test" should have IsIgnored = true
      When I select tests for suite "acceptance"
      Then test "WIP batch test" should NOT be selected

  # ============================================================================
  # Feature Naming for URS
  # ============================================================================

  Rule: Feature names serve as URS identifiers

    @ov
    Scenario: Extract URS identifier from feature name
      Given a Godog feature file:
        """
        @gxp @L2
        Feature: batch_release-quality-control
          @ov @gxp @risk-control:gxp-batch
          Scenario: Batch approved
            When batch is approved
        """
      When I discover tests from the feature
      Then feature name should be "batch_release-quality-control"
      And feature name should serve as URS identifier
