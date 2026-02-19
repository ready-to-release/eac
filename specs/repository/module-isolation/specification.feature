# Intent: Keep the architecture clean and modules independently buildable by enforcing module dependency rules across all Go modules
# Architecture: Affects repository Go module isolation validation; scans import statements in production .go files; enforces that go/core has no local deps, go/cli/clie is fully isolated, go/cli/eac depends only on go/core and go/clibase, and go/mcp/commands depends only on go/core; verifies DAG acyclicity

@L0 @ov @control:ac-6 @control:sa-3
Feature: repository_module-isolation

  As a repository maintainer
  I want to enforce module dependency rules across all Go modules
  So that the architecture remains clean and modules stay independently buildable

  Background:
    Given the repository contains the following Go modules:
      | Module          | Path              | Role                          |
      | go/core         | go/core           | Foundational library          |
      | go/cli/clie      | go/cli/clie        | CLI binary (isolated)         |
      | go/cli/eac      | go/cli/eac        | Command implementations + AI  |
      | go/clibase      | go/clibase        | Shared CLI framework          |
      | eac-ext         | containers/eac-ext| CLIE CLI extension (Docker)    |
      | go/mcp/commands | go/mcp/commands   | MCP server                    |

  Rule: go/core is the foundational module with no local dependencies

    go/core provides shared utilities (contracts, config, testing, git, etc.)
    and must not depend on any other local modules.

    @L0 @ov
    Scenario: go/core has no local module dependencies
      Given I am checking module "go/core"
      When I scan all .go files for import statements
      Then no files should import "github.com/ready-to-release/eac/go/cli/clie"
      And no files should import "github.com/ready-to-release/eac/go/cli/eac"
      And no files should import "github.com/ready-to-release/eac/go/mcp"
      And no files should import "github.com/ready-to-release/eac/go/clibase"

  Rule: go/cli/clie is fully isolated with no local dependencies

    The CLI binary must remain lightweight and independently distributable.
    Production code must not import any other local modules.
    Test code MAY import local modules for test infrastructure.

    @L0 @ov
    Scenario: go/cli/clie production code has no local module imports
      Given I am checking module "go/cli/clie"
      When I scan all production .go files in "go/cli/clie"
      Then no production files should import "github.com/ready-to-release/eac/go/core"
      And no production files should import "github.com/ready-to-release/eac/go/cli/eac"
      And no production files should import "github.com/ready-to-release/eac/go/mcp"
      And no production files should import "github.com/ready-to-release/eac/go/clibase"

  Rule: go/cli/eac depends only on go/core

    Command implementations use core utilities. AI integrations are internal
    to this module. They should not depend on CLI, MCP server, or test specs.

    @L0 @ov
    Scenario: go/cli/eac depends only on go/core and go/clibase
      Given I am checking module "go/cli/eac"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/core"
      And files may import "github.com/ready-to-release/eac/go/clibase"
      But no files should import "github.com/ready-to-release/eac/go/cli/clie"
      And no files should import "github.com/ready-to-release/eac/go/mcp"

  Rule: go/mcp/commands depends only on go/core

    MCP server uses core utilities for contract loading.
    It should not depend on commands or CLI.

    @L0 @ov
    Scenario: go/mcp/commands depends only on go/core
      Given I am checking module "go/mcp/commands"
      When I scan all .go files for import statements
      Then files may import "github.com/ready-to-release/eac/go/core"
      But no files should import "github.com/ready-to-release/eac/go/cli/clie"
      And no files should import "github.com/ready-to-release/eac/go/cli/eac"
      And no files should import "github.com/ready-to-release/eac/go/clibase"

  Rule: No circular dependencies between modules

    The module dependency graph must be a directed acyclic graph (DAG).

    @L0 @ov
    Scenario: Module dependency graph has no cycles
      When I build the module dependency graph from go.mod files
      Then the graph should have no circular dependencies
      And the dependency order should be:
        | Layer | Modules                                                     |
        | 0     | go/core                                                     |
        | 1     | go/cli/clie, go/clibase, go/mcp/commands, go/cli/eac        |
        | 2     | eac-ext                                                     |
