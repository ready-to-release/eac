@skip:wip @L2 @ov
Feature: src-commands_ai-contract-local-override

  As a developer using the cli tool
  I want AI configurations to be customizable in .r2r/eac/ai
  So that I can customize prompts for my project

  Background:
    AI configurations live in .r2r/eac/ai/<command>/ (e.g., .r2r/eac/ai/commit-message/).
    Unlike versioned contracts, AI configs are unversioned since they are project-specific.
    Each command directory contains: contract.yml, anti-corruption.yml, and prompt files (.md).

  Rule: AI configurations include prompt templates as inline content

    @L2 @ov
    Scenario: Commit message config includes top-level and module prompts
      Given the .r2r/eac/ai/commit-message/contract.yml file
      When I inspect the config structure
      Then it includes a "prompts" section
      And the prompts section contains "top_level" template
      And the prompts section contains "module" template

    @L2 @ov
    Scenario: Specification config includes generation prompt
      Given the .r2r/eac/ai/specifications/contract.yml file
      When I inspect the config structure
      Then it includes a "prompts" section
      And the prompts section contains "specification" template

  Rule: Loader loads AI configurations from .r2r/eac/ai

    @L2 @ov
    Scenario: Config loader reads from .r2r/eac/ai directory
      Given .r2r/eac/ai/commit-message/contract.yml exists
      When I load the AI config for "commit-message"
      Then the config is loaded from ".r2r/eac/ai/commit-message/"

    @L2 @ov
    Scenario: Prompt loader reads prompts from AI config directory
      Given .r2r/eac/ai/commit-message/top-level.md exists
      When I load the "top-level" prompt for commit
      Then the prompt is loaded from ".r2r/eac/ai/commit-message/top-level.md"
      And the source indicates "AI config"

    @L2 @ov
    Scenario: Prompt loader falls back to embedded prompts
      Given .r2r/eac/ai/commit-message/top-level.md does not exist
      When I load the "top-level" prompt for commit with embedded fallback
      Then the prompt is loaded from embedded resource
      And the source indicates "embedded default"

  Rule: AI commands use config loader

    @L2 @ov @deps:git
    Scenario: commit uses AI config prompt
      Given I have staged changes in git
      And .r2r/eac/ai/commit-message/contract.yml exists
      When I run "commit"
      Then the AI generation uses the prompt from .r2r/eac/ai/commit-message
      And the commit message follows the configured format

    @L2 @ov
    Scenario: create spec uses AI config prompt
      Given .r2r/eac/ai/specifications/contract.yml exists
      When I run "create spec 'Add user authentication'"
      Then the AI generation uses the prompt from .r2r/eac/ai/specifications
      And the specification follows the configured format

  Rule: Prompt templates can be customized without affecting validation

    @L2 @ov
    Scenario: Custom prompts still enforce config validation
      Given .r2r/eac/ai/commit-message/contract.yml with custom prompt
      And the config's validation rules are unchanged
      When I generate a commit message
      Then the output is validated against the config rules
      And violations are reported even with custom prompt

    @L2 @ov
    Scenario: Anti-corruption rules work with custom prompts
      Given .r2r/eac/ai/specifications/ with custom prompt
      And anti-corruption.yml defines forbidden patterns
      When I generate a specification
      Then anti-corruption filtering is applied
      And forbidden patterns are removed from output
