@ov
Feature: src-commands_ai-contract-local-override

  As a developer using the cli tool
  I want AI contracts to support local customization in .r2r/contracts
  So that I can customize prompts for my project without modifying the tool's source code

  Rule: AI contracts include prompt templates as inline content

    @ov
    Scenario: Commit message contract includes top-level and module prompts
      Given the contracts/ai/commit-message/0.1.0/contract.yml file
      When I inspect the contract structure
      Then it includes a "prompts" section
      And the prompts section contains "top_level" template
      And the prompts section contains "module" template

    @ov
    Scenario: Specification contract includes generation prompt
      Given the contracts/ai/specifications/0.1.0/contract.yml file
      When I inspect the contract structure
      Then it includes a "prompts" section
      And the prompts section contains "specification" template

  Rule: Init command copies AI contracts to .r2r/contracts for local customization

    @ov @deps:git
    Scenario: Init command creates .r2r/contracts directory structure
      Given I am in a git repository root
      When I run "init --ai claude-api"
      Then directory ".r2r/contracts/ai/commit-message/0.1.0" exists
      And directory ".r2r/contracts/ai/specifications/0.1.0" exists

    @ov @deps:git
    Scenario: Init command copies commit message contracts
      Given I am in a git repository root
      When I run "init --ai claude-api"
      Then file ".r2r/contracts/ai/commit-message/0.1.0/contract.yml" exists
      And file ".r2r/contracts/ai/commit-message/0.1.0/anti-corruption.yml" exists
      And the copied contract includes prompts section

    @ov @deps:git
    Scenario: Init command copies specification contracts
      Given I am in a git repository root
      When I run "init --ai claude-api"
      Then file ".r2r/contracts/ai/specifications/0.1.0/contract.yml" exists
      And file ".r2r/contracts/ai/specifications/0.1.0/anti-corruption.yml" exists
      And the copied contract includes prompts section

  Rule: Loader supports .r2r/contracts with fallback to repo defaults

    @ov
    Scenario: Contract loader uses .r2r/contracts when available
      Given .r2r/contracts/ai/commit-message/0.1.0/contract.yml exists
      And contracts/ai/commit-message/0.1.0/contract.yml exists in repo
      When I load the AI contract for "commit-message"
      Then the contract is loaded from ".r2r/contracts/ai/commit-message/0.1.0/"
      And the source path indicates "local override"

    @ov
    Scenario: Contract loader falls back to repo contracts
      Given .r2r/contracts/ai/commit-message/0.1.0/ does not exist
      And contracts/ai/commit-message/0.1.0/contract.yml exists in repo
      When I load the AI contract for "commit-message"
      Then the contract is loaded from "contracts/ai/commit-message/0.1.0/"
      And the source path indicates "repository default"

    @ov
    Scenario: Prompt loader uses .r2r/contracts prompts when available
      Given .r2r/contracts/ai/commit-message/0.1.0/top-level.md exists
      When I load the "top-level" prompt for commit
      Then the prompt is loaded from ".r2r/contracts/ai/commit-message/0.1.0/top-level.md"
      And the source indicates "local contract override"

    @ov
    Scenario: Prompt loader falls back to repo contract prompts
      Given .r2r/contracts/ does not exist
      And contracts/ai/commit-message/0.1.0/top-level.md exists
      When I load the "top-level" prompt for commit
      Then the prompt is loaded from repo contract
      And the source indicates "repository contract"

    @ov
    Scenario: Prompt loader falls back to embedded prompts
      Given .r2r/contracts/ does not exist
      And contracts/ai/commit-message/0.1.0/ does not exist
      When I load the "top-level" prompt for commit
      Then the prompt is loaded from embedded resource
      And the source indicates "embedded default"

  Rule: AI commands use contract loader with local override support

    @ov @deps:git
    Scenario: commit-ai uses local contract prompt when customized
      Given I have staged changes in git
      And .r2r/contracts/ai/commit-message/0.1.0/contract.yml exists with custom prompt
      When I run "commit-ai"
      Then the AI generation uses the custom prompt from .r2r/contracts
      And the commit message follows the custom prompt format

    @ov @deps:git
    Scenario: commit-ai falls back to repo contract prompts when no local override
      Given I have staged changes in git
      And .r2r/contracts directory does not exist
      When I run "commit-ai"
      Then the AI generation uses the prompt from contracts/ai/commit-message
      And the commit message follows the standard format

    @ov
    Scenario: specs create uses local contract prompt when customized
      Given .r2r/contracts/ai/specifications/0.1.0/contract.yml exists with custom prompt
      When I run "specs create 'Add user authentication'"
      Then the AI generation uses the custom prompt from .r2r/contracts
      And the specification follows the custom prompt format

    @ov
    Scenario: specs create falls back to repo contract prompts when no local override
      Given .r2r/contracts directory does not exist
      When I run "specs create 'Add user authentication'"
      Then the AI generation uses the prompt from contracts/ai/specifications
      And the specification follows the standard format

  Rule: Prompt templates can be customized without affecting validation

    @ov
    Scenario: Custom prompts still enforce contract validation
      Given .r2r/contracts/ai/commit-message/0.1.0/contract.yml with custom prompt
      And the contract's validation rules are unchanged
      When I generate a commit message
      Then the output is validated against the contract rules
      And violations are reported even with custom prompt

    @ov
    Scenario: Anti-corruption rules work with custom prompts
      Given .r2r/contracts/ai/specifications/0.1.0/ with custom prompt
      And anti-corruption.yml defines forbidden patterns
      When I generate a specification
      Then anti-corruption filtering is applied
      And forbidden patterns are removed from output
