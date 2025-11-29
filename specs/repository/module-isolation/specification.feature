@L0 @ov
Feature: repository_module-isolation

  As a repository maintainer
  I want to enforce module dependency rules across all Go modules
  So that the architecture remains clean and modules stay independently buildable

  Background:
    Given the repository contains the following Go modules:
      | Module           | Path             | Role                          |
      | src/core         | src/core         | Foundational library          |
      | src/ai           | src/ai           | AI provider integrations      |
      | src/cli          | src/cli          | CLI binary (isolated)         |
      | src/commands     | src/commands     | Command implementations       |
      | src/commands/ext | src/commands/ext | Command extensions            |
      | src/mcp/commands | src/mcp/commands | MCP server                    |
      | src/specs        | src/specs        | BDD test implementations      |

  Rule: src/core is the foundational module with no local dependencies

    src/core provides shared utilities (contracts, config, testing, git, etc.)
    and must not depend on any other local modules.

    @L0 @ov
    Scenario: src/core has no local module dependencies
      Given I am checking module "src/core"
      When I scan all .go files for import statements
      Then no files should import "github.com/ready-to-release/eac/src/ai"
      And no files should import "github.com/ready-to-release/eac/src/cli"
      And no files should import "github.com/ready-to-release/eac/src/commands"
      And no files should import "github.com/ready-to-release/eac/src/mcp"
      And no files should import "github.com/ready-to-release/eac/src/specs"

  Rule: src/cli is fully isolated with no local dependencies

    The CLI binary must remain lightweight and independently distributable.
    Production code must not import any other local modules.
    Test code MAY import local modules for test infrastructure.

    @L0 @ov
    Scenario: src/cli production code has no local module imports
      Given I am checking module "src/cli"
      When I scan all production .go files in "src/cli"
      Then no production files should import "github.com/ready-to-release/eac/src/core"
      And no production files should import "github.com/ready-to-release/eac/src/ai"
      And no production files should import "github.com/ready-to-release/eac/src/commands"
      And no production files should import "github.com/ready-to-release/eac/src/mcp"
      And no production files should import "github.com/ready-to-release/eac/src/specs"

  Rule: src/ai depends only on src/core

    AI provider integrations use core utilities but should not depend on
    commands, CLI, or other higher-level modules.

    @L0 @ov
    Scenario: src/ai only depends on src/core
      Given I am checking module "src/ai"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/src/core"
      But no files should import "github.com/ready-to-release/eac/src/cli"
      And no files should import "github.com/ready-to-release/eac/src/commands"
      And no files should import "github.com/ready-to-release/eac/src/mcp"
      And no files should import "github.com/ready-to-release/eac/src/specs"

  Rule: src/commands depends on src/core and src/ai

    Command implementations use core utilities and AI integrations.
    They should not depend on CLI, MCP server, or test specs.

    @L0 @ov
    Scenario: src/commands depends only on src/core and src/ai
      Given I am checking module "src/commands"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/src/core"
      And files may import "github.com/ready-to-release/eac/src/ai"
      But no files should import "github.com/ready-to-release/eac/src/cli"
      And no files should import "github.com/ready-to-release/eac/src/mcp"
      And no files should import "github.com/ready-to-release/eac/src/specs"

  Rule: src/mcp/commands depends only on src/core

    MCP server uses core utilities for contract loading.
    It should not depend on commands, AI, or CLI.

    @L0 @ov
    Scenario: src/mcp/commands depends only on src/core
      Given I am checking module "src/mcp/commands"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/src/core"
      But no files should import "github.com/ready-to-release/eac/src/cli"
      And no files should import "github.com/ready-to-release/eac/src/ai"
      And no files should import "github.com/ready-to-release/eac/src/commands"
      And no files should import "github.com/ready-to-release/eac/src/specs"

  Rule: src/specs may depend on src/core for test utilities

    BDD test implementations use core utilities.
    They should not import production modules directly.

    @L0 @ov
    Scenario: src/specs depends only on src/core
      Given I am checking module "src/specs"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/src/core"
      But no files should import "github.com/ready-to-release/eac/src/cli"
      And no files should import "github.com/ready-to-release/eac/src/ai"
      And no files should import "github.com/ready-to-release/eac/src/commands"
      And no files should import "github.com/ready-to-release/eac/src/mcp"

  Rule: No circular dependencies between modules

    The module dependency graph must be a directed acyclic graph (DAG).

    @L0 @ov
    Scenario: Module dependency graph has no cycles
      When I build the module dependency graph from go.mod files
      Then the graph should have no circular dependencies
      And the dependency order should be:
        | Layer | Modules                              |
        | 0     | src/core                             |
        | 1     | src/ai, src/cli, src/mcp/commands, src/specs |
        | 2     | src/commands                         |
        | 3     | src/commands/ext                     |
