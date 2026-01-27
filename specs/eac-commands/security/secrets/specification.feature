@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_secrets

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    Scenario: Run secrets scan with mocked Docker  
      When I run the command "scan eac-core --scanner secrets"
      Then the exit code is 0
      And evidence files should exist in directory "out/scan/eac-core/go/"
