# @industry:MEDDEV @industry:HEALTH @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @hipaa @iso13485 @iec62304 @mdr @ivdr @fda-premarket @imdrf
# @implementation:required @automation:full

@risk-control:iot-medical-device
Feature: Medical Device Security

  As a system administrator
  I want to implement Medical Device Security
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:MEDDEV @industry:HEALTH @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @hipaa @iso13485 @iec62304 @mdr @ivdr @fda-premarket @imdrf
# @implementation:required @automation:full

    @risk-control:iot-medical-device-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:MEDDEV @industry:HEALTH @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @hipaa @iso13485 @iec62304 @mdr @ivdr @fda-premarket @imdrf
# @implementation:required @automation:full

    @risk-control:iot-medical-device-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
