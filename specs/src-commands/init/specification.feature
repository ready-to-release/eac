Feature: src-commands_init-ai-provider-configuration
  As a developer
  I want to initialize AI provider configuration with a simple command
  So that I can quickly set up my project to use AI features

  Background:
    Given I am in a git repository

  Rule: Init requires --ai flag
    The init command must require the --ai flag to specify which provider to configure.
    This makes the choice explicit and prevents accidental misconfiguration.

    @L1 @iv
    Scenario: Init without --ai flag shows error
      When I run "init" without any flags
      Then the command exits with code 1
      And stderr contains "Error: --ai flag is required"
      And stderr contains "Available providers: claude-api, claude-cli, openai, gemini"

    @L1 @iv
    Scenario: Init with invalid provider shows error
      When I run "init --ai invalid-provider"
      Then the command exits with code 1
      And stderr contains "unsupported provider"
      And stderr contains "Supported: claude-api, claude-cli, openai, gemini"

  Rule: Init creates .r2r directory structure
    The init command must create the necessary directory structure
    for storing configuration and logs.

    @L1 @iv
    Scenario: Init creates .r2r directory
      Given no .r2r directory exists
      When I run "init --ai claude-api"
      Then the .r2r directory is created
      And the .r2r/logs directory is created
      And the command exits with code 0

    @L1 @iv
    Scenario: Init works when .r2r directory already exists
      Given a .r2r directory already exists
      When I run "init --ai claude-api"
      Then the command exits with code 0
      And a reconfiguration message is shown

  Rule: Init writes valid agent-config.yml
    The generated configuration file must be valid YAML and contain
    environment variable references (not actual secrets).

    @L1 @iv
    Scenario: Init creates valid config for claude-api
      Given no .r2r directory exists
      When I run "init --ai claude-api"
      Then a .r2r/agent-config.yml file is created
      And the file contains "name: claude-api"
      And the file contains "model: claude-3-haiku-20240307"
      And the file contains "api_key: ${ANTHROPIC_API_KEY}"
      And the file does not contain actual secrets

    @L1 @iv
    Scenario: Init creates valid config for claude-cli
      Given no .r2r directory exists
      When I run "init --ai claude-cli"
      Then a .r2r/agent-config.yml file is created
      And the file contains "name: claude-cli"
      And the file contains "model: sonnet"
      And the file does not contain "api_key" field

    @L1 @iv
    Scenario: Init creates valid config for openai
      Given no .r2r directory exists
      When I run "init --ai openai"
      Then a .r2r/agent-config.yml file is created
      And the file contains "name: openai"
      And the file contains "model: gpt-4-turbo"
      And the file contains "api_key: ${OPENAI_API_KEY}"

    @L1 @iv
    Scenario: Init shows helpful provider information
      When I run "init --ai claude-api"
      Then stdout contains provider selection confirmation
      And stdout contains API key instructions
      And stdout contains link to get API key
      And the command exits with code 0

    @L1 @iv
    Scenario: Reinitializing overwrites existing config
      Given a .r2r/agent-config.yml file exists with claude-api
      When I run "init --ai openai"
      Then stdout contains "⚠️  Project already initialized"
      And stdout contains "Reconfiguring agent configuration"
      And the .r2r/agent-config.yml file contains "name: openai"
      And the command exits with code 0
