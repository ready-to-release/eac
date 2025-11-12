# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @hipaa
# @implementation:required @automation:partial

@risk-control:network-wifi
Feature: Wireless Security

  As a system administrator
  I want to implement Wireless Security
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:network-wifi-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:network-wifi-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
