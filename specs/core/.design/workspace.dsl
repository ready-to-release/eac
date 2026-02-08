workspace "EAC Core Library" "Foundational Go library providing domain types, repository operations, tool execution, and shared infrastructure" {

    model {
        # External actors and systems
        filesystem = softwareSystem "File System" "Repository files, contracts, configuration, and cache" "External"
        git_system = softwareSystem "Git" "Version control operations via go-git" "External"
        github_api = softwareSystem "GitHub API" "Workflows, releases, container packages" "External"

        # Dependents (modules that depend on eac-core)
        eac_commands = softwareSystem "EAC Commands" "CLI command implementations" "Dependent"
        eac_mcp_commands = softwareSystem "EAC MCP Commands" "MCP server for commands" "Dependent"
        clibase_mod = softwareSystem "Clibase" "CLI execution framework" "Dependent"
        clie_cli = softwareSystem "CLIE CLI" "Containerized workflow CLI" "Dependent"
        adapters_mod = softwareSystem "Adapters" "Integration adapters (AI, Docker, TUI)" "Dependent"
        repository_mod = softwareSystem "Repository Module" "Repository validation rules" "Dependent"

        # Core library system
        eac_core = softwareSystem "EAC Core Library" "Foundational Go library providing domain types, repository operations, tool execution, scheduling, and shared infrastructure" {

            # Domain
            domain = container "Domain" "Core domain types including contracts, components, validation, and error types" "Go Package"

            # Repository subsystem
            repository = container "Repository" "Repository discovery, file mapping, module operations, and go.mod analysis" "Go Package"

            # Configuration subsystem
            config = container "Configuration" "Three-layer configuration merge (contract, user, personal) with schema validation" "Go Package"

            # Git subsystem
            git_ops = container "Git Operations" "Pure Go git operations using go-git (status, history, tags, branches)" "Go Package"

            # GitHub
            github = container "GitHub" "GitHub API abstractions for workflows, releases, container packages, and safety checks" "Go Package"

            # Changelog subsystem
            changelog = container "Changelog" "Parsing, generation, and version management for Keep a Changelog format" "Go Package"

            # Logging subsystem
            logging = container "Logging" "Dual-sink structured logging (console + rolling file) with TUI integration" "Go Package"

            # Testing subsystem
            testing_pkg = container "Testing" "Test discovery, tag inference, suite selection, and BDD test isolation" "Go Package"

            # Environment subsystem
            environments_pkg = container "Environments" "Environment variable constants, runtime detection (DevBox vs CI), and memory calculation" "Go Package"

            # Tool
            tool = container "Tool" "Unified pluggable tool composition system for any tool assigned to any component" "Go Package"

            # Workspace
            workspace_pkg = container "Workspace" "Detects workspace root using environment, container, or git walk-up" "Go Package"

            # Paths
            paths = container "Paths" "Centralized path constants and builder functions (zero internal dependencies)" "Go Package"

            # Scheduling
            scheduling = container "Scheduling" "Pull-based work unit scheduling with dependency resolution and LPT ordering" "Go Package"

            # Execution
            execution = container "Execution" "Cache verification interface for dependency-based work unit execution" "Go Package"

            # Workunit
            workunit = container "Work Unit" "Unified types for work unit identification, state management, locking, and cache invalidation" "Go Package"

            # AI
            ai = container "AI" "Consolidated AI access with prompt templating, structured generation, retry, and mocks" "Go Package"

            # Resolver
            resolver = container "Resolver" "Unified component-to-tool resolution mapping modules to executable work units" "Go Package"

            # Resource
            resource = container "Resource" "Domain types and port interfaces for resource pool management and capacity allocation" "Go Package"

            # Change Detection
            changedetect = container "Change Detection" "Hybrid git state plus file hash approach for unified change detection" "Go Package"

            # Hash
            hash = container "Hash" "Deterministic file content hashing with mtime-based cache layer" "Go Package"

            # Cache
            cache = container "Cache" "Two-dimensional taxonomy (Level x Type) for fine-grained cache control" "Go Package"

            # Evidence
            evidence = container "Evidence" "Writes and verifies security scan evidence files with SHA256 integrity" "Go Package"

            # Ownership
            ownership = container "Ownership" "Resolves file ownership to modules and components via directory-root specificity" "Go Package"

            # Ghost
            ghost = container "Ghost" "Discovers ghost-prefixed files for dark launching and feature toggles" "Go Package"

            # Docsync
            docsync = container "Docsync" "Scans CLI commands for missing or orphaned documentation files" "Go Package"

            # Release Notes
            releasenotes = container "Release Notes" "Parses, validates, and generates RELEASE-NOTES.md files" "Go Package"

            # Specs
            specs = container "Specs" "BDD specification parsing, scenario export, and Godog test runner" "Go Package"

            # Markdown
            markdown = container "Markdown" "Validation utilities for Markdown structure, code blocks, and heading hierarchy" "Go Package"

            # Module Dependencies
            module_deps = container "Module Dependencies" "Verifies availability of internal module dependencies by artifact or source" "Go Package"

            # Token Size
            tokensize = container "Token Size" "Estimates token counts for source files using character-based heuristic" "Go Package"

            # Defaults
            defaults = container "Defaults" "Default values and path derivation for module contracts" "Go Package"

            # Validation
            validation = container "Validation" "Structured validation types, error codes, and formatting utilities" "Go Package"

            # Platform
            platform = container "Platform" "Platform-specific abstractions for command execution and console output" "Go Package"

            # Adapters (bridge layer)
            adapters = container "Adapters" "Dependency-inversion bridge wrapping concrete domain types to satisfy port interfaces" "Go Package"
        }

        # Dependent relationships
        eac_commands -> eac_core "Uses core libraries" "Go Import"
        eac_mcp_commands -> eac_core "Uses repository utilities" "Go Import"
        clibase_mod -> eac_core "Uses config, modules, scheduling, tool registry" "Go Import"
        clie_cli -> eac_core "Uses repository discovery" "Go Import"
        adapters_mod -> eac_core "Uses config, testing, domain types" "Go Import"
        repository_mod -> eac_core "Uses validation libraries" "Go Import"

        # External system relationships
        repository -> filesystem "Reads repository structure" "File I/O"
        repository -> git_system "Gets git status and changed files" "go-git"
        config -> filesystem "Loads .eac configuration" "File I/O"
        git_ops -> git_system "Executes git operations" "go-git"
        changelog -> filesystem "Reads/writes changelogs" "File I/O"
        github -> github_api "Queries workflows, releases, packages" "HTTPS"
        evidence -> filesystem "Writes scan evidence with SHA256" "File I/O"
        hash -> filesystem "Reads files for hashing" "File I/O"
        logging -> filesystem "Writes rolling log files" "File I/O"
        changedetect -> filesystem "Reads files for change detection" "File I/O"
        workspace_pkg -> filesystem "Walks directories to find root" "File I/O"

        # Internal container relationships - Configuration layer
        config -> domain "Uses domain types"
        config -> defaults "Gets default values"
        config -> validation "Validates configuration"

        # Internal container relationships - Repository layer
        repository -> config "Loads module types"
        repository -> git_ops "Gets git context"
        repository -> domain "Uses domain types"
        repository -> ownership "Resolves file ownership"
        repository -> paths "Uses path utilities"

        # Internal container relationships - Execution layer
        tool -> domain "Uses tool definitions"
        tool -> paths "Uses workspace paths"
        resolver -> tool "Resolves tools for components"
        resolver -> config "Gets component types"
        scheduling -> workunit "Schedules work units"
        scheduling -> resource "Uses capacity allocation"
        execution -> workunit "Verifies work unit cache"
        workunit -> hash "Hashes for cache invalidation"

        # Internal container relationships - Change detection
        changedetect -> git_ops "Gets git status"
        changedetect -> hash "Hashes file content"
        changedetect -> config "Gets module boundaries"

        # Internal container relationships - Testing
        testing_pkg -> config "Gets test suites and tags"
        testing_pkg -> logging "Logs test execution"
        testing_pkg -> domain "Uses test domain types"
        specs -> testing_pkg "Uses test infrastructure"

        # Internal container relationships - Reporting
        changelog -> git_ops "Gets commit history"
        releasenotes -> changelog "Uses version information"
        docsync -> repository "Scans command documentation"

        # Internal container relationships - Utilities
        module_deps -> config "Gets module dependencies"
        ghost -> repository "Discovers ghost-prefixed files"
        evidence -> hash "Verifies integrity with SHA256"
        cache -> hash "Content-addressable caching"
        ai -> config "Gets AI configuration"

        # Internal container relationships - Bridge
        adapters -> domain "Wraps domain types for port interfaces"
        adapters -> config "Wraps config for port interfaces"
        adapters -> repository "Wraps repository for port interfaces"
    }

    views {
        systemContext eac_core "SystemContext" {
            include *
            autoLayout lr
            title "EAC Core Library - System Context"
            description "Shows core library with consumers and external systems"
        }

        container eac_core "Containers" {
            include *
            autoLayout tb
            title "EAC Core Library - Package Architecture"
            description "Shows all 34 packages in the core library"
        }

        container eac_core "ContractsComponents" {
            include domain config defaults validation adapters
            autoLayout tb
            title "Contracts Components"
            description "Contract loading, parsing, and validation"
        }

        container eac_core "RepositoryComponents" {
            include repository ownership paths workspace_pkg config git_ops domain
            autoLayout tb
            title "Repository Components"
            description "Repository discovery and file operations"
        }

        container eac_core "GitComponents" {
            include git_ops changelog changedetect hash github
            autoLayout tb
            title "Git Components"
            description "Git operations and repository state management"
        }

        container eac_core "LoggingComponents" {
            include logging environments_pkg platform
            autoLayout tb
            title "Logging Components"
            description "Structured logging with dual-sink output"
        }

        container eac_core "TestingComponents" {
            include testing_pkg specs domain logging config
            autoLayout tb
            title "Testing Components"
            description "Test utilities and shared test infrastructure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
