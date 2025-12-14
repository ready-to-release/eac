 @L3 @iv @env:isolated-test-project
Feature: eac-commands_init
  As a developer
  I want to initialize EAC project configuration with a simple command
  So that I can quickly set up my project and optionally configure AI features

  Background:
    Given I am in a git repository

  Rule: Init without --ai-provider initializes project structure only
    The init command can be run without --ai-provider to just create the
    directory structure. AI provider configuration can be done later.


    Scenario: Init without --ai-provider creates directory structure
      Given no .r2r directory exists
      When I run "init" without any flags
      Then the command exits with code 0
      And the .r2r/eac directory is created
      And stdout contains "EAC project initialized"
      And stdout contains "To configure AI provider"

    Scenario: Init with invalid provider shows error
      Given no .r2r directory exists
      When I run "init --ai-provider invalid-provider"
      Then the command exits with code 1
      And stderr contains "unsupported provider"
      And stderr contains "Supported: claude-api, openai, gemini"

  Rule: Init creates .r2r/eac directory structure
    The init command must create the necessary directory structure
    for storing configuration at .r2r/eac/ai-provider.yml.

    Scenario: Init creates .r2r directory
      Given no .r2r directory exists
      When I run "init --ai-provider claude-api"
      Then the .r2r/eac directory is created
      And the command exits with code 0

    Scenario: Init works when .r2r directory already exists
      Given a .r2r directory already exists
      And no .r2r/eac/ai-provider.yml file exists
      When I run "init --ai-provider claude-api"
      Then the command exits with code 0
      And a .r2r/eac/ai-provider.yml file is created

  Rule: Init writes valid ai-provider.yml to .r2r/eac
    The generated configuration file must be valid YAML and contain
    environment variable references (not actual secrets).
    Config is stored at .r2r/eac/ai-provider.yml.

    Scenario: Init creates valid config for claude-api
      Given no .r2r directory exists
      When I run "init --ai-provider claude-api"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: claude-api"
      And the .r2r/eac/ai-provider.yml file contains "model: claude-3-haiku-20240307"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${ANTHROPIC_API_KEY}"

    Scenario: Init creates valid config for gemini
      Given no .r2r directory exists
      When I run "init --ai-provider gemini"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: gemini"
      And the .r2r/eac/ai-provider.yml file contains "model: gemini-1.5-pro"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${GOOGLE_API_KEY}"

    Scenario: Init creates valid config for openai
      Given no .r2r directory exists
      When I run "init --ai-provider openai"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: openai"
      And the .r2r/eac/ai-provider.yml file contains "model: gpt-4-turbo"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${OPENAI_API_KEY}"

    Scenario: Init shows helpful provider information
      Given no .r2r directory exists
      When I run "init --ai-provider claude-api"
      Then stdout contains provider selection confirmation
      And stdout contains API key instructions
      And stdout contains link to get API key
      And the command exits with code 0

    Scenario: Reinitializing requires --force flag
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      When I run "init --ai-provider openai"
      Then the command exits with code 1
      And stdout contains "Configuration already exists"
      And stdout contains "Use --force to overwrite existing config"

    Scenario: Reinitializing with --force overwrites existing config
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      When I run "init --ai-provider openai --force"
      Then stdout contains "Overwriting existing configuration"
      And the .r2r/eac/ai-provider.yml file contains "provider: openai"
      And the command exits with code 0
