@ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_sast

  As a security auditor
  I want to perform static application security testing on source code
  So that I can identify security vulnerabilities and code quality issues

  Rule: SAST scanner creates evidence file with integrity verification

    Scenario: Perform SAST scan with mocked tool
      When I run the command "security sast eac-core"
      Then the exit code is 0
      And evidence files should exist in directory "out/security/eac-core/sast/"
      And the latest evidence file should have JSON field "module" with value "eac-core"
      And the latest evidence file should have JSON field "scanner" with value "sast"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: SAST scan with specific ruleset
      When I run the command "security sast eac-core --config p/security-audit"
      Then the exit code is 0
      And evidence files should exist in directory "out/security/eac-core/sast/"

    Scenario: SAST scan with debug logging
      When I run the command "security sast eac-core --debug"
      Then the exit code is 0
      And a log file should exist in directory "out/logs/security/"

    Scenario: SAST scan with invalid module
      When I run the command "security sast nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: SAST help is accessible
      When I run the command "security sast --help"
      Then the exit code is 0
      And I should see "sast" or "static" or "Usage"
