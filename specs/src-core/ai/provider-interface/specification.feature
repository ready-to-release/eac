Feature: src-core-ai_provider-interface
  As a developer
  I want a consistent interface for all AI providers
  So that I can switch between providers without changing command code

  Background:
    Given the AI provider abstraction layer exists

  Rule: All providers must implement Execute()
    The Execute() method is the core interface that all providers must implement.
    It takes a string input and returns a string output with optional parameters.

    @L0 @ov
    Scenario: Claude API provider executes successfully
      Given a valid claude-api provider configuration
      And ANTHROPIC_API_KEY environment variable is set
      When I execute a prompt "Say hello"
      Then the provider returns a non-empty response
      And no error is returned

    @L0 @ov
    Scenario: Claude CLI provider executes successfully
      Given a valid claude-cli provider configuration
      And the claude CLI tool is available
      When I execute a prompt "Say hello"
      Then the provider returns a non-empty response
      And no error is returned

    @L0 @ov
    Scenario: Provider execution with model option
      Given a valid claude-api provider configuration
      When I execute a prompt "Test" with model "claude-3-haiku-20240307"
      Then the provider uses the specified model
      And returns a non-empty response

  Rule: Providers validate configuration before use
    Providers must check their configuration and fail fast with clear errors
    if the configuration is invalid or required credentials are missing.

    @L0 @ov
    Scenario: Provider fails fast when API key is missing
      Given a claude-api provider configuration
      And ANTHROPIC_API_KEY environment variable is not set
      When I attempt to execute a prompt
      Then an error is returned
      And the error message indicates the missing API key

    @L0 @ov
    Scenario: Provider fails fast when configuration is invalid
      Given an invalid provider configuration with empty name
      When I attempt to create a provider
      Then an error is returned
      And the error message indicates the invalid configuration

  Rule: Providers return clear errors on failure
    When provider execution fails, errors must be wrapped with context
    to help diagnose the issue.

    @L0 @ov
    Scenario: Provider wraps API errors with context
      Given a claude-api provider configuration
      And the API endpoint is unreachable
      When I execute a prompt
      Then an error is returned
      And the error message includes the original API error
      And the error message includes provider context

  Rule: All standard providers are supported
    The system supports multiple AI providers including Claude, OpenAI, and Gemini.
    Each provider implements the same interface for consistent usage.

    @L0 @ov @dep:openai
    Scenario: OpenAI provider executes successfully
      Given a valid openai provider configuration
      And OPENAI_API_KEY environment variable is set
      When I execute a prompt "Say hello"
      Then the provider returns a non-empty response
      And no error is returned

    @L0 @ov @dep:openai
    Scenario: OpenAI provider fails when API key is missing
      Given an openai provider configuration
      And OPENAI_API_KEY environment variable is not set
      When I attempt to create a provider
      Then an error is returned
      And the error message indicates the missing API key

    @L0 @ov @dep:gemini
    Scenario: Gemini provider executes successfully
      Given a valid gemini provider configuration
      And GOOGLE_API_KEY environment variable is set
      When I execute a prompt "Say hello"
      Then the provider returns a non-empty response
      And no error is returned

    @L0 @ov @dep:gemini
    Scenario: Gemini provider fails when API key is missing
      Given a gemini provider configuration
      And GOOGLE_API_KEY environment variable is not set
      When I attempt to create a provider
      Then an error is returned
      And the error message indicates the missing API key
