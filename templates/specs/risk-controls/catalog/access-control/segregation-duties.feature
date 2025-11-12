# @industry:PHARMA @industry:FINANCE @industry:MEDDEV @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @sox @pci-dss @fda-21cfr11
# @implementation:required @automation:partial

@risk-control:access-sod
Feature: Segregation of Duties

  As a system administrator
  I want to implement Segregation of Duties
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:access-sod-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:access-sod-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
