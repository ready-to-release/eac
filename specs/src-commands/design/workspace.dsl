workspace "Commands Module Architecture" "CLI command handlers and orchestration for the EAC framework" {

    model {
        # External actors and systems
        developer = person "Developer" "Developer using CLI commands"
        repository_module = softwareSystem "Repository Module" "Provides file and module information" "External"
        contracts_module = softwareSystem "Contracts Module" "Provides module contract definitions" "External"
        docker_daemon = softwareSystem "Docker Daemon" "Runs test and documentation containers" "External"
        claude_api = softwareSystem "Claude API" "AI service for commit message generation" "External"

        # Main commands system
        commands_system = softwareSystem "Commands Module" "CLI command handlers and orchestration" {

            # Core containers
            command_registry = container "Command Registry" "Registers and routes CLI commands to handlers, manages command discovery, validates command paths, and provides command metadata." "Go" "Core" {
                commandMap = component "Command Map" "Maps command names to handler functions" "Go (map[string]func)"
                commandRegistrar = component "Command Registrar" "init() registration for commands" "Go"
                pathMatcher = component "Path Matcher" "Matches multi-part command paths (e.g., 'test module')" "Go"
            }

            commit_ai_handler = container "Commit AI Handler" "Generates AI-powered commit messages using multi-agent architecture with top-level and per-module agents." "Go" "Command" {
                topLevelAgent = component "Top-Level Agent" "Generates commit header and overall summary" "Go (Claude API)"
                moduleAgent = component "Module Agent" "Generates per-module commit sections" "Go (Claude API)"
                contextBuilder = component "Context Builder" "Builds agent context from staged files and diff" "Go"
                messageValidator = component "Message Validator" "Validates commit message against contract" "Go"
                messageCleanup = component "Message Cleanup" "Auto-fixes formatting issues" "Go"
            }

            file_commands = container "File Commands" "Handles file listing, filtering, and module mapping commands (get files, show files, show files changed/staged)." "Go" "Command" {
                fileGetter = component "File Getter" "Gets files with module mappings" "Go"
                fileRenderer = component "File Renderer" "Renders files in table format" "Go"
                gitIntegration = component "Git Integration" "Gets changed and staged files" "Go"
            }

            module_commands = container "Module Commands" "Handles module listing, filtering, and inspection commands (get modules, show modules, show module types)." "Go" "Command" {
                moduleGetter = component "Module Getter" "Gets module contracts" "Go"
                moduleRenderer = component "Module Renderer" "Renders modules in table format" "Go"
                typeInspector = component "Type Inspector" "Lists available module types" "Go"
            }

            test_commands = container "Test Commands" "Manages Godog test execution in Docker containers with Cucumber/JUnit reporting (test module, test modules)." "Go" "Command" {
                testOrchestrator = component "Test Orchestrator" "Orchestrates multi-module test execution" "Go"
                containerRunner = component "Container Runner" "Runs Godog in Docker containers" "Go (Docker SDK)"
                reportGenerator = component "Report Generator" "Generates Cucumber/JUnit reports" "Go"
            }

            build_commands = container "Build Commands" "Manages Go module builds with proper module path resolution and dependency handling (build module)." "Go" "Command" {
                moduleBuilder = component "Module Builder" "Executes go build for modules" "Go (os/exec)"
                dependencyResolver = component "Dependency Resolver" "Resolves module dependencies" "Go"
            }

            summary_commands = container "Summary Commands" "Generates test summary reports for single and multi-module test runs (generate summary, generate summary multi)." "Go" "Command" {
                summaryGenerator = component "Summary Generator" "Generates formatted test summaries" "Go"
                reportAggregator = component "Report Aggregator" "Aggregates multi-module reports" "Go"
            }

            docs_commands = container "Docs Commands" "Manages MkDocs documentation server in Docker container (docs serve, docs build, docs stop)." "Go" "Command" {
                docsServer = component "Docs Server" "Starts/stops MkDocs server" "Go (Docker SDK)"
                docsBuilder = component "Docs Builder" "Builds static documentation site" "Go (Docker SDK)"
            }

            design_commands = container "Design Commands" "Manages Structurizr design server in Docker container (design serve, design build, design stop)." "Go" "Command" {
                designServer = component "Design Server" "Starts/stops Structurizr server" "Go (Docker SDK)"
                designBuilder = component "Design Builder" "Builds design artifacts" "Go (Docker SDK)"
            }

            render_engine = container "Render Engine" "Provides table rendering, custom formatters, and output formatting for all commands." "Go (go-pretty)" "Infrastructure" {
                tableBuilder = component "Table Builder" "Builds formatted tables" "Go (jedib0t/go-pretty)"
                customFormatter = component "Custom Formatter" "Custom column formatters" "Go"
                outputWriter = component "Output Writer" "Writes formatted output to stdout" "Go"
            }

        }

        # User interactions
        developer -> commands_system "Executes CLI commands" "CLI"

        # External system relationships
        commands_system -> repository_module "Gets file and module information" "Go package import"
        commands_system -> contracts_module "Loads module contracts" "Go package import"
        test_commands -> docker_daemon "Runs Godog test containers" "Docker API"
        docs_commands -> docker_daemon "Runs MkDocs containers" "Docker API"
        design_commands -> docker_daemon "Runs Structurizr containers" "Docker API"
        commit_ai_handler -> claude_api "Generates commit message text" "HTTPS/REST"

        # Internal relationships

        # Command Registry relationships
        command_registry -> commit_ai_handler "Routes commit-ai command" "Function calls"
        command_registry -> file_commands "Routes file commands" "Function calls"
        command_registry -> module_commands "Routes module commands" "Function calls"
        command_registry -> test_commands "Routes test commands" "Function calls"
        command_registry -> build_commands "Routes build commands" "Function calls"
        command_registry -> summary_commands "Routes summary commands" "Function calls"
        command_registry -> docs_commands "Routes docs commands" "Function calls"
        command_registry -> design_commands "Routes design commands" "Function calls"

        # Commit AI Handler relationships
        commit_ai_handler -> render_engine "Formats validation output" "Function calls"

        # File Commands relationships
        file_commands -> render_engine "Renders file tables" "Function calls"

        # Module Commands relationships
        module_commands -> render_engine "Renders module tables" "Function calls"

        # Test Commands relationships
        test_commands -> render_engine "Renders test results" "Function calls"

        # Summary Commands relationships
        summary_commands -> render_engine "Renders summaries" "Function calls"

        # Component relationships

        # Command Registry components
        commandRegistrar -> commandMap "Registers command handlers" "Go function calls"
        pathMatcher -> commandMap "Looks up commands by path" "Go function calls"

        # Commit AI Handler components
        contextBuilder -> topLevelAgent "Provides context for header generation" "Go function calls"
        contextBuilder -> moduleAgent "Provides context for module sections" "Go function calls"
        topLevelAgent -> messageValidator "Sends generated header for validation" "Go function calls"
        moduleAgent -> messageValidator "Sends generated sections for validation" "Go function calls"
        messageValidator -> messageCleanup "Sends validation results for cleanup" "Go function calls"

        # File Commands components
        fileGetter -> fileRenderer "Provides file data for rendering" "Go function calls"
        gitIntegration -> fileGetter "Provides git status information" "Go function calls"

        # Module Commands components
        moduleGetter -> moduleRenderer "Provides module data for rendering" "Go function calls"
        typeInspector -> moduleGetter "Gets module types" "Go function calls"

        # Test Commands components
        testOrchestrator -> containerRunner "Executes tests in containers" "Go function calls"
        containerRunner -> reportGenerator "Provides test results" "Go function calls"
    }

    views {
        # System Context view
        systemContext commands_system "SystemContext" {
            include *
            autoLayout lr
            title "Commands Module - System Context"
            description "Shows the Commands module with external dependencies"
        }

        # Container view
        container commands_system "Containers" {
            include *
            autoLayout tb
            title "Commands Module - Container Architecture"
            description "Shows all command handlers and infrastructure"
        }

        # Component views for key containers

        component command_registry "RegistryComponents" {
            include *
            autoLayout tb
            title "Command Registry - Components"
            description "Command registration and routing components"
        }

        component commit_ai_handler "CommitAIComponents" {
            include *
            autoLayout tb
            title "Commit AI Handler - Components"
            description "Multi-agent commit message generation"
        }

        component render_engine "RenderComponents" {
            include *
            autoLayout tb
            title "Render Engine - Components"
            description "Table rendering and output formatting"
        }

        # Filtered views

        container commands_system "CoreCommands" {
            include ->command_registry->
            include ->file_commands->
            include ->module_commands->
            include ->render_engine->
            autoLayout lr
            title "Core Commands"
            description "Core command handlers for file and module operations"
        }

        container commands_system "DockerCommands" {
            include ->test_commands->
            include ->docs_commands->
            include ->design_commands->
            autoLayout lr
            title "Docker Commands"
            description "Commands that orchestrate Docker containers"
        }

        # Styles
        styles {
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
                shape roundedbox
            }
            element "Core" {
                background #2E7D32
                color #ffffff
            }
            element "Command" {
                background #1976D2
                color #ffffff
            }
            element "Infrastructure" {
                background #F57C00
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
                shape component
            }
            relationship "Relationship" {
                thickness 2
                color #707070
                style solid
            }
            relationship "Function calls" {
                thickness 2
                color #2E7D32
            }
            relationship "Go package import" {
                thickness 2
                color #5E35B1
                style dashed
            }
            relationship "Docker API" {
                thickness 3
                color #1976D2
                style dashed
            }
            relationship "HTTPS/REST" {
                thickness 2
                color #D32F2F
                style dashed
            }
        }

        # Theme
        theme default
    }

    configuration {
        scope softwaresystem
    }

}
