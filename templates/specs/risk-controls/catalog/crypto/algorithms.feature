# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @fips-140-2 @fips-140-3 @nist-800-53
# @implementation:required @automation:full

@risk-control:crypto-algorithms
Feature: Approved Cryptographic Algorithms

  As a system administrator
  I want to implement Approved Cryptographic Algorithms
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @fips-140-2 @fips-140-3 @nist-800-53
# @implementation:required @automation:full

    @risk-control:crypto-algorithms-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @fips-140-2 @fips-140-3 @nist-800-53
# @implementation:required @automation:full

    @risk-control:crypto-algorithms-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
