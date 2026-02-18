# Intent: Allow developers to understand diagram structure without full decoding by viewing metadata about DrawIO diagrams
# Architecture: Affects eac-cli drawio info command; invokes drawio-oci Docker image; reads .drawio.png file; displays diagram count and page names; depends on Docker availability

@L2 @deps:docker @deps:go @ov @env:isolated-test-project
Feature: eac-cli_drawio-info

  As a developer using EAC
  I want to view metadata about DrawIO diagrams
  So that I can understand diagram structure without full decoding

  Background:
    Given I am in a git repository
    And the drawio-oci Docker image is available

  Rule: Shows diagram metadata

    Scenario: Show info for valid .drawio.png
      Given a valid .drawio.png file "diagram.drawio.png"
      When I run the command "drawio info diagram.drawio.png"
      Then the exit code is 0
      And stdout contains "Diagrams"

    Scenario: Show page name in info
      Given a .drawio.png file "named.drawio.png" with page name "MyDiagram"
      When I run the command "drawio info named.drawio.png"
      Then the exit code is 0
      And stdout contains "MyDiagram"

  Rule: Handles error cases gracefully

    Scenario: Error on non-existent file
      When I run the command "drawio info nonexistent.drawio.png"
      Then the exit code is 1
      And stderr contains "not found"

    Scenario: Error on missing input file
      When I run the command "drawio info"
      Then the exit code is 1
      And stderr contains "required"

  Rule: Command accessibility

    Scenario: Help is available
      When I run the command "drawio info --help"
      Then the exit code is 0
      And stdout contains "info"
      And stdout contains "drawio"
