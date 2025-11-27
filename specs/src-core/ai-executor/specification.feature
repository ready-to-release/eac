@skip:deprecated
Feature: src-core_ai-executor
  As a CLI command
  I want to execute AI requests through a unified interface
  So that I can use different AI providers transparently

  # NOTE: This feature is redundant with Go tests in src/core/ai/executor_test.go
  # src/core is a library package and should be tested with Go unit tests, not Godog specs

  Rule: Executor loads provider from configuration

    The executor reads agent-config.yml to determine which provider to use.
    If the config exists and is valid, the executor uses the specified provider.

    @L2 @ov
    Scenario: Execute with claude-cli configured
      Given agent config exists with provider "claude-cli"
      When I execute a prompt "Hello"
      Then the executor uses claude-cli provider
      And no API key is required

  Rule: Executor requires valid configuration or fails with clear instructions

    When no agent configuration file exists or configuration is invalid,
    the executor MUST throw an error with clear instructions to run r2r init.
    There is NO automatic fallback behavior.

    @L2 @ov
    Scenario: Execute with no config file
      Given no agent config file exists
      When I execute a prompt "Hello"
      Then an error is returned
      And the error message contains "agent-config.yml not found"
      And the error message contains "run: r2r agent init"

    @L2 @ov
    Scenario: Execute with malformed config file
      Given agent config file is malformed
      When I execute a prompt "Hello"
      Then an error is returned
      And the error message contains "failed to parse agent-config.yml"
      And the error message contains "run: r2r agent init"

    @L2 @ov
    Scenario: Execute with invalid provider name
      Given agent config exists with provider "invalid-provider"
      When I execute a prompt "Hello"
      Then an error is returned
      And the error message contains "unknown provider: invalid-provider"
      And the error message contains "run: r2r agent init"

  Rule: Executor logs interactions based on debug parameter

    The executor supports a debug parameter that controls logging behavior.
    When debug is true, logs are included in the prompt output.
    When debug is false, no logs are recorded or stored anywhere.
    There is NO logging to .r2r directory.

    @L2 @ov
    Scenario: Execute with debug enabled
      Given agent config exists with provider "claude-cli"
      And debug parameter is true
      When I execute a prompt "Hello"
      Then the execution succeeds
      And debug logs are included in the output
      And the debug logs contain timestamp
      And the debug logs contain provider "claude-cli"
      And the debug logs contain success status

    @L2 @ov
    Scenario: Execute with debug disabled
      Given agent config exists with provider "claude-cli"
      And debug parameter is false
      When I execute a prompt "Hello"
      Then the execution succeeds
      And no debug logs are included in the output
      And no log files are created

    @L2 @ov
    Scenario: Execute with debug disabled by default
      Given agent config exists with provider "claude-cli"
      And no debug parameter is specified
      When I execute a prompt "Hello"
      Then the execution succeeds
      And no debug logs are included in the output
      And no log files are created

    @L2 @ov
    Scenario: Failed execution with debug enabled
      Given agent config exists with provider "claude-cli"
      And debug parameter is true
      And the provider will fail
      When I execute a prompt "Hello"
      Then the execution fails
      And debug logs are included in the output
      And the debug logs contain failure status
      And the debug logs contain error message

  Rule: Executor supports functional options

    The executor accepts functional options to override default behavior,
    such as specifying a different model or temperature.

    @L2 @ov
    Scenario: Execute with custom model option
      Given agent config exists with provider "claude-cli"
      When I execute a prompt "Hello" with model "opus"
      Then the executor uses the specified model

    @L2 @ov
    Scenario: Execute with custom temperature option
      Given agent config exists with provider "claude-cli"
      When I execute a prompt "Hello" with temperature 0.7
      Then the executor uses the specified temperature

  Rule: Default configuration is included in repository

    The repository includes a default .r2r/agent-config.yml file
    with claude-cli as the only configured provider.

    @L1 @ov
    Scenario: Repository contains default agent-config.yml
      Given a fresh clone of the repository
      Then agent-config.yml exists in the .r2r directory
      And the config specifies provider "claude-cli"
      And the config does not require API keys
      And the config is safe to commit to version control
