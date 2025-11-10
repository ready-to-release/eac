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
# 4. Link scenarios to Rules using @ac1, @ac2, etc.
# 5. Save this file in specs/<module>/<feature>/
# 6. Implement step definitions separately in src/<module>/tests/steps_test.go

@module @critical
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

    @success @ac1 @IV
    Scenario: [Happy path for AC1 - Installation]
      Given [specific precondition]
      When [installation/setup action]
      Then [installation verified]
      And [configuration verified]

    @success @ac1
    Scenario: [Happy path for AC1 - Operational]
      Given [specific precondition]
      When [user action]
      Then [observable outcome]
      And [verification]

    @error @ac1
    Scenario: [Error case for AC1]
      Given [error precondition]
      When [invalid action]
      Then [error behavior]

  Rule: [Acceptance Criterion 2]

    @success @ac2 @PV
    Scenario: [Performance case for AC2]
      Given [performance precondition]
      When [action with load]
      Then [outcome within SLA]
      And [resource usage within limits]

    @success @ac2 @risk1
    Scenario: [Risk control for AC2]
      Given [security precondition]
      When [authenticated action]
      Then [access granted]
      And [audit logged]
