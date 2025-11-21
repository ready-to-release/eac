@skip:redundant
Feature: src-core_ai-config-loading
  As a developer
  I want to load AI provider configuration from agent-config.yml
  So that the project can specify which AI provider to use

  # NOTE: This feature is redundant with Go tests in src/core/ai/config_loader_test.go
  # src/core is a library package and should be tested with Go unit tests, not Godog specs

  Background:
    Given a repository with agent-config.yml

  Rule: Configuration loads from agent-config.yml in .r2r directory

    The system must read configuration from .r2r/agent-config.yml.
    This file is committed to version control and provides default configuration.

    @L2 @ov
    Scenario: Load valid configuration file
      Given a valid agent-config.yml file with claude-cli provider
      When I load the configuration
      Then the configuration is loaded successfully
      And the provider name is "claude-cli"
      And the model is "claude-3-haiku-20240307"

    @L2 @ov
    Scenario: Configuration file does not exist
      Given no agent-config.yml file exists
      When I attempt to load the configuration
      Then an error is returned
      And the error indicates "agent-config.yml not found"
      And the error suggests "run: r2r agent init"

    @L2 @ov
    Scenario: Configuration file contains malformed YAML
      Given an agent-config.yml file with invalid YAML syntax
      When I attempt to load the configuration
      Then an error is returned
      And the error indicates "failed to parse agent-config.yml"
      And the error suggests "run: r2r agent init"

  Rule: Environment variables are substituted at runtime

    The configuration file should use ${VAR_NAME} syntax for secrets,
    and the system must substitute these with actual environment variable values.
    This is only needed for providers that require API keys (not claude-cli).

    @L2 @ov
    Scenario: Substitute environment variable in API key
      Given an agent-config.yml file with api_key: ${ANTHROPIC_API_KEY}
      And ANTHROPIC_API_KEY environment variable is set to "sk-test-123"
      When I load the configuration
      Then the API key in the loaded config is "sk-test-123"

    @L2 @ov
    Scenario: Missing environment variable results in empty string
      Given an agent-config.yml file with api_key: ${MISSING_VAR}
      And MISSING_VAR environment variable is not set
      When I load the configuration
      Then the API key in the loaded config is empty

    @L2 @ov
    Scenario: Multiple environment variables are substituted
      Given an agent-config.yml file with multiple ${VAR} references
      And all referenced environment variables are set
      When I load the configuration
      Then all variables are substituted correctly

    @L2 @ov
    Scenario: Claude-cli requires no environment variables
      Given an agent-config.yml file with provider "claude-cli"
      And no environment variables are set
      When I load the configuration
      Then the configuration is loaded successfully
      And no API key is required

  Rule: Invalid config fails fast with clear error

    Configuration validation must happen early and provide
    actionable error messages with instructions to run r2r init.

    @L2 @ov
    Scenario: Missing required provider name
      Given an agent-config.yml file without provider name
      When I load the configuration
      Then an error is returned
      And the error indicates "provider name is required"
      And the error suggests "run: r2r agent init"

    @L2 @ov
    Scenario: Empty configuration file
      Given an empty agent-config.yml file
      When I load the configuration
      Then an error is returned
      And the error indicates "configuration is invalid"
      And the error suggests "run: r2r agent init"

  Rule: Default configuration is safe to commit

    The repository includes a default .r2r/agent-config.yml file
    that contains no secrets and is safe to commit to version control.

    @L1 @ov
    Scenario: Default config uses claude-cli
      Given the default .r2r/agent-config.yml from repository
      When I inspect the configuration
      Then the provider is "claude-cli"
      And no API keys are specified
      And no sensitive information is present
      And the file is safe to commit to git
