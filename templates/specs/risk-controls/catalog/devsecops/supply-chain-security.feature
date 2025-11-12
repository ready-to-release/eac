# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

@risk-control:devsecops-supply-chain
Feature: DevSecOps Supply Chain

  As a system administrator
  I want to implement DevSecOps Supply Chain
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

    @risk-control:devsecops-supply-chain-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @nist-800-53 @owasp @nist-ssdf @slsa
# @implementation:required @automation:full

    @risk-control:devsecops-supply-chain-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
