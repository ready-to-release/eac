Feature: src-core-ai_config-loading
  As a developer
  I want to load AI provider configuration from .r2r/agent-config.yml
  So that the project can specify which AI provider to use

  Background:
    Given a repository with .r2r directory

  Rule: Configuration loads from .r2r/agent-config.yml
    The system must read configuration from the standard location
    and parse it correctly.

    @L2 @ov
    Scenario: Load valid configuration file
      Given a valid .r2r/agent-config.yml file with claude-api provider
      When I load the configuration
      Then the configuration is loaded successfully
      And the provider name is "claude-api"
      And the model is "claude-3-haiku-20240307"
      And the endpoint is "https://api.anthropic.com/v1"

    @L2 @ov
    Scenario: Configuration file does not exist
      Given no .r2r/agent-config.yml file exists
      When I attempt to load the configuration
      Then an error is returned
      And the error indicates the file was not found

    @L2 @ov
    Scenario: Configuration file contains malformed YAML
      Given a .r2r/agent-config.yml file with invalid YAML syntax
      When I attempt to load the configuration
      Then an error is returned
      And the error indicates YAML parsing failed

  Rule: Environment variables are substituted at runtime
    The configuration file should use ${VAR_NAME} syntax for secrets,
    and the system must substitute these with actual environment variable values.

    @L2 @ov
    Scenario: Substitute environment variable in API key
      Given a .r2r/agent-config.yml file with api_key: ${ANTHROPIC_API_KEY}
      And ANTHROPIC_API_KEY environment variable is set to "sk-test-123"
      When I load the configuration
      Then the API key in the loaded config is "sk-test-123"

    @L2 @ov
    Scenario: Missing environment variable results in empty string
      Given a .r2r/agent-config.yml file with api_key: ${MISSING_VAR}
      And MISSING_VAR environment variable is not set
      When I load the configuration
      Then the API key in the loaded config is empty

    @L2 @ov
    Scenario: Multiple environment variables are substituted
      Given a .r2r/agent-config.yml file with multiple ${VAR} references
      And all referenced environment variables are set
      When I load the configuration
      Then all variables are substituted correctly

  Rule: Invalid config fails fast with clear error
    Configuration validation must happen early and provide
    actionable error messages.

    @L2 @ov
    Scenario: Missing required provider name
      Given a .r2r/agent-config.yml file without provider name
      When I load the configuration
      Then an error is returned
      And the error indicates "provider name is required"

    @L2 @ov
    Scenario: Empty configuration file
      Given an empty .r2r/agent-config.yml file
      When I load the configuration
      Then an error is returned
      And the error indicates the configuration is invalid
