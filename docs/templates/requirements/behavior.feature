# Feature ID: <component>_<feature_name>
# Acceptance Spec: acceptance.spec
# Component: <Component Name>
# Description: [Brief description of what this feature does]
# Risk Control: @risk<ID> (optional - if this feature implements a risk control)

@<component> @critical @<feature_name>
Feature: [Feature Name - User-Facing Description]

  As a [role/persona]
  I want [capability/functionality]
  So that [business value/benefit]

  Background:
    Given [common precondition for all scenarios in this feature]
    And [additional setup that applies to all scenarios]

  # ========================================
  # Installation Verification (IV)
  # Scenarios tagged with @IV verify installation, setup, configuration, and deployment
  # These appear in "Installation Verification" section of implementation reports
  # ========================================

  @success @ac1 @IV
  Scenario: [Installation/Setup - Success Case]
    Given [precondition describing initial state - e.g., "a clean environment"]
    And [additional context if needed]
    When [installation or setup action performed]
    Then [installation verification step - e.g., "the system should be installed"]
    And [configuration should be correct - e.g., "configuration file exists"]
    And [version verification - e.g., "version matches expected"]
    And [exit code or status verification]

  @error @ac1 @IV
  Scenario: [Installation/Setup - Error Case]
    Given [precondition describing problematic state - e.g., "system already installed"]
    When [installation action attempted]
    Then [expected error behavior - e.g., "installation should fail"]
    And [error message verification - e.g., "error message indicates conflict"]
    And [exit code or status verification]

  @success @ac1 @IV
  Scenario: [Configuration Verification]
    Given [system is installed]
    When [configuration is checked]
    Then [configuration should be valid]
    And [all required settings should be present]
    And [settings should have expected values]

  # ========================================
  # Operational Verification (OV)
  # Scenarios WITHOUT @IV or @PV tags verify functional behavior
  # These appear in "Operational Verification" section of implementation reports
  # ========================================

  @success @ac2
  Scenario: [Functional Behavior - Happy Path]
    Given [precondition describing ready state - e.g., "system is ready"]
    And [additional setup - e.g., "required data is available"]
    When [user action - e.g., "user invokes operation"]
    Then [expected functional outcome - e.g., "operation succeeds"]
    And [verification of expected behavior - e.g., "results are correct"]
    And [output verification - e.g., "output contains expected data"]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Functional Behavior - With Options/Parameters]
    Given [precondition]
    When [action with specific options - e.g., "operation with custom parameters"]
    Then [expected outcome with options applied]
    And [verification of option effects - e.g., "output reflects parameters"]
    And [exit code or status verification]

  @error @ac2
  Scenario: [Functional Behavior - Error Handling]
    Given [precondition that will cause error - e.g., "invalid input is provided"]
    When [user action with invalid input]
    Then [operation should fail gracefully]
    And [error message should be clear and actionable]
    And [verification that system state is safe/unchanged]
    And [exit code or status verification]

  @error @ac2
  Scenario: [Functional Behavior - Missing Required Input]
    Given [precondition]
    When [action without required input]
    Then [operation should fail]
    And [error message should indicate what is missing]
    And [error message should provide guidance]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Functional Behavior - Multiple Operations]
    Given [precondition]
    When [first operation is performed]
    And [second operation is performed]
    Then [both operations should succeed]
    And [combined results should be correct]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Functional Behavior - Data Processing]
    Given [input data is available]:
      """
      [example multi-line input data]
      [line 2]
      [line 3]
      """
    When [data processing operation is performed]
    Then [output data should be produced]
    And [output should contain expected content]
    And [data transformations should be correct]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Functional Behavior - State Transitions]
    Given [system is in initial state]
    When [action that changes state is performed]
    Then [system should transition to expected state]
    And [state should be persisted]
    And [state should be verifiable]
    And [exit code or status verification]

  # ========================================
  # Performance Verification (PV)
  # Scenarios tagged with @PV verify performance requirements, response times, and resource usage
  # These appear in "Performance Verification" section of implementation reports
  # ========================================

  @success @ac3 @PV
  Scenario: [Performance - Response Time Requirement]
    Given [precondition with typical data size]
    When [operation is performed]
    Then the operation should complete within [X] seconds
    And [functional verification that it worked correctly]
    And [exit code or status verification]

  @success @ac3 @PV
  Scenario: [Performance - Large Data Set]
    Given [precondition with large amount of data]
    And the system has [Y] items to process
    When [operation is performed]
    Then the operation should complete within [X] seconds
    And all [Y] items should be processed correctly
    And memory usage should remain under [Z]MB
    And [exit code or status verification]

  @success @ac3 @PV
  Scenario: [Performance - Concurrent Operations]
    Given [precondition]
    When [N] concurrent instances of [operation] are executed
    Then all instances should complete within [X] seconds
    And no instance should fail
    And [verification of correctness under load]
    And [resource usage should remain acceptable]

  @success @ac3 @PV
  Scenario: [Performance - Resource Usage]
    Given [precondition]
    When [operation with large input is performed]
    Then memory usage should not exceed [X]MB
    And CPU usage should remain under [Y]%
    And [functional verification]
    And [exit code or status verification]

  @success @ac3 @PV
  Scenario: [Performance - Throughput]
    Given [system is under load]
    When [operations are performed continuously]
    Then throughput should be at least [X] operations per second
    And response time should remain under [Y] seconds
    And error rate should remain under [Z]%

  # ========================================
  # Edge Cases and Boundary Conditions
  # Tag these as OV (no @IV or @PV) unless they specifically test installation or performance
  # ========================================

  @success @ac2
  Scenario: [Edge Case - Empty Input]
    Given [precondition with empty/minimal data]
    When [operation is performed with empty input]
    Then [expected behavior with empty input - e.g., "appropriate default is used"]
    And [no errors should occur]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Edge Case - Maximum Allowed Input]
    Given [precondition at upper boundary]
    When [operation with max allowed value is performed]
    Then [expected behavior at boundary - e.g., "operation succeeds"]
    And [results should be correct]
    And [exit code or status verification]

  @error @ac2
  Scenario: [Edge Case - Beyond Maximum Allowed Input]
    Given [precondition]
    When [operation with value exceeding maximum is attempted]
    Then [operation should fail]
    And [error message should indicate limit exceeded]
    And [error message should state the limit]
    And [exit code or status verification]

  @success @ac2
  Scenario: [Edge Case - Special Characters]
    Given [precondition]
    When [operation with special characters is performed]
    Then [special characters should be handled correctly]
    And [no encoding issues should occur]
    And [exit code or status verification]

  # ========================================
  # Integration Scenarios
  # When interacting with external systems (databases, APIs, services)
  # ========================================

  @integration @success @ac2
  Scenario: [Integration - External System Interaction]
    Given [external system is available and configured]
    And [precondition]
    When [operation that interacts with external system is performed]
    Then [expected interaction outcome]
    And [verification of data in external system]
    And [exit code or status verification]

  @integration @error @ac2
  Scenario: [Integration - External System Unavailable]
    Given [external system is unavailable]
    When [operation requiring external system is attempted]
    Then [operation should fail gracefully]
    And [error message should indicate connectivity issue]
    And [system should not be left in inconsistent state]
    And [exit code or status verification]

  @integration @success @ac2
  Scenario: [Integration - Data Synchronization]
    Given [data exists in external system]
    When [synchronization operation is performed]
    Then [data should be synchronized correctly]
    And [no data should be lost]
    And [timestamps should be updated]
    And [exit code or status verification]

  # ========================================
  # Scenario Outlines (for repetitive test cases)
  # Use these to reduce duplication when testing multiple similar inputs
  # ========================================

  @success @ac2
  Scenario Outline: [Parameterized Test - Multiple Valid Inputs]
    Given [precondition]
    When [operation] is performed with "<input>"
    Then [expected outcome]
    And output should contain "<expected_output>"
    And [exit code or status verification]

    Examples:
      | input       | expected_output        | description                |
      | value1      | result1                | Basic valid input          |
      | value2      | result2                | Alternative valid input    |
      | value3      | result3                | Another valid case         |
      | edge_case   | edge_result            | Edge case within bounds    |

  @error @ac2
  Scenario Outline: [Parameterized Test - Multiple Invalid Inputs]
    Given [precondition]
    When [operation] is performed with "<invalid_input>"
    Then [operation should fail]
    And error message should contain "<error_message>"
    And [exit code verification with "<exit_code>"]

    Examples:
      | invalid_input | error_message              | exit_code | description           |
      | invalid1      | invalid format             | 1         | Wrong format          |
      | invalid2      | out of range               | 1         | Value too large       |
      | ""            | required value missing     | 1         | Empty input           |

  @success @ac2
  Scenario Outline: [Parameterized Test - Different Configurations]
    Given system is configured with "<config_option>"
    When [operation] is performed
    Then result should match "<expected_result>"
    And behavior should reflect configuration

    Examples:
      | config_option | expected_result | description                    |
      | option_a      | result_a        | Configuration A behavior       |
      | option_b      | result_b        | Configuration B behavior       |
      | default       | result_default  | Default configuration behavior |

  # ========================================
  # Security Scenarios (if applicable)
  # ========================================

  @security @error @ac2
  Scenario: [Security - Unauthorized Access Attempt]
    Given [user is not authenticated]
    When [operation requiring authentication is attempted]
    Then [operation should be rejected]
    And error message should indicate authentication required
    And [no sensitive information should be leaked]

  @security @success @ac2
  Scenario: [Security - Authorized Access]
    Given [user is authenticated with appropriate permissions]
    When [protected operation is performed]
    Then [operation should succeed]
    And [audit log should be created]

  @security @error @ac2
  Scenario: [Security - Input Validation Against Injection]
    Given [precondition]
    When [operation with potentially malicious input is attempted]
    Then [malicious input should be sanitized or rejected]
    And [no injection should occur]
    And [appropriate error or safe result should be returned]

  # ========================================
  # Work in Progress (WIP)
  # Tag scenarios with @wip to exclude them from CI runs
  # Remove @wip tag when scenario is ready
  # ========================================

  @wip @success @ac4
  Scenario: [Future Feature - Not Yet Implemented]
    Given [precondition]
    When [operation for future feature is performed]
    Then [expected behavior when implemented]
    And [verification of future functionality]

