# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @nist-800-53
# @implementation:required @automation:full

@risk-control:threat-analysis
Feature: Threat Analysis

  As a system administrator
  I want to implement Threat Analysis
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:threat-analysis-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:threat-analysis-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
