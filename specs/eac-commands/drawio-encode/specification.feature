@L2 @deps:docker @deps:go @ov @env:isolated-test-project
Feature: eac-commands_drawio-encode

  As a developer using EAC
  I want to encode human-readable XML to DrawIO format
  So that I can prepare content for embedding in PNG files

  Background:
    Given I am in a git repository
    And the drawio-cli Docker image is available

  Rule: Encodes readable XML to DrawIO format

    Scenario: Encode valid mxGraphModel XML to stdout
      Given a file "diagram.xml" with valid mxGraphModel content
      When I run the command "drawio encode --input diagram.xml"
      Then the exit code is 0
      And stdout contains "<mxfile"

    Scenario: Encode to output file
      Given a file "diagram.xml" with valid mxGraphModel content
      When I run the command "drawio encode --input diagram.xml --output encoded.xml"
      Then the exit code is 0
      And the file "encoded.xml" should exist
      And the file "encoded.xml" should contain "<mxfile"

  Rule: Handles error cases gracefully

    Scenario: Error on empty stdin input
      When I run the command "drawio encode"
      Then the exit code is 1
      And stderr contains "Invalid" or "XML"

    Scenario: Error on non-existent file
      When I run the command "drawio encode --input nonexistent.xml"
      Then the exit code is 1
      And stderr contains "not found" or "does not exist"

  Rule: Command accessibility

    Scenario: Help is available
      When I run the command "drawio encode --help"
      Then the exit code is 0
      And stdout contains "encode"
