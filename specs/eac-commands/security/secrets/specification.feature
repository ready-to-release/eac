@ov @deps:go @deps:docker @env:isolated-test-project @skip:broken @control:sc-28 @control:ia-5
Feature: security_secrets

  As a security auditor
  I want to detect hardcoded secrets and credentials in source code
  So that I can prevent accidental exposure of sensitive data

  Rule: Secrets scanner creates evidence file with integrity verification

    Scenario: Scan module for secrets with mocked tool
      When I run the command "scan secrets eac-core"
      Then the exit code is 0 or 1
      And evidence files should exist in directory "out/scan/eac-core/secrets/"
      And the latest evidence file should have JSON field "module" with value "eac-core"
      And the latest evidence file should have JSON field "scanner" with value "secrets"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: Secrets scan with debug logging
      When I run the command "scan secrets eac-core --debug"
      Then the exit code is 0 or 1
      And a log file should exist in directory "out/logs/security/"

    Scenario: Secrets scan with invalid module
      When I run the command "scan secrets nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: Secrets scan for all modules
      When I run the command "scan secrets"
      Then the exit code is 0 or 1
      And I should see "modules" or "Scanning"

    Scenario: Secrets scan help is accessible
      When I run the command "scan secrets --help"
      Then the exit code is 0
      And I should see "secrets" or "credentials" or "Usage"
