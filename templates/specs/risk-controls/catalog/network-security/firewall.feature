# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @hipaa @nist-800-53
# @implementation:required @automation:full

@risk-control:network-firewall
Feature: Firewall Configuration

  As a system administrator
  I want to implement Firewall Configuration
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:network-firewall-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:network-firewall-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
