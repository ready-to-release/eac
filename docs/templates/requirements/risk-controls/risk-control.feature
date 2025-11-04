# Risk Control Feature Template

Feature: [Control Category] Risk Controls

  [Brief description of the risk controls in this file]
  Source: <Assessment-ID> (e.g., Assessment-2025-001)
  Assessment Date: <YYYY-MM-DD>

  # ========================================
  # Risk Control Scenario Template
  #
  # Each risk control defines WHAT must be controlled (the requirement).
  # Use this pattern for all risk control scenarios.
  # ========================================

  @risk<ID>
  Scenario: RC-<ID> - [Short description of the control]
    Given [context describing the system or situation]
    Then [requirement] MUST [condition or behavior]
    And [additional requirement] MUST [condition or behavior]
    And [additional requirement] MUST [condition or behavior]

  # ========================================
  # Example Risk Controls by Category
  # ========================================

  # Authentication Controls
  @risk1
  Scenario: RC-001 - User authentication required
    Given a system with protected resources
    Then all user access MUST be authenticated
    And authentication MUST occur before granting access
    And failed authentication attempts MUST be logged

  @risk2
  Scenario: RC-002 - Password complexity requirements
    Given a user registration or password change
    Then passwords MUST be at least 12 characters
    And passwords MUST contain uppercase, lowercase, numbers, and symbols
    And weak passwords MUST be rejected

  @risk3
  Scenario: RC-003 - Session timeout requirements
    Given an authenticated user session
    Then sessions MUST timeout after 30 minutes of inactivity
    And users MUST re-authenticate after timeout
    And session expiry MUST be logged

  # Data Protection Controls
  @risk10
  Scenario: RC-010 - Data encryption at rest
    Given sensitive data is stored
    Then all sensitive data MUST be encrypted at rest
    And encryption MUST use AES-256 or stronger
    And encryption keys MUST be stored separately
    And plaintext sensitive data MUST NOT be stored

  @risk11
  Scenario: RC-011 - Data encryption in transit
    Given data is transmitted over network
    Then all data transmission MUST use TLS 1.3 or higher
    And certificate validation MUST be enforced
    And self-signed certificates MUST be rejected in production
    And encryption failures MUST prevent transmission

  @risk12
  Scenario: RC-012 - PII data protection
    Given personal identifiable information (PII) is processed
    Then PII MUST be encrypted both at rest and in transit
    And PII access MUST be logged
    And PII deletion MUST be verifiable
    And PII export MUST be controlled

  # Audit and Traceability Controls
  @risk5
  Scenario: RC-005 - Audit trail completeness
    Given system operations occur
    Then all changes MUST create audit trail entries
    And audit entries MUST include: timestamp, user, action, before/after values
    And audit entries MUST be immutable
    And audit trails MUST be retained for required period

  @risk6
  Scenario: RC-006 - Audit trail integrity
    Given audit trail entries exist
    Then audit entries MUST be protected from modification
    And unauthorized access to audit logs MUST be prevented
    And audit log tampering MUST be detectable
    And audit log access MUST be logged

  @risk7
  Scenario: RC-007 - Change traceability
    Given system changes are made
    Then all changes MUST be traceable to requirements
    And all changes MUST be traceable to authorized users
    And change history MUST be complete and accurate
    And unauthorized changes MUST be prevented

  # Access Control
  @risk20
  Scenario: RC-020 - Role-based access control
    Given users with different roles
    Then access permissions MUST be based on user roles
    And users MUST NOT access resources beyond their role
    And role changes MUST require authorization
    And role assignments MUST be auditable

  @risk21
  Scenario: RC-021 - Least privilege principle
    Given user access requirements
    Then users MUST be granted minimum necessary permissions
    And default permissions MUST be restrictive
    And permission escalation MUST require approval
    And excessive permissions MUST be detected and removed

  # Privacy Controls
  @risk30
  Scenario: RC-030 - Right to access personal data
    Given a data subject request
    Then users MUST be able to access all their personal data
    And data export MUST be in machine-readable format
    And data access MUST be provided within required timeframe
    And data access MUST be logged

  @risk31
  Scenario: RC-031 - Right to erasure
    Given a data subject deletion request
    Then all personal data MUST be deleted
    And deletion MUST be verifiable
    And deletion MUST be logged for audit
    And exceptions to deletion MUST be documented

  # AI/ML Controls
  @risk40
  Scenario: RC-040 - AI model transparency
    Given an AI/ML model in production
    Then model decisions MUST be explainable
    And model training data MUST be documented
    And model versioning MUST be maintained
    And model performance MUST be monitored

  @risk41
  Scenario: RC-041 - AI bias detection
    Given AI/ML model outputs
    Then model predictions MUST be tested for bias
    And bias detection MUST run continuously
    And bias findings MUST be reported
    And biased models MUST be retrained or removed