# ========================================
# Template Instructions
# ========================================
#
# To use this template:
#
# 1. Replace all placeholders:
#    - <component>: Your component name (e.g., api, ui, service, backend)
#    - <feature_name>: Your feature identifier (e.g., user_auth, data_export)
#    - <Component Name>: Full component name
#    - [descriptive text]: Replace with your specific scenario details
#
# 2. Define Feature ID:
#    - Pattern: <component>_<feature_name>
#    - Examples: api_user_authentication, ui_dashboard_widgets, service_notifications
#
# 3. Tag scenarios appropriately:
#    - @IV: Installation, setup, configuration, deployment scenarios
#    - @PV: Performance, response time, resource usage scenarios
#    - (no tag): Operational/functional scenarios (most common)
#    - @ac1, @ac2, @ac3: Link to acceptance criteria in acceptance.spec
#    - @risk<ID>: Link to risk control requirements (optional, e.g., @risk1, @risk5)
#      Risk controls are defined in requirements/risk-controls/
#      See: docs/how-to-guides/testing/link-risk-controls.md
#
# 4. Verification type guidelines:
#    - Installation Verification (@IV): System setup, config, version checks
#    - Operational Verification (default): Functional behavior, business logic, errors
#    - Performance Verification (@PV): Response times, throughput, resource limits
#
# 5. Write clear scenarios using Given/When/Then:
#    - Given: Set up preconditions
#    - When: Perform the action
#    - Then: Verify expected outcomes
#    - And/But: Continue previous step type
#
# 6. Remove unused sections or scenarios that don't apply to your feature
#
# 7. Keep scenarios focused - one behavior per scenario
#
# 8. Use Scenario Outline for multiple similar test cases
#
# For more information, see the testing specifications documentation in your project.
