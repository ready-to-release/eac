# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @hipaa @pci-dss @fda-21cfr11 @nist-800-53
# @implementation:required @automation:partial

@risk-control:access-remote
Feature: Remote Access Controls

  As a system administrator
  I want to implement Remote Access Controls
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:access-remote-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:access-remote-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
