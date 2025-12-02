@ov @deps:go @deps:docker @env:isolated-test-project @skip:broken
Feature: security_iac

  As a security auditor
  I want to scan Infrastructure as Code files for misconfigurations
  So that I can identify security issues before deployment

  Rule: IaC scanner creates evidence file with integrity verification

    Scenario: Scan IaC files with mocked tool
      When I run the command "security iac eac-core"
      Then the exit code is 0 or 1
      And evidence files should exist in directory "out/security/eac-core/iac/"
      And the latest evidence file should have JSON field "module" with value "eac-core"
      And the latest evidence file should have JSON field "scanner" with value "iac"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: IaC scan with debug logging
      When I run the command "security iac eac-core --debug"
      Then the exit code is 0 or 1
      And a log file should exist in directory "out/logs/security/"

    Scenario: IaC scan with invalid module
      When I run the command "security iac nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: IaC scan for all modules
      When I run the command "security iac"
      Then the exit code is 0 or 1
      And I should see "modules" or "Scanning"

    Scenario: IaC scan help is accessible
      When I run the command "security iac --help"
      Then the exit code is 0
      And I should see "iac" or "Infrastructure" or "Usage"
