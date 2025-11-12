# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53
# @implementation:required @automation:full

@risk-control:auth-sso
Feature: Single Sign-On Security

  As a system administrator
  I want to implement Single Sign-On Security
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:auth-sso-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:auth-sso-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
