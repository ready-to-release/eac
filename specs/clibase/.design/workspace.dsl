workspace "Clibase" "Shared CLI execution framework providing command orchestration, parallel execution, output rendering, and infrastructure services" {

    model {
        # External actors
        developer = person "Developer" "Developer executing CLI commands"

        # Dependents
        eac_cli = softwareSystem "EAC CLI" "CLI command implementations using clibase framework" "Dependent"
        mcp_server = softwareSystem "MCP Server" "MCP command server using clibase infrastructure" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Domain libraries for config, repository, modules" "Dependency"
        contracts_core = softwareSystem "Contracts Core" "Port interfaces and value types" "Dependency"

        # External systems
        filesystem = softwareSystem "File System" "Lock files, cache manifests, config files" "External"
        terminal = softwareSystem "Terminal" "stdout/stderr, TTY detection, terminal size" "External"
        docker_engine = softwareSystem "Docker Engine" "Container memory and CPU detection" "External"

        # Clibase system
        clibase = softwareSystem "Clibase" "Shared CLI execution framework with command orchestration, parallel execution, output rendering, and infrastructure services" {

            # Command Framework
            cmdframework = container "Command Framework" "5-phase command lifecycle: init, resolve, verify, execute, summary" "Go Package" {
                run_entry = component "Run Entry" "Run() and RunSimple() orchestration entry points" "Go"
                init_phase = component "Init Phase" "Config loading, orchestrator setup, TUI bootstrap" "Go"
                resolve_phase = component "Resolve Phase" "Module discovery, moniker expansion, dependency inclusion" "Go"
                verify_phase = component "Verify Phase" "Dependency verification, artifact validation, incremental detection" "Go"
                execute_phase = component "Execute Phase" "Work unit dispatch via orchestrator" "Go"
                summary_phase = component "Summary Phase" "TUI and console summary generation, exit codes" "Go"
                summary_builder = component "Summary Builder" "Incremental result accumulation" "Go"
                command_config = component "Command Config" "CommandConfig, ExecutionContext, Hooks types" "Go"
            }

            # Orchestrator
            orchestrator = container "Orchestrator" "Parallel execution engine with weighted semaphores, capacity management, and dependency scheduling" "Go Package" {
                orchestrator_core = component "Orchestrator Core" "Run(), RunLayered(), RunUnitsParallel() coordination" "Go"
                unit_scheduler = component "Unit Scheduler" "Component-level scheduler with LPT dispatch" "Go"
                display_manager = component "Display Manager" "Single-writer console loop preventing interleaved output" "Go"
                weighted_semaphore = component "Weighted Semaphore" "Capacity-based concurrency primitive with lock tracking" "Go"
                memory_detection = component "Memory Detection" "Docker/WSL/host CPU and memory detection" "Go"
                log_parser = component "Log Parser" "JSON events, Cucumber, CTRF output parsing" "Go"
            }

            # Registry
            registry = container "Command Registry" "Command registration and dispatch layer" "Go Package" {
                command_registration = component "Command Registration" "Register(), Get(), All() with metadata" "Go"
                flag_metadata = component "Flag Metadata" "FlagMetadata, DeclarativeFlagDefLike for validation" "Go"
            }

            # Flags
            flags = container "Flags" "Composable flag system with declarative support and typed accessors" "Go Package" {
                flag_sets = component "Flag Sets" "FlagSet interface, FlagDef definitions" "Go"
                flag_parser = component "Flag Parser" "Parser and ParsedFlags with typed accessors" "Go"
                execution_flags = component "Execution Flags" "Concurrency, sequential, turbo controls" "Go"
                output_flags = component "Output Flags" "TUI, ASCII mode, debug, timings" "Go"
                cache_flags = component "Cache Flags" "Cache control flags" "Go"
                module_flags = component "Module Flags" "Module selection and skip filters" "Go"
                declarative_flags = component "Declarative Flags" "Enable/disable pairs with conflict detection" "Go"
                flag_docs = component "Flag Documentation" "CommandDoc, FlagDocGenerator" "Go"
            }

            # Capacity
            capacity = container "Capacity" "Cross-process capacity management with weighted semaphores" "Go Package" {
                global_semaphore = component "Global Semaphore" "File-based weighted semaphore with stale cleanup" "Go"
                dual_pool = component "Dual Pool Semaphore" "Separate host and Docker capacity pools" "Go"
            }

            # Locking
            locking = container "Locking" "File-based distributed locking with tracker integration" "Go Package" {
                lock_acquire = component "Lock Acquire" "Acquire(), AcquireTracked(), AcquireWithWait()" "Go"
                lock_configs = component "Lock Configs" "BuildConfig, TestConfig, ScanConfig convenience" "Go"
            }

            # Lock Tracker
            locktracker = container "Lock Tracker" "Lock tracking and visualization registry with pub/sub events" "Go Package" {
                lock_registry = component "Lock Registry" "Thread-safe registry with Add/Remove/Summary" "Go"
                tracked_semaphore = component "Tracked Semaphore" "Channel-based semaphore with auto-registration" "Go"
                lock_events = component "Lock Events" "LockEvent, EventType for pub/sub notification" "Go"
            }

            # Display
            display = container "Display" "Interface definitions for TUI implementations" "Go Package" {
                console_interface = component "Console Interface" "Console with 22 methods for TUI rendering" "Go"
                display_types = component "Display Types" "Phase, SummaryData, InitSummary, PlannedWorkItem" "Go"
            }

            # Output
            output = container "Output" "Console output formatting for non-TUI mode" "Go Package" {
                console_observer = component "Console Observer" "ExecutionObserver for console output" "Go"
                format_utils = component "Format Utilities" "ResultLine, PackageDisplayName, ListFormat" "Go"
            }

            # Render
            render = container "Render" "Multi-format output rendering: markdown, JSON, YAML, TOML, custom" "Go Package" {
                table_builder = component "Table Builder" "Fluent builder for markdown tables" "Go"
                console_table = component "Console Table" "Terminal-width-aware table rendering" "Go"
                json_renderer = component "JSON Renderer" "JSON with OrderedMap for key ordering" "Go"
                yaml_renderer = component "YAML Renderer" "YAML rendering" "Go"
                toml_renderer = component "TOML Renderer" "TOML via YAML intermediate" "Go"
                custom_registry = component "Custom Registry" "Pluggable renderer registry (summary, count)" "Go"
            }

            # ANSI
            ansi = container "ANSI" "ANSI escape sequence filtering for clean text output" "Go Package" {
                ansi_filter = component "ANSI Filter" "Writer wrapper with StripAll and StripBad modes" "Go"
            }

            # Init Summary
            initsummary = container "Init Summary" "Data structures and formatters for initialization summaries" "Go Package" {
                summary_data = component "Summary Data" "Summary, Flags, DepsStatus, IncrementalInfo, ParallelismInfo" "Go"
                summary_formatter = component "Summary Formatter" "FormatCompact (console) and FormatDetailed (TUI)" "Go"
            }

            # Caching
            caching = container "Caching" "Incremental change detection and content-addressable item caching" "Go Package" {
                incremental = component "Incremental Detection" "DetectIncrementalChanges for module-level cache" "Go"
                item_cache = component "Item Cache" "Content-addressable cache with persistent Manifest" "Go"
            }

            # CTRF
            ctrf = container "CTRF" "Common Test Report Format types and utilities" "Go Package" {
                ctrf_types = component "CTRF Types" "Report, Results, Summary, Test, Tool, Status" "Go"
                ctrf_ops = component "CTRF Operations" "Parse, Merge, AddTest for report aggregation" "Go"
            }

            # Test Runners
            testrunners = container "Test Runners" "Registry-based test runner dispatch system" "Go Package" {
                runner_registry = component "Runner Registry" "Register, Get, AllDescriptors for test frameworks" "Go"
                streaming_runner = component "Streaming Runner" "Real-time go test -json parser" "Go"
                runner_interface = component "TestTypeRunner Interface" "6-method interface for test framework runners" "Go"
            }

            # Services
            services = container "Services" "Service initialization and dependency injection" "Go Package" {
                services_impl = component "Services Implementation" "Implements SimpleServicesPort with 15 adapter types" "Go"
            }

            # Environment
            environment = container "Environment" "Runtime environment detection" "Go Package" {
                env_detect = component "Environment Detector" "Detect(), ShouldUseTUI(), ValidateTUI()" "Go"
            }

            # FileUtil
            fileutil = container "File Utilities" "Atomic writes and platform-aware file cleanup" "Go Package" {
                atomic_write = component "Atomic Write" "AtomicWrite, AtomicWriteJSON via temp+rename" "Go"
                remove_retry = component "Remove With Retry" "Platform-specific cleanup with retry" "Go"
            }

            # Template
            template = container "Template" "Go template rendering utilities" "Go Package" {
                template_renderer = component "Template Renderer" "NewRenderer, RenderToString, RenderToFile" "Go"
            }

            # Test Util
            testutil = container "Test Utilities" "Test fixtures, assertions, output capture" "Go Package" {
                test_helpers = component "Test Helpers" "AssertContains, AssertEqual, Capture, FixtureConfig" "Go"
            }
        }

        # Actor relationships
        developer -> clibase "Executes CLI commands through framework" "CLI"

        # Dependent relationships
        eac_cli -> clibase "Uses framework for all commands" "Go Import"
        mcp_server -> clibase "Uses framework infrastructure" "Go Import"

        # Dependency relationships
        clibase -> core "Uses config, modules, scheduling, tool registry" "Go Import"
        clibase -> contracts_core "Uses port interfaces and value types" "Go Import"

        # External relationships
        capacity -> filesystem "File-based semaphore state" "File I/O"
        locking -> filesystem "File-based distributed locks" "File I/O"
        caching -> filesystem "Cache manifests and items" "File I/O"
        fileutil -> filesystem "Atomic file operations" "File I/O"
        environment -> terminal "Detects terminal capabilities" "System calls"
        orchestrator -> docker_engine "Detects container memory and CPU" "Docker API"
        orchestrator -> terminal "Single-writer console output" "stdout"

        # Internal container relationships
        cmdframework -> orchestrator "Dispatches work units for execution"
        cmdframework -> registry "Looks up command handlers"
        cmdframework -> flags "Parses command flags"
        cmdframework -> services "Initializes dependency injection"
        cmdframework -> environment "Detects runtime environment"
        cmdframework -> display "Sends events to TUI"
        cmdframework -> initsummary "Generates init-phase summary"
        cmdframework -> caching "Runs incremental change detection"
        orchestrator -> capacity "Acquires execution capacity"
        orchestrator -> locktracker "Reports lock state for visualization"
        orchestrator -> display "Sends execution events"
        orchestrator -> output "Formats console output"
        orchestrator -> ansi "Filters ANSI sequences from logs"
        capacity -> locktracker "Auto-registers capacity allocations"
        locking -> locktracker "Auto-registers file locks"
        testrunners -> ctrf "Uses CTRF report format"

        # Component relationships - cmdframework
        run_entry -> init_phase "Phase 1: Initialize"
        run_entry -> resolve_phase "Phase 2: Resolve scope"
        run_entry -> verify_phase "Phase 3: Verify dependencies"
        run_entry -> execute_phase "Phase 4: Execute work"
        run_entry -> summary_phase "Phase 5: Generate summary"
        execute_phase -> summary_builder "Accumulates results"
        summary_phase -> summary_builder "Reads accumulated results"

        # Component relationships - orchestrator
        orchestrator_core -> unit_scheduler "Delegates component-level scheduling"
        orchestrator_core -> display_manager "Routes output to single writer"
        unit_scheduler -> weighted_semaphore "Acquires execution capacity"
        unit_scheduler -> memory_detection "Recalculates capacity dynamically"
        orchestrator_core -> log_parser "Parses worker output"

        # Component relationships - capacity
        global_semaphore -> dual_pool "Coordinates host and Docker pools"

        # Component relationships - locking
        lock_acquire -> lock_configs "Uses pre-configured lock settings"

        # Component relationships - render
        table_builder -> console_table "Console-aware rendering"
        custom_registry -> json_renderer "Custom renderers use JSON"
    }

    views {
        systemContext clibase "SystemContext" {
            include *
            autoLayout lr
            title "Clibase - System Context"
            description "Shows clibase framework with CLI consumers and external dependencies"
        }

        container clibase "Containers" {
            include *
            autoLayout tb
            title "Clibase - Package Architecture"
            description "Shows all packages in the CLI infrastructure framework"
        }

        component cmdframework "CommandFramework" {
            include *
            autoLayout tb
            title "Command Framework - Components"
            description "5-phase command lifecycle orchestration"
        }

        component orchestrator "OrchestratorComponents" {
            include *
            autoLayout tb
            title "Orchestrator - Components"
            description "Parallel execution engine with capacity management"
        }

        component flags "FlagsComponents" {
            include *
            autoLayout tb
            title "Flags - Components"
            description "Composable flag system with typed accessors"
        }

        component render "RenderComponents" {
            include *
            autoLayout tb
            title "Render - Components"
            description "Multi-format output rendering"
        }

        component locktracker "LockTrackerComponents" {
            include *
            autoLayout tb
            title "Lock Tracker - Components"
            description "Lock tracking and visualization registry"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
