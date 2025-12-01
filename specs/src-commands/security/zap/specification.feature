@ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_zap

  As a security auditor
  I want to perform dynamic application security testing on running applications
  So that I can identify runtime vulnerabilities and security misconfigurations

  Rule: ZAP scanner creates evidence file with integrity verification

    Scenario: Perform DAST scan with mocked tool
      When I run the command "security zap src-core --target http://localhost:8080"
      Then the exit code is 0
      And evidence files should exist in directory "out/security/src-core/zap/"
      And the latest evidence file should have JSON field "module" with value "src-core"
      And the latest evidence file should have JSON field "scanner" with value "zap"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: ZAP scan with full scan type
      When I run the command "security zap src-core --target http://localhost:8080 --scan-type full"
      Then the exit code is 0
      And evidence files should exist in directory "out/security/src-core/zap/"

    Scenario: ZAP scan with debug logging
      When I run the command "security zap src-core --target http://localhost:8080 --debug"
      Then the exit code is 0
      And a log file should exist in directory "out/logs/security/"

    Scenario: ZAP scan with missing target URL
      When I run the command "security zap src-api"
      Then the exit code is 1
      And I should see "target" or "required" or "Error"

    Scenario: ZAP help is accessible
      When I run the command "security zap --help"
      Then the exit code is 0
      And I should see "zap" or "dynamic" or "Usage"
