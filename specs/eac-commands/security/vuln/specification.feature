@L2 @ov @deps:go @deps:docker @env:isolated-test-project @skip:broken
Feature: security_vuln

  As a security auditor
  I want to scan modules for known vulnerabilities
  So that I can identify and remediate security risks

  Rule: Vulnerability scanner creates evidence file with integrity verification

    @control:si-2 @control:si-3
    Scenario: Scan module for vulnerabilities with mocked tool
      When I run the command "scan vuln eac-core"
      Then the exit code is 0
      And evidence files should exist in directory "out/scan/eac-core/go/"
      And the latest evidence file should have JSON field "module" with value "eac-core"
      And the latest evidence file should have JSON field "scanner" with value "vuln"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    @control:si-2
    Scenario: Vulnerability scan with severity filter
      When I run the command "scan vuln eac-core --severity HIGH,CRITICAL"
      Then the exit code is 0
      And evidence files should exist in directory "out/scan/eac-core/go/"

    Scenario: Vulnerability scan with debug logging
      When I run the command "scan vuln eac-core --debug"
      Then the exit code is 0
      And a log file should exist in directory "out/logs/security/"

    Scenario: Vulnerability scan with invalid module
      When I run the command "scan vuln nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: Vulnerability scan help is accessible
      When I run the command "scan vuln --help"
      Then the exit code is 0
      And I should see "vuln" or "vulnerability" or "Usage"
