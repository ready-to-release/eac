# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

@risk-control:cloud-logging
Feature: Cloud Audit Logging

  As a system administrator
  I want to implement Cloud Audit Logging
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

    @risk-control:cloud-logging-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @hipaa @csa-ccm @iso27017 @iso27018
# @implementation:required @automation:full

    @risk-control:cloud-logging-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
