# ============================================================================
# ARCHITECTURAL NOTE:
# - This file: specs/<module>/<feature>/specification.feature
# - Step definitions: go/eac/specs/impl/<module>/steps.go (SEPARATE LOCATION)
# - Test runner: go/eac/specs/impl/<module>/godog_test.go
#
# This template is for SPECIFICATIONS ONLY (business-readable WHAT).
# Implementation code (HOW) goes in go/, not in specs/.
# ============================================================================
#
# INSTRUCTIONS:
# 1. Replace [placeholders] with actual content
# 2. Rules represent acceptance criteria
# 3. Scenarios under Rules are executable examples
# 4. Save this file in specs/<module>/<feature>/
# 5. Implement step definitions separately in go/eac/specs/impl/<module>/steps.go
#
# TAG USAGE:
# - Scenario level: REQUIRED verification tag (@ov, @iv, @pv, @piv, @ppv)
# - Optional: Test level (@L2, @L3, @L4), dependencies (@deps:docker),
#   risk controls (risk-control:NAME-ID format, e.g. @risk-control:auth-mfa-01)
# - See docs/explanation/specifications/tag-reference.md for complete taxonomy

Feature: [module-name_feature-name]

  As a [role]
  I want [capability]
  So that [business value]

  Background:
    Given [common precondition]
    And [common setup]

  Rule: [Acceptance Criterion 1 - Business Rule]

    # Tag Guidelines (testing taxonomy tags only):
    # - @ov (operational verification) - REQUIRED for all functional tests
    # - @iv (installation verification) - Use for deployment/smoke tests in PLTE
    # - @pv (performance verification) - Use for performance tests in PLTE
    # - @L2 (default) - Emulated system test
    # - @L3 - In-situ vertical test (PLTE) - auto-inferred from @iv or @pv
    # - deps:NAME - Declare system dependencies (e.g. @deps:docker, @deps:git)

    # Example: Installation verification test (runs in PLTE, L3)
    @iv
    Scenario: [Happy path for AC1 - Installation]
      Given [specific precondition]
      When [installation/setup action]
      Then [installation verified]
      And [configuration verified]

    # Example: Operational verification test (default L2 - emulated system)
    @ov
    Scenario: [Happy path for AC1 - Operational]
      Given [specific precondition]
      When [user action]
      Then [observable outcome]
      And [verification]

    # Example: Error case (still @ov, tests operational behavior)
    @ov
    Scenario: [Error case for AC1]
      Given [error precondition]
      When [invalid action]
      Then [error behavior]

  Rule: [Acceptance Criterion 2]

    # Example: Performance verification test (runs in PLTE, L3)
    @pv
    Scenario: [Performance case for AC2]
      Given [performance precondition]
      When [action with load]
      Then [outcome within SLA]
      And [resource usage within limits]

    # Example: Risk control scenario (links to compliance requirement)
    # Format: risk-control:CONTROL-NAME-ID (e.g., @risk-control:auth-example-01)
    # See docs/explanation/specifications/risk-controls.md for details
    @ov @risk-control:auth-example-01
    Scenario: [Risk control for AC2]
      Given [security precondition]
      When [authenticated action]
      Then [access granted]
      And [audit logged]

  Rule: [Acceptance Criterion 3 - With System Dependencies]

    # Example: Test requiring Docker (declare with @deps:docker)
    # Dependencies checked in CI (fail) and local dev (warn+skip)
    @L2 @deps:docker @ov
    Scenario: [Container-based test for AC3]
      Given [container precondition]
      When [docker action]
      Then [container outcome]

  Rule: [Acceptance Criterion 4 - Production Testing]

    # Example: Production installation verification (L4)
    # Runs post-deployment in production with controlled side effects
    @piv
    Scenario: [Production smoke test for AC4]
      Given [production environment]
      When [health check action]
      Then [service responds]
      And [monitoring shows healthy]

    # Example: Production performance monitoring (L4)
    # Continuous validation in production
    @ppv
    Scenario: [Production SLA monitoring for AC4]
      Given [production service running]
      When [synthetic monitoring runs]
      Then [response times within SLA]
      And [error rates below threshold]

  Rule: [Acceptance Criterion 5 - Work in Progress]

    # Example: @skip:wip excludes from all test suites
    # TODO: Issue #123 - Complete OAuth integration
    @skip:wip @ov
    Scenario: [OAuth authentication (WIP)]
      Given [OAuth provider configured]
      When [user authenticates with OAuth]
      Then [authentication succeeds]

# ============================================================================
# TAG REFERENCE SUMMARY
# ============================================================================
# For complete documentation, see: docs/explanation/specifications/tag-reference.md
#
# TESTING TAXONOMY TAGS (defined in tag-reference.md):
#
# Verification Tags (REQUIRED for all scenarios):
#   @ov   - Operational Verification (functional tests, L2/L3)
#   @iv   - Installation Verification (smoke tests in PLTE, auto-infers L3)
#   @pv   - Performance Verification (load tests in PLTE, auto-infers L3)
#   @piv  - Production Installation Verification (smoke in production, auto-infers L4)
#   @ppv  - Production Performance Verification (monitoring in production, auto-infers L4)
#
# Test Level Tags (optional, with inference rules):
#   @L0   - Fast unit tests (Go: //go:build L0)
#   @L1   - Unit tests (Go: default, no tag needed)
#   @L2   - Emulated system tests (Godog: default if no level specified)
#   @L3   - In-situ vertical tests in PLTE (auto-inferred from @iv, @pv)
#   @L4   - Testing in production (auto-inferred from @piv, @ppv)
#
# Test Execution Control:
#   skip:REASON - Exclude from test suites (e.g. @skip:wip, @skip:blocked)
#                 Valid reasons: wip, broken, flaky, deprecated, blocked
#
# System Dependencies (declare required tooling):
#   @deps:docker   - Docker engine required
#   @deps:git      - Git CLI required
#   @deps:go       - Go toolchain required
#   @deps:az-cli   - Azure CLI required
#
# Risk Controls (compliance traceability):
#   risk-control:NAME-ID (e.g. @risk-control:auth-mfa-01)
#   Links to: specs/risk-controls/NAME.feature
#
# GxP/Regulatory Tags (for pharmaceutical/medical device development):
#   For complete documentation, see: docs/explanation/specifications/gxp-tagging.md
#   Feature naming:    module_feature-name serves as URS identifier
#   @gxp              - GxP-related requirement (requires risk-control:gxp-NAME)
#   @gmp-critical-aspect  - GmP Critical Aspect (GmP products only)
#   risk-control:gxp-NAME - Link to GxP risk control (e.g. @risk-control:gxp-audit-trail)
#   @Manual           - Manual test scenario (general use, includes GxP contexts)
#
# NOTE: This template uses ONLY testing taxonomy tags from tag-reference.md.
# For organizational tags, see: docs/explanation/specifications/gherkin-concepts.md
#
# TAG INHERITANCE:
# - Feature tags accumulate to Rules and Scenarios
# - Test level tags can be overridden at scenario level
# - Dependencies and verification tags accumulate (additive)
#
# TEST SUITES (tag-based selection):
#   pre-commit:  @L0 + @L1 + @L2 (5-30 min, excludes skip:*)
#   acceptance:  @iv + @ov + @pv (1-8 hours, PLTE, excludes skip:*)
#   production:  @piv + @ppv (continuous, production, excludes skip:*)
#
# NOTE: All test suites automatically exclude tests tagged with skip:REASON
# ============================================================================
