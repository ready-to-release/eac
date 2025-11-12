# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @nist-800-53
# @implementation:required @automation:full

@risk-control:auth-lockout
Feature: Account Lockout

  As a system administrator
  I want to implement Account Lockout
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:auth-lockout-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:auth-lockout-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