# ========================================
# Template Instructions
# ========================================
#
# Risk control definitions are REQUIREMENTS from risk assessment documents.
# They define WHAT must be controlled, not HOW to implement it.
#
# Key Principles:
# 1. Risk controls live ONLY in requirements/risk-controls/
# 2. Risk controls have ONLY .feature files (NO acceptance.spec)
# 3. Each control defines a requirement using "MUST" statements
# 4. Implementation features reference these controls with @risk<ID> tags
#
# To use this template:
#
# 1. **Create feature file**:
#    - File location: requirements/risk-controls/<category>-controls.feature
#    - Examples: authentication-controls.feature, data-protection-controls.feature
#
# 2. **Define feature header**:
#    - Feature name: [Category] Risk Controls
#    - Description: Brief explanation of what controls are in this file
#    - Source: Reference to the source assessment document
#    - Assessment Date: Date of the assessment
#
# 3. **Create risk control scenarios**:
#    - Tag: @risk<ID> (e.g., @risk1, @risk5, @risk10)
#    - Name: RC-<ID> - <Description> (e.g., RC-001 - User authentication required)
#    - Use MUST for mandatory requirements
#    - Keep focused on one control per scenario
#
# 4. **Numbering scheme** (suggested):
#    - 1-9: Authentication controls
#    - 10-19: Data protection controls
#    - 5-9: Audit controls
#    - 20-29: Access control
#    - 30-39: Privacy controls
#    - 40-49: AI/ML controls
#    - 50+: Other categories
#
# 5. **Writing effective risk controls**:
#    - Start with Given to set context
#    - Use Then/And with MUST for requirements
#    - Be specific and measurable
#    - Focus on WHAT must happen, not HOW
#    - Avoid implementation details
#
# 6. **Linking to implementations**:
#    - Implementation features tag scenarios with @risk<ID>
#    - Multiple features can implement the same control
#    - One feature can implement multiple controls
#    - See: docs/how-to-guides/testing/link-risk-controls.md
#
# Good Example:
# @risk1
# Scenario: RC-001 - User authentication required
#   Given a system with protected resources
#   Then all user access MUST be authenticated
#   And authentication MUST occur before granting access
#
# Bad Example (too implementation-specific):
# @risk1
# Scenario: RC-001 - Use JWT tokens for auth
#   Given a user login request
#   Then the system MUST use JWT tokens
#   And tokens MUST expire after 30 minutes
#   # ❌ This specifies HOW (JWT), not WHAT (authentication required)
#
# Remember:
# - Risk controls are REQUIREMENTS, not implementations
# - They define WHAT must be controlled
# - Implementation features define HOW
# - Keep controls technology-agnostic when possible
#
# For more information:
# - How-to guide: docs/how-to-guides/testing/link-risk-controls.md
# - BDD reference: docs/reference/testing/bdd-format.md
# - Examples: Look for @risk tags in requirements/<module>/ directories
