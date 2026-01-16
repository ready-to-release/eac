 @L3 @iv @env:isolated-test-project
Feature: eac-commands_init
  As a developer
  I want to initialize EAC project configuration with a simple command
  So that I can quickly set up my project and optionally configure AI features

  Background:
    Given I am in a git repository

  Rule: Init creates config files with calculated defaults
    The init command creates the .r2r/eac directory structure and generates
    configuration files (repository.yml, books.yml, environments.yml) with
    calculated defaults. AI provider configuration is optional.

    Scenario: Init without --ai-provider creates config files
      Given no .r2r directory exists
      When I run "init" without any flags
      Then the command exits with code 0
      And the .r2r/eac directory is created
      And a .r2r/eac/repository.yml file is created
      And a .r2r/eac/books.yml file is created
      And a .r2r/eac/environments.yml file is created
      And stdout contains "EAC project initialized"
      And stdout contains "Configuration files created"

    Scenario: Init with invalid provider shows error
      Given no .r2r directory exists
      When I run "init --ai-provider invalid-provider"
      Then the command exits with code 1
      And stderr contains "unsupported provider"
      And stderr contains "Supported: claude-api, openai, gemini"

  Rule: Init fails if config files already exist without --force
    The init command must fail if configuration files already exist,
    unless the --force flag is provided.

    Scenario: Init fails when config files exist
      Given a .r2r/eac/repository.yml file exists
      When I run "init"
      Then the command exits with code 1
      And stdout contains "Configuration files already exist"
      And stdout contains "Use --force to overwrite existing files"

    Scenario: Init with --force overwrites existing config files
      Given a .r2r/eac/repository.yml file exists
      When I run "init --force"
      Then the command exits with code 0
      And stdout contains "Overwriting existing configuration files"
      And stdout contains "EAC project initialized"

  Rule: Init creates valid ai-provider.yml when AI provider specified
    When --ai-provider is specified, init also creates ai-provider.yml
    with environment variable references (not actual secrets).

    Scenario: Init creates valid config for claude-api
      Given no .r2r directory exists
      When I run "init --ai-provider claude-api"
      Then a .r2r/eac/ai-provider.yml file is created
      And the .r2r/eac/ai-provider.yml file contains "provider: claude-api"
      And the .r2r/eac/ai-provider.yml file contains "model: claude-3-haiku-20240307"
      And the .r2r/eac/ai-provider.yml file contains "api_key: ${ANTHROPIC_API_KEY}"
      And a .r2r/eac/repository.yml file is created
      And a .r2r/eac/books.yml file is created
      And a .r2r/eac/environments.yml file is created

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

  Rule: Init handles AI provider reconfiguration with --force
    When ai-provider.yml already exists, the --force flag allows
    overwriting the AI configuration.

    Scenario: Reinitializing AI provider requires --force flag
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      And a .r2r/eac/repository.yml file exists
      When I run "init --ai-provider openai"
      Then the command exits with code 1
      And stdout contains "Configuration files already exist"
      And stdout contains "Use --force to overwrite existing files"

    Scenario: Reinitializing AI provider with --force overwrites config
      Given a .r2r/eac/ai-provider.yml file exists with claude-api
      And a .r2r/eac/repository.yml file exists
      And a .r2r/eac/books.yml file exists
      And a .r2r/eac/environments.yml file exists
      When I run "init --ai-provider openai --force"
      Then stdout contains "Overwriting existing configuration files"
      And the .r2r/eac/ai-provider.yml file contains "provider: openai"
      And the command exits with code 0
