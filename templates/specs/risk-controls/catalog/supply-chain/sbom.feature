# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53
# @implementation:required @automation:full

@risk-control:supply-sbom
Feature: Software Bill of Materials

  As a system administrator
  I want to implement Software Bill of Materials
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:supply-sbom-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:supply-sbom-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
