# @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @gdpr @hipaa
# @implementation:required @automation:partial

@risk-control:auth-biometric
Feature: Biometric Authentication

  As a system administrator
  I want to implement Biometric Authentication
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:auth-biometric-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:auth-biometric-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
