@L0 @ov
Feature: repository_module-isolation

  As a repository maintainer
  I want to enforce module dependency rules across all Go modules
  So that the architecture remains clean and modules stay independently buildable

  Background:
    Given the repository contains the following Go modules:
      | Module              | Path                | Role                          |
      | go/eac/core         | go/eac/core         | Foundational library          |
      | go/eac/ai           | go/eac/ai           | AI provider integrations      |
      | go/r2r/cli          | go/r2r/cli          | CLI binary (isolated)         |
      | go/eac/commands     | go/eac/commands     | Command implementations       |
      | ext-eac             | containers/ext-eac  | R2R CLI extension (Docker)    |
      | go/eac/mcp/commands | go/eac/mcp/commands | MCP server                    |
      | go/eac/specs        | go/eac/specs        | BDD test implementations      |

  Rule: go/eac/core is the foundational module with no local dependencies

    go/eac/core provides shared utilities (contracts, config, testing, git, etc.)
    and must not depend on any other local modules.

    @L0 @ov
    Scenario: go/eac/core has no local module dependencies
      Given I am checking module "go/eac/core"
      When I scan all .go files for import statements
      Then no files should import "github.com/ready-to-release/eac/go/eac/ai"
      And no files should import "github.com/ready-to-release/eac/go/r2r/cli"
      And no files should import "github.com/ready-to-release/eac/go/eac/commands"
      And no files should import "github.com/ready-to-release/eac/go/eac/mcp"
      And no files should import "github.com/ready-to-release/eac/go/eac/specs"

  Rule: go/r2r/cli is fully isolated with no local dependencies

    The CLI binary must remain lightweight and independently distributable.
    Production code must not import any other local modules.
    Test code MAY import local modules for test infrastructure.

    @L0 @ov
    Scenario: go/r2r/cli production code has no local module imports
      Given I am checking module "go/r2r/cli"
      When I scan all production .go files in "go/r2r/cli"
      Then no production files should import "github.com/ready-to-release/eac/go/eac/core"
      And no production files should import "github.com/ready-to-release/eac/go/eac/ai"
      And no production files should import "github.com/ready-to-release/eac/go/eac/commands"
      And no production files should import "github.com/ready-to-release/eac/go/eac/mcp"
      And no production files should import "github.com/ready-to-release/eac/go/eac/specs"

  Rule: go/eac/ai depends only on go/eac/core

    AI provider integrations use core utilities but should not depend on
    commands, CLI, or other higher-level modules.

    @L0 @ov
    Scenario: go/eac/ai only depends on go/eac/core
      Given I am checking module "go/eac/ai"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/eac/core"
      But no files should import "github.com/ready-to-release/eac/go/r2r/cli"
      And no files should import "github.com/ready-to-release/eac/go/eac/commands"
      And no files should import "github.com/ready-to-release/eac/go/eac/mcp"
      And no files should import "github.com/ready-to-release/eac/go/eac/specs"

  Rule: go/eac/commands depends on go/eac/core and go/eac/ai

    Command implementations use core utilities and AI integrations.
    They should not depend on CLI, MCP server, or test specs.

    @L0 @ov
    Scenario: go/eac/commands depends only on go/eac/core and go/eac/ai
      Given I am checking module "go/eac/commands"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/eac/core"
      And files may import "github.com/ready-to-release/eac/go/eac/ai"
      But no files should import "github.com/ready-to-release/eac/go/r2r/cli"
      And no files should import "github.com/ready-to-release/eac/go/eac/mcp"
      And no files should import "github.com/ready-to-release/eac/go/eac/specs"

  Rule: go/eac/mcp/commands depends only on go/eac/core

    MCP server uses core utilities for contract loading.
    It should not depend on commands, AI, or CLI.

    @L0 @ov
    Scenario: go/eac/mcp/commands depends only on go/eac/core
      Given I am checking module "go/eac/mcp/commands"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/eac/core"
      But no files should import "github.com/ready-to-release/eac/go/r2r/cli"
      And no files should import "github.com/ready-to-release/eac/go/eac/ai"
      And no files should import "github.com/ready-to-release/eac/go/eac/commands"
      And no files should import "github.com/ready-to-release/eac/go/eac/specs"

  Rule: go/eac/specs may depend on go/eac/core for test utilities

    BDD test implementations use core utilities.
    They should not import production modules directly.

    @L0 @ov
    Scenario: go/eac/specs depends only on go/eac/core
      Given I am checking module "go/eac/specs"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/eac/core"
      But no files should import "github.com/ready-to-release/eac/go/r2r/cli"
      And no files should import "github.com/ready-to-release/eac/go/eac/ai"
      And no files should import "github.com/ready-to-release/eac/go/eac/commands"
      And no files should import "github.com/ready-to-release/eac/go/eac/mcp"

  Rule: No circular dependencies between modules

    The module dependency graph must be a directed acyclic graph (DAG).

    @L0 @ov
    Scenario: Module dependency graph has no cycles
      When I build the module dependency graph from go.mod files
      Then the graph should have no circular dependencies
      And the dependency order should be:
        | Layer | Modules                                               |
        | 0     | go/eac/core                                           |
        | 1     | go/eac/ai, go/r2r/cli, go/eac/mcp/commands, go/eac/specs |
        | 2     | go/eac/commands                                       |
        | 3     | ext-eac                                               |
