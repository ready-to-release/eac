# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @pci-dss
# @implementation:required @automation:full

@risk-control:audit-time-sync
Feature: Time Synchronization

  As a system administrator
  I want to implement Time Synchronization
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:audit-time-sync-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:audit-time-sync-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
