 @L3 @iv @env:isolated-test-project
Feature: eac-commands_init
  As a developer
  I want to initialize AI provider configuration with a simple command
  So that I can quickly set up my project to use AI features

  Background:
    Given I am in a git repository

  Rule: Init requires --ai flag
    The init command must require the --ai flag to specify which provider to configure.
    This makes the choice explicit and prevents accidental misconfiguration.

   
    Scenario: Init without --ai flag shows error
      When I run "init" without any flags
      Then the command exits with code 1
      And stderr contains "--ai flag is required"
      And stderr contains "Available providers: claude-api, openai, gemini"

    Scenario: Init with invalid provider shows error
      Given no .r2r directory exists
      When I run "init --ai invalid-provider"
      Then the command exits with code 1
      And stderr contains "unsupported provider"
      And stderr contains "Supported: claude-api, openai, gemini"

  Rule: Init creates .r2r/eac directory structure
    The init command must create the necessary directory structure
    for storing configuration at .r2r/eac/ai-provider.yml.

    Scenario: Init creates .r2r directory
      Given no .r2r directory exists
      When I run "init --ai claude-api"
      Then the .r2r/eac directory is created
      And the command exits with code 0

    Scenario: Init works when .r2r directory already exists
      Given a .r2r directory already exists
      And no .r2r/eac/ai-provider.yml file exists
      When I run "init --ai claude-api"
      Then the command exits with code 0
      And a .r2r/eac/ai-provider.yml file is created

  Rule: Init writes valid ai-provider.yml to .r2r/eac
    The generated configuration file must be valid YAML and contain
    environment variable references (not actual secrets).
    Config is stored at .r2r/eac/ai-provider.yml.

    Scenario: Init creates valid config for claude-api
      Given no .r2r directory exists
      When I run "init --ai claude-api"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: claude-api"
      And the .r2r/eac/ai-provider.yml file contains "model: claude-3-haiku-20240307"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${ANTHROPIC_API_KEY}"

    Scenario: Init creates valid config for gemini
      Given no .r2r directory exists
      When I run "init --ai gemini"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: gemini"
      And the .r2r/eac/ai-provider.yml file contains "model: gemini-1.5-pro"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${GOOGLE_API_KEY}"

    Scenario: Init creates valid config for openai
      Given no .r2r directory exists
      When I run "init --ai openai"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: openai"
      And the .r2r/eac/ai-provider.yml file contains "model: gpt-4-turbo"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${OPENAI_API_KEY}"

    Scenario: Init shows helpful provider information
      Given no .r2r directory exists
      When I run "init --ai claude-api"
      Then stdout contains provider selection confirmation
      And stdout contains API key instructions
      And stdout contains link to get API key
      And the command exits with code 0

    Scenario: Reinitializing requires --force flag
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      When I run "init --ai openai"
      Then the command exits with code 1
      And stdout contains "Configuration already exists"
      And stdout contains "Use --force to overwrite existing config"

    Scenario: Reinitializing with --force overwrites existing config
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      When I run "init --ai openai --force"
      Then stdout contains "Overwriting existing configuration"
      And the .r2r/eac/ai-provider.yml file contains "provider: openai"
      And the command exits with code 0
