# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

@risk-control:cloud-backup
Feature: Cloud Backup Security

  As a system administrator
  I want to implement Cloud Backup Security
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

    @risk-control:cloud-backup-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

    @risk-control:cloud-backup-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
