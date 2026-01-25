@L2 @deps:docker @deps:go @ov @env:isolated-test-project
Feature: eac-commands_drawio-decode

  As a developer using EAC
  I want to decode DrawIO diagrams to human-readable XML
  So that I can view and edit diagram content programmatically

  Background:
    Given I am in a git repository
    And the drawio-cli Docker image is available

  Rule: Decodes .drawio.png files to readable XML

    Scenario: Decode valid .drawio.png file to stdout
      Given a valid .drawio.png file "test-decode.drawio.png"
      When I run the command "drawio decode --input test-decode.drawio.png"
      Then the exit code is 0
      And stdout contains "<mxGraphModel"
      And stdout contains "<mxCell"

    Scenario: Decode to output file
      Given a valid .drawio.png file "test-decode.drawio.png"
      When I run the command "drawio decode --input test-decode.drawio.png --output decoded.xml"
      Then the exit code is 0
      And the file "decoded.xml" should exist
      And the file "decoded.xml" should contain "<mxGraphModel"

  Rule: Handles error cases gracefully

    Scenario: Error on missing input flag
      When I run the command "drawio decode"
      Then the exit code is 1
      And stderr contains "input" or "required"

    Scenario: Error on non-existent file
      When I run the command "drawio decode --input nonexistent.drawio.png"
      Then the exit code is 1
      And stderr contains "not found" or "does not exist"

  Rule: Command accessibility

    Scenario: Help is available
      When I run the command "drawio decode --help"
      Then the exit code is 0
      And stdout contains "decode"
      And stdout contains "input"
