workspace "Contracts" "Port interface definitions, embedded schemas, and configuration types for the EAC system" {

    model {
        # Consumers (modules that depend on contracts)
        eac_core = softwareSystem "EAC Core" "Foundational domain libraries" "Dependent"
        clibase = softwareSystem "Clibase" "CLI execution framework" "Dependent"
        eac_cli = softwareSystem "EAC CLI" "CLI command implementations" "Dependent"
        clie_cli = softwareSystem "CLIE CLI" "Containerized workflow CLI" "Dependent"
        adapters = softwareSystem "Adapters" "Integration adapters (AI, Docker, TUI, test runners)" "Dependent"
        mcp_server = softwareSystem "MCP Server" "MCP command server" "Dependent"

        # External systems
        filesystem = softwareSystem "File System" "YAML configs, JSON schemas, Go source" "External"

        # Contracts system
        contracts = softwareSystem "Contracts" "Hexagonal port interfaces, embedded schemas, and configuration types for the EAC system" {

            # Core contract module
            core_contract = container "Core Contract" "Foundational port interfaces, value types, and embedded schemas for the entire system" "Go Package" {
                config_ports = component "Configuration Ports" "ConfigPort, RepositoryConfigPort, EnvironmentsConfigPort, TestSuitesConfigPort, ComponentTypesConfigPort" "Go"
                logging_ports = component "Logging Ports" "LoggerPort, LoggerFactoryPort for structured logging" "Go"
                repository_ports = component "Repository Ports" "Workspace, Repository, GitRepository port interfaces" "Go"
                module_ports = component "Module Ports" "ModuleRegistryPort, ModuleContractPort, ComponentTypePort" "Go"
                tool_ports = component "Tool Ports" "ToolRegistryPort, ToolConfigPort, ToolDefPort for tool execution" "Go"
                unit_ports = component "Unit Ports" "UnitIDPort, UnitSpecPort, UnitRegistryPort, UnitResolverPort" "Go"
                observer_ports = component "Observer Ports" "ExecutionObserver, WriterFactory for event-driven monitoring" "Go"
                testing_ports = component "Testing Ports" "TestCachePort, TestIsolationPort, TestConfigPort, SuitePort" "Go"
                output_ports = component "Output Ports" "OutputReaderPort, OutputBufferPort, UoWManifestPort" "Go"
                tui_hooks = component "TUI Hooks" "TUIHooks interface for command-to-TUI bridge" "Go"
                services_port = component "Services Port" "SimpleServicesPort aggregate for dependency injection" "Go"
                validation_ports = component "Validation Ports" "RepositoryValidationPort, ConfigValidationPort" "Go"
                value_types = component "Value Types" "UnitID, UnitSpec, ActionType, TagSummary, PoolAllocation" "Go"
                embedded_schemas = component "Embedded Schemas" "40+ JSON schemas and 15+ YAML defaults via go:embed" "Go"
            }

            # AI provider contract
            ai_provider_contract = container "AI Provider Contract" "Interface for AI language model provider integration" "Go Package" {
                provider_port = component "Provider Interface" "Provider with Name() and Execute() methods" "Go"
                executor_port = component "Executor Interface" "Executor with RegisterProvider and GetLastUsedProvider" "Go"
                config_loader_port = component "ConfigLoader Interface" "Load and LoadWithOverrides for AI configuration" "Go"
                execute_options = component "Execute Options" "Model, Temperature, MaxTokens, Debug with functional Option pattern" "Go"
                ai_config_types = component "Config Types" "Config, AIConfig, GitConfig with env var substitution" "Go"
            }

            # Container runtime contract
            container_runtime_contract = container "Container Runtime Contract" "Runtime-agnostic interface for OCI container operations" "Go Package" {
                container_port = component "ContainerPort Interface" "Execute, Build, Pull, ImageExists, IsAvailable, Close" "Go"
                container_config = component "ContainerConfig" "Image, Command, Env, Mounts, User, Resources, Timeout, I/O writers" "Go"
                build_config = component "BuildConfig" "Context, Dockerfile, Tag, BuildArgs, Platforms, Push, Cache" "Go"
                container_result = component "ContainerResult" "ExitCode, Stdout, Stderr, Duration" "Go"
                resource_config = component "ResourceConfig" "CPUs, Memory, ShmSize limits" "Go"
            }

            # Scanner contract
            scanner_contract = container "Scanner Contract" "Security scanning configuration, scanner definitions, and OSCAL risk profiles" "Go Package" {
                security_config_port = component "SecurityConfigPort" "GetScanner, ListScanners, GetDefaultScanners, ShouldSkipModule" "Go"
                scanner_port = component "ScannerPort" "ID, Category, Image, Tag, Command, Timeout, Description" "Go"
                risk_config_port = component "RiskConfigPort" "GetProfile, GetModuleProfile, GetCatalogURL, GetScoring" "Go"
                profile_port = component "ProfilePort" "ControlIDs, HasControl, Title, Version, CatalogHref" "Go"
                scanner_types = component "Scanner Types" "ScannerDefinition, ScannersConfig, PoliciesConfig, category constants" "Go"
            }

            # TUI contract
            tui_contract = container "TUI Contract" "Contract between command orchestration and terminal UI implementations" "Go Package" {
                console_port = component "Console Interface" "Start, Stop, NewWriter, SendLine, SetPhase, UoW tracking (20 methods)" "Go"
                console_factory = component "ConsoleFactory" "Factory function type creating Console from Config" "Go"
                registry_port = component "Registry Interface" "Register, NewForCommand, MustHaveDefault, ListBindings" "Go"
                tui_config = component "Config Types" "Config, TUIConfig with timeout and layout parameters" "Go"
                tui_types = component "Display Types" "Line, Status, Phase, Level, SummaryData, PlannedWorkItem, UoWEnrichment" "Go"
            }

            # CLIE CLI contract
            clie_cli_contract = container "CLIE CLI Contract" "Embedded schemas and EBNF grammar for CLIE CLI configuration" "Go Package" {
                embedded_fs = component "Embedded FS" "go:embed filesystem carrier for schemas" "Go"
                cli_schema = component "CLI Schema" "clie-cli.schema.json for configuration validation" "JSON Schema"
                command_ebnf = component "Command EBNF" "EBNF grammar for CLI command parsing" "EBNF"
            }

            # Docs contract
            docs_contract = container "Docs Contract" "Embedded JSON schema for documentation manifest validation" "Go Package" {
                docs_embedded_fs = component "Embedded FS" "go:embed filesystem carrier for docs schema" "Go"
                manifest_schema = component "Manifest Schema" "manifest.schema.json for docs manifest validation" "JSON Schema"
            }
        }

        # Consumer relationships
        eac_core -> contracts "Implements port interfaces" "Go Import"
        clibase -> contracts "Uses port interfaces for orchestration" "Go Import"
        eac_cli -> contracts "Uses port interfaces for commands" "Go Import"
        clie_cli -> contracts "Uses configuration schemas" "Go Import"
        adapters -> contracts "Implements port interfaces" "Go Import"
        mcp_server -> contracts "Uses port interfaces" "Go Import"

        # External system relationships
        core_contract -> filesystem "Loads embedded schemas and YAML defaults" "go:embed"
        clie_cli_contract -> filesystem "Loads embedded CLI schemas" "go:embed"
        docs_contract -> filesystem "Loads embedded docs schema" "go:embed"
        scanner_contract -> filesystem "Loads embedded scanner defaults" "go:embed"

        # Internal dependency: TUI contract depends on Core contract
        tui_contract -> core_contract "Uses ActionType and TagSummary value types" "Go Import"

        # Core contract internal relationships
        config_ports -> value_types "Uses ActionType, configuration types"
        module_ports -> value_types "Uses module value types"
        tool_ports -> value_types "Uses tool value types"
        unit_ports -> value_types "Uses UnitID, UnitSpec, PoolAllocation"
        observer_ports -> value_types "Uses ExecutionEvent types"
        services_port -> config_ports "Aggregates configuration ports"
        services_port -> module_ports "Aggregates module ports"
        services_port -> tool_ports "Aggregates tool ports"
        testing_ports -> config_ports "Uses test configuration"
        validation_ports -> config_ports "Validates configuration"
        embedded_schemas -> config_ports "Provides default values"

        # AI provider internal relationships
        executor_port -> provider_port "Manages provider lifecycle"
        config_loader_port -> ai_config_types "Returns Config instances"
        provider_port -> execute_options "Accepts execution options"

        # Container runtime internal relationships
        container_port -> container_config "Accepts ContainerConfig for Execute"
        container_port -> build_config "Accepts BuildConfig for Build"
        container_port -> container_result "Returns ContainerResult"
        container_config -> resource_config "Uses resource limits"

        # Scanner internal relationships
        security_config_port -> scanner_port "Returns scanner instances"
        risk_config_port -> profile_port "Returns risk profiles"
        scanner_types -> scanner_port "Implements ScannerPort interface"

        # TUI internal relationships
        console_factory -> console_port "Creates Console instances"
        registry_port -> console_factory "Stores factory bindings"
        console_port -> tui_types "Uses Line, Status, Phase, SummaryData"
        console_port -> tui_config "Configured via Config"
    }

    views {
        systemContext contracts "SystemContext" {
            include *
            autoLayout lr
            title "Contracts - System Context"
            description "Shows the contracts system with all consumers and external systems"
        }

        container contracts "Containers" {
            include *
            autoLayout tb
            title "Contracts - Module Architecture"
            description "Shows all 7 contract modules and their relationships"
        }

        component core_contract "CoreContractComponents" {
            include *
            autoLayout tb
            title "Core Contract - Components"
            description "Port interfaces, value types, and embedded schemas"
        }

        component ai_provider_contract "AIProviderComponents" {
            include *
            autoLayout tb
            title "AI Provider Contract - Components"
            description "AI language model provider interfaces"
        }

        component container_runtime_contract "ContainerRuntimeComponents" {
            include *
            autoLayout tb
            title "Container Runtime Contract - Components"
            description "OCI container runtime interfaces and configuration types"
        }

        component scanner_contract "ScannerComponents" {
            include *
            autoLayout tb
            title "Scanner Contract - Components"
            description "Security scanning and OSCAL risk profile interfaces"
        }

        component tui_contract "TUIContractComponents" {
            include *
            autoLayout tb
            title "TUI Contract - Components"
            description "Terminal UI interfaces and display types"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
