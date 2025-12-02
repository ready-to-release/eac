@ov @deps:go @deps:docker @env:isolated-test-project @skip:broken
Feature: security_compliance

  As a security auditor
  I want to check modules for compliance with security standards
  So that I can verify adherence to CIS benchmarks and industry standards

  Rule: Compliance scanner creates evidence file with integrity verification

    Scenario: Check module compliance with mocked tool
      When I run the command "security compliance eac-core --compliance k8s-cis"
      Then the exit code is 0 or 1
      And evidence files should exist in directory "out/security/eac-core/compliance/"
      And the latest evidence file should have JSON field "module" with value "eac-core"
      And the latest evidence file should have JSON field "scanner" with value "compliance"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: Compliance check with debug logging
      When I run the command "security compliance eac-core --debug"
      Then the exit code is 0 or 1
      And a log file should exist in directory "out/logs/security/"

    Scenario: Compliance check with invalid module
      When I run the command "security compliance nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: Compliance check for all modules
      When I run the command "security compliance --compliance docker-cis"
      Then the exit code is 0 or 1
      And I should see "modules" or "Scanning"

    Scenario: Compliance help is accessible
      When I run the command "security compliance --help"
      Then the exit code is 0
      And I should see "compliance" or "CIS" or "Usage"
