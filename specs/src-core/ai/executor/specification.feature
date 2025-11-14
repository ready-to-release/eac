Feature: src-core-ai_executor
  As a CLI command
  I want to execute AI requests through a unified interface
  So that I can use different AI providers transparently

  Rule: Executor loads provider from configuration

    The executor reads .r2r/agent-config.yml to determine which provider to use.
    If the config exists and is valid, the executor uses the specified provider.

    @L2 @ov
    Scenario: Execute with claude-api configured
      Given agent config exists with provider "claude-api"
      And environment variable "ANTHROPIC_API_KEY" is set
      When I execute a prompt "Hello"
      Then the executor uses claude-api provider
      And the API key from environment is used

    @L2 @ov
    Scenario: Execute with claude-cli configured
      Given agent config exists with provider "claude-cli"
      When I execute a prompt "Hello"
      Then the executor uses claude-cli provider
      And no API key is required

    @L2 @ov
    Scenario: Execute with openai configured
      Given agent config exists with provider "openai"
      And environment variable "OPENAI_API_KEY" is set
      When I execute a prompt "Hello"
      Then the executor uses openai provider
      And the API key from environment is used

  Rule: Executor falls back to claude-cli when no config exists

    When no agent configuration file exists, the executor automatically falls back
    to the claude-cli provider, which uses Claude Pro subscription authentication.
    This provides zero-configuration usage for all users.

    @L2 @ov
    Scenario: Execute with no config file
      Given no agent config file exists
      When I execute a prompt "Hello"
      Then the executor uses claude-cli provider
      And no API key is required

    @L2 @ov
    Scenario: Execute with malformed config file
      Given agent config file is malformed
      When I execute a prompt "Hello"
      Then the executor uses claude-cli provider
      And a warning is logged about config error

  Rule: Executor validates provider before execution

    The executor validates that the selected provider is properly configured
    before attempting execution. If validation fails, it falls back to claude-cli
    or returns a clear error.

    @L2 @ov @negative
    Scenario: Execute with missing API key for claude-api
      Given agent config exists with provider "claude-api"
      And environment variable "ANTHROPIC_API_KEY" is not set
      When I execute a prompt "Hello"
      Then the executor falls back to claude-cli provider
      And a warning is logged about missing API key

    @L2 @ov @negative
    Scenario: Execute with invalid provider name
      Given agent config exists with provider "invalid-provider"
      When I execute a prompt "Hello"
      Then the executor falls back to claude-cli provider
      And a warning is logged about invalid provider

  Rule: Executor logs all AI interactions

    All AI executions are logged to .r2r/logs/ai-executions.jsonl for debugging
    and audit purposes. Each log entry includes timestamp, provider, prompt,
    response, duration, and success/failure status.

    @L2 @ov
    Scenario: Successful execution is logged
      Given agent config exists with provider "claude-cli"
      When I execute a prompt "Hello"
      And the execution succeeds
      Then a log entry is written to .r2r/logs/ai-executions.jsonl
      And the log entry contains timestamp
      And the log entry contains provider "claude-cli"
      And the log entry contains success status

    @L2 @ov @negative
    Scenario: Failed execution is logged
      Given agent config exists with provider "claude-api"
      And environment variable "ANTHROPIC_API_KEY" is set to invalid value
      When I execute a prompt "Hello"
      And the execution fails
      Then a log entry is written to .r2r/logs/ai-executions.jsonl
      And the log entry contains failure status
      And the log entry contains error message

  Rule: Executor supports functional options

    The executor accepts functional options to override default behavior,
    such as specifying a different model or temperature.

    @L2 @ov
    Scenario: Execute with custom model option
      Given agent config exists with provider "claude-api"
      And environment variable "ANTHROPIC_API_KEY" is set
      When I execute a prompt "Hello" with model "claude-3-opus-20240229"
      Then the executor uses the specified model
      And the log entry contains model "claude-3-opus-20240229"

    @L2 @ov
    Scenario: Execute with custom temperature option
      Given agent config exists with provider "claude-api"
      And environment variable "ANTHROPIC_API_KEY" is set
      When I execute a prompt "Hello" with temperature 0.7
      Then the executor uses the specified temperature
      And the log entry contains temperature 0.7
