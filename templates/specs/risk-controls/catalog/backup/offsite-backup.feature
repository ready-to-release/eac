# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @nist-800-53
# @implementation:required @automation:full

@risk-control:backup-offsite
Feature: Offsite Backup

  As a system administrator
  I want to implement Offsite Backup
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:backup-offsite-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:backup-offsite-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
