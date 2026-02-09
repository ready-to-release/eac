 @L3 @iv @env:isolated-test-project
Feature: eac-cli_init
  As a developer
  I want to initialize EAC project configuration with a simple command
  So that I can quickly set up my project and optionally configure AI features

  Background:
    Given I am in a git repository

  Rule: Init creates config files with calculated defaults
    The init command creates the .eac directory structure and generates
    configuration files (repository.yml, books.yml, environments.yml) with
    calculated defaults. AI provider configuration is optional.

    Scenario: Init without --ai-provider creates config files
      Given no .eac directory exists
      When I run "init" without any flags
      Then the command exits with code 0
      And the .eac directory is created
      And a .eac/repository.yml file is created
      And a .eac/books.yml file is created
      And a .eac/environments.yml file is created
      And stdout contains "EAC project initialized"
      And stdout contains "Configuration files created"

    Scenario: Init with invalid provider shows error
      Given no .eac directory exists
      When I run "init --ai-provider invalid-provider"
      Then the command exits with code 1
      And stderr contains "unsupported provider"
      And stderr contains "Supported: claude-api, openai, gemini"

  Rule: Init supports smart re-initialization
    The init command automatically detects existing configuration
    and intelligently updates it, preserving user customizations.

    Scenario: Init detects existing config and re-initializes
      Given a .eac/repository.yml file exists
      When I run "init"
      Then the command exits with code 0
      And stdout contains "Re-initializing EAC project"
      And stdout contains "Existing configuration detected"
      And stdout contains "EAC project re-initialized"

    Scenario: Init preserves user customizations during re-init
      Given a .eac/repository.yml file exists with custom module names
      When I run "init"
      Then the command exits with code 0
      And stdout contains "Re-initializing EAC project"
      And the custom module names are preserved in .eac/repository.yml

  Rule: Init creates valid ai-provider.yml when AI provider specified
    When --ai-provider is specified, init also creates ai-provider.yml
    with environment variable references (not actual secrets).

    Scenario: Init creates valid config for claude-api
      Given no .eac directory exists
      When I run "init --ai-provider claude-api"
      Then a .eac/ai-provider.yml file is created
      And the .eac/ai-provider.yml file contains "provider: claude-api"
      And the .eac/ai-provider.yml file contains "model: claude-3-haiku-20240307"
      And the .eac/ai-provider.yml file contains "api_key: ${ANTHROPIC_API_KEY}"
      And a .eac/repository.yml file is created
      And a .eac/books.yml file is created
      And a .eac/environments.yml file is created

    Scenario: Init creates valid config for gemini
      Given no .eac directory exists
      When I run "init --ai-provider gemini"
      Then a .eac/ai-provider.yml file is created
      And the .eac/ai-provider.yml file contains "provider: gemini"
      And the .eac/ai-provider.yml file contains "model: gemini-1.5-pro"
      And the .eac/ai-provider.yml file contains "api_key: ${GOOGLE_API_KEY}"

    Scenario: Init creates valid config for openai
      Given no .eac directory exists
      When I run "init --ai-provider openai"
      Then a .eac/ai-provider.yml file is created
      And the .eac/ai-provider.yml file contains "provider: openai"
      And the .eac/ai-provider.yml file contains "model: gpt-4-turbo"
      And the .eac/ai-provider.yml file contains "api_key: ${OPENAI_API_KEY}"

    Scenario: Init shows helpful provider information
      Given no .eac directory exists
      When I run "init --ai-provider claude-api"
      Then stdout contains provider selection confirmation
      And stdout contains API key instructions
      And stdout contains link to get API key
      And the command exits with code 0

  Rule: Init supports smart AI provider switching
    When ai-provider.yml already exists, re-running init with
    a different provider smoothly switches the configuration.

    Scenario: Reinitializing reuses existing AI provider
      Given a .eac/ai-provider.yml file exists with claude-api
      And a .eac/repository.yml file exists
      When I run "init"
      Then the command exits with code 0
      And stdout contains "Reusing existing AI provider: claude-api"
      And the .eac/ai-provider.yml file contains "provider: claude-api"

    Scenario: Reinitializing with different provider switches smoothly
      Given a .eac/ai-provider.yml file exists with claude-api
      And a .eac/repository.yml file exists
      And a .eac/books.yml file exists
      And a .eac/environments.yml file exists
      When I run "init --ai-provider openai"
      Then stdout contains "Switching AI provider: claude-api → openai"
      And the .eac/ai-provider.yml file contains "provider: openai"
      And the command exits with code 0
