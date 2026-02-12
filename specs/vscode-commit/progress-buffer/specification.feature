@ov @depm:vscode-commit
Feature: vscode-commit_progress-buffer

  As a developer
  I want the progress frame buffer to manage display updates correctly
  So that commit progress is shown smoothly without flickering

  Background:
    Given a new progress frame buffer

  Rule: Buffer initialization and empty state handling

    @L0 @ov
    Scenario: Empty buffer returns fallback text
      When I check the current frame
      Then the frame should be a non-empty string
      And the buffer size should be 0

  Rule: Progress messages are filtered and normalized

    @L0 @ov
    Scenario: Valid progress messages are added to buffer
      When I add progress "Loading data"
      Then the buffer size should be 1
      And the current frame should be "Loading data"

    @L0 @ov
    Scenario: Empty messages are rejected
      When I add progress ""
      Then the buffer size should be 0

    @L0 @ov
    Scenario: Messages with only special characters are rejected
      When I add progress "---"
      And I add progress "..."
      Then the buffer size should be 0

    @L0 @ov
    Scenario: Very short messages are rejected
      When I add progress "ab"
      Then the buffer size should be 0

    @L1 @ov
    Scenario: Whitespace is normalized in messages
      When I add progress "  Multiple   spaces  here  "
      Then the current frame should be "Multiple spaces here"

    @L1 @ov
    Scenario: Long messages are truncated
      When I add progress with 50 characters
      Then the current frame length should be at most 40

  Rule: Buffer respects maximum capacity

    @L1 @ov
    Scenario: Buffer drops oldest messages when full
      Given a progress buffer with max 3 frames
      When I add progress "Message 1"
      And I add progress "Message 2"
      And I add progress "Message 3"
      And I add progress "Message 4"
      Then the buffer size should be 3
      And the current frame should be "Message 4"

  Rule: Priority events override progress messages

    @L1 @ov
    Scenario: Event takes priority over progress
      When I add progress "Progress message"
      And I push event "Priority event"
      Then the current frame should be "Priority event"
      And the buffer should have a priority frame

    @L1 @ov
    Scenario: Clearing event falls back to progress
      When I add progress "Background progress"
      And I push event "Priority event"
      And I clear the event
      Then the current frame should be "Background progress"
      And the buffer should not have a priority frame

  Rule: Buffer can be completely cleared

    @L1 @ov
    Scenario: Clear removes all frames and priority
      When I add progress "Some progress"
      And I push event "Some event"
      And I clear the buffer
      Then the buffer size should be 0
      And the buffer should not have a priority frame
