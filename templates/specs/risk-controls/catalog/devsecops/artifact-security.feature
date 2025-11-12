# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

@risk-control:devsecops-artifacts
Feature: Artifact Security

  As a system administrator
  I want to implement Artifact Security
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

    @risk-control:devsecops-artifacts-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

    @risk-control:devsecops-artifacts-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
