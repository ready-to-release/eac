# Intent: Allow developers to start new diagrams programmatically by creating new DrawIO diagram files with EAC visual defaults
# Architecture: Affects eac-cli drawio create command; invokes drawio-oci Docker image; outputs .drawio.png files with embedded mxGraphModel XML; applies default gray background and shadow settings; supports custom XML and name

@L2 @deps:docker @deps:go @ov @env:isolated-test-project
Feature: eac-cli_drawio-create

  As a developer using EAC
  I want to create new DrawIO diagram files
  So that I can start new diagrams programmatically

  Background:
    Given I am in a git repository
    And the drawio-oci Docker image is available

  Rule: Creates new .drawio.png files

    Scenario: Create blank diagram
      When I run the command "drawio create --output new-diagram.drawio.png"
      Then the exit code is 0
      And the file "new-diagram.drawio.png" should exist
      And the file "new-diagram.drawio.png" should be a valid PNG
      And decoding "new-diagram.drawio.png" should contain "mxGraphModel"

    Scenario: Create diagram with custom name
      When I run the command "drawio create --output named.drawio.png --name Architecture"
      Then the exit code is 0
      And decoding "named.drawio.png" should contain "Architecture"

    Scenario: Create diagram with custom XML
      Given a file "custom.xml" with valid mxGraphModel content containing "CustomShape"
      When I run the command "drawio create --output custom.drawio.png --xml custom.xml"
      Then the exit code is 0
      And decoding "custom.drawio.png" should contain "CustomShape"

  Rule: Uses EAC visual defaults

    Scenario: New diagram has gray background
      When I run the command "drawio create --output styled.drawio.png"
      Then the exit code is 0
      And decoding "styled.drawio.png" should contain "background"
      And decoding "styled.drawio.png" should contain "#CFCFCF"

    Scenario: New diagram has shadow enabled
      When I run the command "drawio create --output shadow.drawio.png"
      Then the exit code is 0
      And decoding "shadow.drawio.png" should contain 'shadow="1"'

  Rule: Handles error cases gracefully

    Scenario: Error when output not specified
      When I run the command "drawio create"
      Then the exit code is 1
      And stderr contains "output" or "required"

    Scenario: Error when XML file does not exist
      When I run the command "drawio create --output out.drawio.png --xml nonexistent.xml"
      Then the exit code is 1
      And stderr contains "not found" or "does not exist"

  Rule: Command accessibility

    Scenario: Help is available
      When I run the command "drawio create --help"
      Then the exit code is 0
      And stdout contains "create"
      And stdout contains "output"
