# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp
# @implementation:required @automation:partial

@risk-control:app-hardening
Feature: Application Hardening

  As a system administrator
  I want to implement Application Hardening
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:app-hardening-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:app-hardening-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
