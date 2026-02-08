@L2 @deps:docker @deps:go @ov @env:isolated-test-project
Feature: eac-cli_drawio-embed

  As a developer using EAC
  I want to embed encoded XML into PNG files
  So that I can update diagram content while preserving the PNG format

  Background:
    Given I am in a git repository
    And the drawio-oci Docker image is available

  Rule: Embeds XML into existing PNG files

    Scenario: Embed XML into existing .drawio.png (in-place)
      Given a valid .drawio.png file "diagram.drawio.png"
      And an encoded XML file "new-content.xml" with different content
      When I run the command "drawio embed --png diagram.drawio.png --xml new-content.xml"
      Then the exit code is 0
      And decoding "diagram.drawio.png" should show the new content

    Scenario: Embed with explicit output path
      Given a valid .drawio.png file "source.drawio.png"
      And an encoded XML file "content.xml"
      When I run the command "drawio embed --png source.drawio.png --xml content.xml --output result.drawio.png"
      Then the exit code is 0
      And the file "result.drawio.png" should exist

  Rule: Creates PNG if it does not exist

    Scenario: Creates PNG when target does not exist
      Given an encoded XML file "content.xml"
      When I run the command "drawio embed --png new-diagram.drawio.png --xml content.xml"
      Then the exit code is 0
      And the file "new-diagram.drawio.png" should exist

  Rule: Handles error cases gracefully

    Scenario: Error when XML does not exist
      Given a valid .drawio.png file "diagram.drawio.png"
      When I run the command "drawio embed --png diagram.drawio.png --xml nonexistent.xml"
      Then the exit code is 1
      And stderr contains "not found" or "does not exist"

  Rule: Command accessibility

    Scenario: Help is available
      When I run the command "drawio embed --help"
      Then the exit code is 0
      And stdout contains "embed"
      And stdout contains "png"
