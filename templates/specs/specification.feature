# ============================================================================
# ARCHITECTURAL NOTE:
# - This file: specs/<module>/<feature>/specification.feature
# - Step definitions: src/<module>/tests/steps_test.go (SEPARATE LOCATION)
# - Test runner: src/<module>/tests/godog_test.go
#
# This template is for SPECIFICATIONS ONLY (business-readable WHAT).
# Implementation code (HOW) goes in src/, not in specs/.
# ============================================================================
#
# INSTRUCTIONS:
# 1. Replace [placeholders] with actual content
# 2. Rules represent acceptance criteria (ATDD layer)
# 3. Scenarios under Rules are executable examples (BDD layer)
# 4. Save this file in specs/<module>/<feature>/
# 5. Implement step definitions separately in src/<module>/tests/steps_test.go
#
# TAG USAGE:
# - Scenario level: REQUIRED verification tag (@ov, @iv, @pv, @piv, @ppv)
# - Optional: Test level (@L2, @L3, @L4), dependencies (@dep:docker),
#   risk controls (@risk-control:<name>-<id>)
# - See docs/explanation/specifications/tag-reference.md for complete taxonomy

Feature: [module-name_feature-name]

  As a [role]
  I want [capability]
  So that [business value]

  Background:
    Given [common precondition]
    And [common setup]

  Rule: [Acceptance Criterion 1 - Business Rule]

    # ATDD: This Rule defines an acceptance criterion
    # BDD: Scenarios below provide executable examples
    #
    # Tag Guidelines (testing taxonomy tags only):
    # - @ov (operational verification) - REQUIRED for all functional tests
    # - @iv (installation verification) - Use for deployment/smoke tests in PLTE
    # - @pv (performance verification) - Use for performance tests in PLTE
    # - @L2 (default) - Emulated system test
    # - @L3 - In-situ vertical test (PLTE) - auto-inferred from @iv/@pv
    # - @dep:* - Declare system dependencies (docker, git, etc.)

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
    # Format: @risk-control:<control-name>-<scenario-id>
    # See docs/explanation/specifications/risk-controls.md for details
    @ov @risk-control:[control-name]-01
    Scenario: [Risk control for AC2]
      Given [security precondition]
      When [authenticated action]
      Then [access granted]
      And [audit logged]

  Rule: [Acceptance Criterion 3 - With System Dependencies]

    # Example: Test requiring Docker (declare with @dep:docker)
    # Dependencies checked in CI (fail) and local dev (warn+skip)
    @L2 @dep:docker @ov
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
# System Dependencies (declare required tooling):
#   @dep:docker   - Docker engine required
#   @dep:git      - Git CLI required
#   @dep:go       - Go toolchain required
#   @dep:claude   - Claude API access required
#   @dep:az-cli   - Azure CLI required
#
# Risk Controls (compliance traceability):
#   @risk-control:<control-name>-<id>
#   Example: @risk-control:auth-mfa-01
#   Links to: specs/risk-controls/<control-name>.feature
#
# NOTE: This template uses ONLY testing taxonomy tags from tag-reference.md.
# For organizational tags (@cli, @critical, @ac1, etc.), see:
# docs/explanation/specifications/gherkin-concepts.md
#
# TAG INHERITANCE:
# - Feature tags accumulate to Rules and Scenarios
# - Test level tags can be overridden at scenario level
# - Dependencies and verification tags accumulate (additive)
#
# TEST SUITES (tag-based selection):
#   pre-commit:  @L0 + @L1 + @L2 (5-30 min)
#   acceptance:  @iv + @ov + @pv (1-8 hours, PLTE)
#   production:  @piv + @ppv (continuous, production)
# ============================================================================
