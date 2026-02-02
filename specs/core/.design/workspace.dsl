workspace "EAC Core Library" "Core domain libraries for contracts, repository operations, and shared infrastructure" {

    model {
        # External actors and systems
        filesystem = softwareSystem "File System" "Repository files, contracts, and configuration" "External"
        git_system = softwareSystem "Git" "Version control operations" "External"

        # Dependents (modules that depend on eac-core per contracts)
        eac_commands = softwareSystem "EAC Commands" "CLI command implementations" "Dependent"
        eac_mcp_commands = softwareSystem "EAC MCP Commands" "MCP server for commands" "Dependent"
        r2r_cli = softwareSystem "R2R CLI" "Containerized workflow CLI" "Dependent"
        repository_mod = softwareSystem "Repository Module" "Repository validation rules" "Dependent"
        scripts_cli_installer = softwareSystem "Scripts CLI Installer" "CLI installer scripts" "Dependent"

        # Core library system
        eac_core = softwareSystem "EAC Core Library" "Foundational Go library providing contracts, repository operations, and shared infrastructure" {

            # Contracts subsystem
            contracts = container "Contracts" "Module contract definitions, loading, validation, and schema management" "Go Package" {
                contract_types = component "Contract Types" "Module, Environment, TestSuite type definitions" "Go"
                contract_loader = component "Contract Loader" "Loads contracts from YAML files" "Go"
                contract_validator = component "Contract Validator" "Validates contracts against JSON schema" "Go"
                contract_registry = component "Contract Registry" "In-memory registry of loaded contracts" "Go"
                schema_validator = component "Schema Validator" "JSON Schema validation engine" "Go"
                ai_types = component "AI Types" "AI prompt and template type definitions" "Go"
                ai_loader = component "AI Loader" "Loads AI prompts and templates" "Go"
                gherkin_validator = component "Gherkin Validator" "Validates Gherkin specification contracts" "Go"
            }

            # Repository subsystem
            repository = container "Repository" "Repository discovery, file mapping, and module operations" "Go Package" {
                repo_discovery = component "Repository Discovery" "Finds repository root and structure" "Go"
                file_mapper = component "File Mapper" "Maps files to owning modules" "Go"
                module_resolver = component "Module Resolver" "Resolves module dependencies and order" "Go"
                definitions_merger = component "Definitions Merger" "Merges definitions.yml files hierarchically" "Go"
                gomod_analyzer = component "Go.mod Analyzer" "Parses and validates go.mod dependencies" "Go"
                git_context = component "Git Context" "Provides git status and changed files" "Go"
            }

            # Configuration subsystem
            config = container "Configuration" "EAC configuration loading and management" "Go Package" {
                config_loader = component "Config Loader" "Loads .eac configuration" "Go"
                module_types = component "Module Types" "Defines available module types" "Go"
                test_suites = component "Test Suites" "Test suite configuration" "Go"
                environments = component "Environments" "Environment definitions" "Go"
                system_deps = component "System Dependencies" "External tool requirements" "Go"
                testing_tags = component "Testing Tags" "Test tag contract definitions" "Go"
            }

            # Git subsystem
            git_ops = container "Git Operations" "Git command execution and history analysis" "Go Package" {
                git_interface = component "Git Interface" "Abstract git operations interface" "Go"
                git_executor = component "Git Executor" "Executes git commands via os/exec" "Go"
                git_mock = component "Git Mock" "Mock implementation for testing" "Go"
                history_analyzer = component "History Analyzer" "Analyzes commit history" "Go"
            }

            # Changelog subsystem
            changelog = container "Changelog" "Changelog parsing, generation, and versioning" "Go Package" {
                changelog_parser = component "Changelog Parser" "Parses CHANGELOG.md files" "Go"
                changelog_writer = component "Changelog Writer" "Generates changelog entries" "Go"
                conventional_commits = component "Conventional Commits" "Parses conventional commit messages" "Go"
                semver_handler = component "Semver Handler" "Semantic versioning operations" "Go"
            }

            # Logging subsystem
            logging = container "Logging" "Structured logging with context and formatting" "Go Package" {
                logger_core = component "Logger Core" "Core logging implementation" "Go"
                log_config = component "Log Config" "Logging configuration" "Go"
                formatters = component "Formatters" "Output formatters (JSON, text)" "Go"
                component_logger = component "Component Logger" "Per-component logging context" "Go"
            }

            # Testing subsystem
            testing_pkg = container "Testing" "Test framework utilities and isolation" "Go Package" {
                test_context = component "Test Context" "Shared test context management" "Go"
                test_isolation = component "Test Isolation" "Isolated directory test environments" "Go"
                test_suite = component "Test Suite" "Test suite execution framework" "Go"
                feature_parser = component "Feature Parser" "Parses Gherkin feature files" "Go"
                test_reports = component "Test Reports" "Test result reporting" "Go"
            }

            # Environment subsystem
            environments_pkg = container "Environments" "Environment contract management and runtime detection" "Go Package" {
                env_contracts = component "Environment Contracts" "Environment type definitions" "Go"
                env_runtime = component "Environment Runtime" "Runtime environment detection" "Go"
            }

            # Module Dependencies
            module_deps = container "Module Dependencies" "Module dependency graph and verification" "Go Package" {
                dep_types = component "Dependency Types" "Dependency graph type definitions" "Go"
                dep_verifier = component "Dependency Verifier" "Verifies module dependencies" "Go"
            }

            # System Dependencies
            system_deps_pkg = container "System Dependencies" "External tool dependency verification" "Go Package" {
                sys_types = component "System Dep Types" "System dependency type definitions" "Go"
                sys_verifier = component "System Verifier" "Verifies external tools are available" "Go"
            }

            # Platform utilities
            platform = container "Platform" "Cross-platform utilities" "Go Package" {
                newline_handler = component "Newline Handler" "Platform-specific line endings" "Go"
                command_executor = component "Command Executor" "Platform-specific command execution" "Go"
            }

            # Other utilities
            utilities = container "Utilities" "Shared utility packages" "Go Package" {
                markdown_validator = component "Markdown Validator" "Validates markdown syntax" "Go"
                ai_mock = component "AI Mock" "Mock AI provider for testing" "Go"
                definitions_loader = component "Definitions Loader" "Loads definitions.yml files" "Go"
            }
        }

        # Dependent relationships (modules that import eac-core per contracts)
        eac_commands -> eac_core "Uses core libraries" "Go Import"
        eac_mcp_commands -> eac_core "Uses repository utilities" "Go Import"
        r2r_cli -> eac_core "Uses repository discovery" "Go Import"
        repository_mod -> eac_core "Uses validation libraries" "Go Import"
        scripts_cli_installer -> eac_core "Uses installer utilities" "Go Import"

        # External system relationships
        contracts -> filesystem "Loads contract YAML files" "File I/O"
        repository -> filesystem "Reads repository structure" "File I/O"
        repository -> git_system "Gets git status" "CLI"
        config -> filesystem "Loads configuration" "File I/O"
        git_ops -> git_system "Executes git commands" "CLI"
        changelog -> filesystem "Reads/writes changelogs" "File I/O"

        # Internal container relationships
        repository -> contracts "Uses contract definitions"
        repository -> git_ops "Gets git context"
        repository -> config "Loads module types"
        config -> contracts "Uses contract types"
        testing_pkg -> contracts "Validates test contracts"
        testing_pkg -> logging "Logs test execution"
        changelog -> git_ops "Gets commit history"
        module_deps -> contracts "Uses module contracts"
        system_deps_pkg -> config "Gets dependency config"

        # Component relationships - Contracts
        contract_loader -> contract_types "Creates contract instances"
        contract_loader -> schema_validator "Validates against schema"
        contract_validator -> schema_validator "Uses schema validation"
        contract_registry -> contract_loader "Stores loaded contracts"
        ai_loader -> ai_types "Creates AI prompt instances"
        gherkin_validator -> contract_types "Uses spec contracts"

        # Component relationships - Repository
        repo_discovery -> git_context "Uses git root detection"
        file_mapper -> module_resolver "Gets module boundaries"
        definitions_merger -> repo_discovery "Finds definition files"
        gomod_analyzer -> repo_discovery "Finds go.mod files"

        # Component relationships - Git
        git_executor -> git_interface "Implements interface"
        git_mock -> git_interface "Implements interface"
        history_analyzer -> git_interface "Uses git operations"

        # Component relationships - Changelog
        changelog_parser -> semver_handler "Parses versions"
        changelog_writer -> conventional_commits "Formats commits"
        changelog_writer -> semver_handler "Generates versions"

        # Component relationships - Logging
        logger_core -> log_config "Uses configuration"
        logger_core -> formatters "Formats output"
        component_logger -> logger_core "Wraps core logger"

        # Component relationships - Testing
        test_context -> test_isolation "Creates isolated environments"
        test_suite -> test_context "Uses shared context"
        test_suite -> feature_parser "Parses test features"
        test_suite -> test_reports "Generates reports"
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
            description "Shows all packages in the core library"
        }

        component contracts "ContractsComponents" {
            include *
            autoLayout tb
            title "Contracts Package - Components"
            description "Contract loading, validation, and schema management"
        }

        component repository "RepositoryComponents" {
            include *
            autoLayout tb
            title "Repository Package - Components"
            description "Repository discovery and file mapping"
        }

        component git_ops "GitComponents" {
            include *
            autoLayout tb
            title "Git Operations - Components"
            description "Git command execution and history"
        }

        component logging "LoggingComponents" {
            include *
            autoLayout tb
            title "Logging Package - Components"
            description "Structured logging infrastructure"
        }

        component testing_pkg "TestingComponents" {
            include *
            autoLayout tb
            title "Testing Package - Components"
            description "Test framework utilities"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
