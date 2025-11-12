# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @sox @gdpr
# @implementation:required @automation:partial

@risk-control:doc-retention
Feature: Document Retention

  As a system administrator
  I want to implement Document Retention
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:doc-retention-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:doc-retention-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
