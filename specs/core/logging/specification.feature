@deps:go @L2 @ov @control:au-2 @control:au-3 @control:au-12
Feature: core_logging
  As a developer
  I want logs routed to console and/or file based on debug mode
  So that I can see important messages on console and persist all logs when debugging

  Background:
    Given a logging module configured for "commit" command

  Rule: No file logging by default

    In default mode (debug disabled), no logs are written to file.
    This keeps the filesystem clean during normal operation.

    Scenario: No log file created by default
      Given debug mode is disabled
      When I log an Info message "Commit message generated"
      Then the message appears on console stdout
      And no log file is created

    Scenario: Debug messages not logged anywhere by default
      Given debug mode is disabled
      When I log a Debug message "Processing staged files"
      Then the message does not appear on console
      And no log file is created

  Rule: Info, Warn, and Error always appear on console

    Higher-level logs are always visible to the user on console
    regardless of debug mode.

    Scenario: Info logs appear on console
      Given debug mode is disabled
      When I log an Info message "Commit message generated"
      Then the message appears on console stdout

    Scenario: Warn logs appear on console
      Given debug mode is disabled
      When I log a Warn message "Large diff detected"
      Then the message appears on console stdout

    Scenario: Error logs appear on console
      Given debug mode is disabled
      When I log an Error message "Git repository not found"
      Then the message appears on console stderr

  Rule: Debug flag enables file logging to unified commands.log

    When the --debug flag is enabled, all log levels are written to
    the unified out/commands.log file for troubleshooting and audit.

    Scenario: All levels write to unified log when debug enabled
      Given debug mode is enabled
      When I log messages at all levels
      Then all messages appear in file "out/commands.log"

    Scenario: Debug logs write to unified log when debug enabled
      Given debug mode is enabled
      When I log a Debug message "Processing staged files"
      Then the message appears in file "out/commands.log"

    Scenario: Info logs write to unified log when debug enabled
      Given debug mode is enabled
      When I log an Info message "Commit message generated"
      Then the message appears in file "out/commands.log"

  Rule: Debug flag enables Debug output on console

    When the --debug flag is enabled, Debug messages also
    appear on console for interactive troubleshooting.

    Scenario: Debug logs shown on console when debug enabled
      Given debug mode is enabled
      When I log a Debug message "Processing staged files"
      Then the message appears on console stdout
      And the message appears in file "out/commands.log"

    Scenario: All levels shown on console when debug enabled
      Given debug mode is enabled
      When I log messages at all levels
      Then all messages appear on console
      And all messages appear in the log file

  Rule: Logger uses single Zap instance with multiple cores

    The logger must use a single Zap logger instance constructed from
    multiple cores (Tee) for efficient dual-output when debug is enabled.

    Scenario: Single logger writes to multiple outputs when debug enabled
      Given debug mode is enabled
      When I log an Info message "Test message"
      Then only one Zap logger instance exists
      And both console and file receive the message

  Rule: All commands use unified log file

    All commands write to a single unified out/commands.log file.
    This simplifies log management and debugging.

    Scenario: Log file created at unified path when debug enabled
      Given debug mode is enabled
      And a logger configured for module "commit"
      When I log an Info message "Test"
      Then the log file exists at "out/commands.log"

    Scenario: Multiple loggers write to same unified log
      Given debug mode is enabled
      And a logger configured for module "commit"
      And a logger configured for module "build"
      When I log Info messages to both loggers
      Then "out/commands.log" contains commit logs
      And "out/commands.log" contains build logs
