@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_sast

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    Scenario: Run sast scan with mocked Docker  
      When I run the command "scan eac-core --scanner sast"
      Then the exit code is 0
      And evidence files should exist in directory "out/scan/eac-core/go/"
