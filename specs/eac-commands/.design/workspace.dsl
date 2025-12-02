workspace "Commands Module Architecture" "CLI command handlers and orchestration for the EAC framework" {

    model {
        # External actors and systems
        developer = person "Developer" "Developer using CLI commands"
        repository_module = softwareSystem "Repository Module" "Provides file and module information" "External"
        contracts_module = softwareSystem "Contracts Module" "Provides module contract definitions" "External"
        docker_daemon = softwareSystem "Docker Daemon" "Runs test and documentation containers" "External"
        claude_api = softwareSystem "Claude API" "AI service for generation tasks" "External"
        github_api = softwareSystem "GitHub API" "GitHub integration for PRs and CI" "External"
        git_system = softwareSystem "Git" "Version control and worktree management" "External"

        # Main commands system
        commands_system = softwareSystem "Commands Module" "CLI command handlers and orchestration" {

            # Core containers
            command_registry = container "Command Registry" "Registers and routes CLI commands to handlers, manages command discovery, validates command paths, and provides command metadata." "Go" "Core" {
                commandMap = component "Command Map" "Maps command names to handler functions" "Go"
                commandRegistrar = component "Command Registrar" "Auto-registration for commands via init()" "Go"
                pathMatcher = component "Path Matcher" "Matches multi-part command paths" "Go"
                metadataExtractor = component "Metadata Extractor" "Extracts metadata from file comments" "Go"
            }

            commit_handler = container "Commit Command" "Generates AI-powered commit messages from staged changes using Claude API." "Go" "Command" {
                messageGenerator = component "Message Generator" "Generates commit messages via Claude" "Go"
                changeFilter = component "Change Filter" "Filters and groups staged changes" "Go"
                moduleAssembly = component "Module Assembly" "Assembles changes by module" "Go"
                validator = component "Validator" "Validates messages against contracts" "Go"
                cleanup = component "Cleanup" "Auto-fixes formatting issues" "Go"
            }

            query_commands = container "Query Commands" "Handles file, module, test, dependency and environment retrieval commands (get/show modules, files, dependencies, environments, tests)." "Go" "Command" {
                fileGetter = component "File Getter" "Gets files with module mappings" "Go"
                moduleGetter = component "Module Getter" "Gets module contracts and metadata" "Go"
                dependencyGetter = component "Dependency Getter" "Analyzes module dependencies" "Go"
                environmentGetter = component "Environment Getter" "Retrieves environment configurations" "Go"
                testGetter = component "Test Getter" "Retrieves test suite information" "Go"
                gitIntegration = component "Git Integration" "Gets changed and staged files" "Go"
            }

            inspection_commands = container "Inspection Commands" "Handles module type inspection, suite listing, and command description (show moduletypes, list-suites, describe)." "Go" "Command" {
                typeInspector = component "Type Inspector" "Lists available module types" "Go"
                suiteInspector = component "Suite Inspector" "Lists test suites" "Go"
                commandInspector = component "Command Inspector" "Describes available commands" "Go"
                completionProvider = component "Completion Provider" "Shell completion integration" "Go"
            }

            test_commands = container "Test Commands" "Manages Godog test execution with Docker containers and generates Cucumber/JUnit reports." "Go" "Command" {
                suiteRunner = component "Suite Runner" "Orchestrates test suite execution" "Go"
                containerRunner = component "Container Runner" "Runs Godog in Docker containers" "Go"
                reportGenerator = component "Report Generator" "Generates Cucumber/JUnit reports" "Go"
                reportAggregator = component "Report Aggregator" "Aggregates multi-module test results" "Go"
            }

            build_commands = container "Build Commands" "Manages Go module builds with dependency ordering and parallel execution." "Go" "Command" {
                moduleBuilder = component "Module Builder" "Executes go build for modules" "Go"
                buildOrchestrator = component "Build Orchestrator" "Orchestrates multi-module builds" "Go"
                dependencyResolver = component "Dependency Resolver" "Resolves module dependencies" "Go"
            }

            pipeline_commands = container "Pipeline Commands" "Orchestrates module pipelines with dependency respecting and status tracking." "Go" "Command" {
                pipelineRunner = component "Pipeline Runner" "Executes module pipelines" "Go"
                pipelineOrchestrator = component "Pipeline Orchestrator" "Manages pipeline execution state" "Go"
                pipelineStatus = component "Pipeline Status" "Tracks and reports pipeline state" "Go"
            }

            work_commands = container "Work Commands" "Git worktree-based parallel development with branch, PR, and merge management." "Go" "Command" {
                worktreeManager = component "Worktree Manager" "Manages git worktrees" "Go"
                workflowOrchestrator = component "Workflow Orchestrator" "Coordinates work operations" "Go"
                branchManager = component "Branch Manager" "Handles branch operations" "Go"
                prManager = component "PR Manager" "Creates and manages pull requests" "Go"
            }

            design_commands = container "Design Commands" "Manages Structurizr DSL architecture diagrams with AI-powered generation and validation." "Go" "Command" {
                designGenerator = component "Design Generator" "Generates DSL from architecture description" "Go"
                designValidator = component "Design Validator" "Validates DSL syntax" "Go"
                designServer = component "Design Server" "Runs Structurizr Lite web viewer" "Go"
                validationFormatter = component "Validation Formatter" "Formats validation errors" "Go"
            }

            docs_commands = container "Docs Commands" "Manages MkDocs documentation server and static site generation in Docker." "Go" "Command" {
                docsServer = component "Docs Server" "Starts/stops MkDocs server" "Go"
                docsBuilder = component "Docs Builder" "Builds static documentation" "Go"
                browserIntegration = component "Browser Integration" "Auto-opens documentation in browser" "Go"
            }

            templates_commands = container "Templates Commands" "Manages project templates with variable substitution and AI validation." "Go" "Command" {
                templateApplier = component "Template Applier" "Applies templates to projects" "Go"
                templateInstaller = component "Template Installer" "Installs template dependencies" "Go"
                templateScanner = component "Template Scanner" "Lists available templates" "Go"
                valueSubstitutor = component "Value Substitutor" "Substitutes template variables" "Go"
                securityValidator = component "Security Validator" "Validates template security" "Go"
            }

            specs_commands = container "Specs Commands" "Manages Gherkin specifications with AI-powered generation and validation." "Go" "Command" {
                specsGenerator = component "Specs Generator" "Generates specs from descriptions" "Go"
                specsValidator = component "Specs Validator" "Validates spec syntax and contracts" "Go"
                specSecurity = component "Spec Security" "Validates spec security requirements" "Go"
            }

            validate_commands = container "Validate Commands" "Validates module contracts and dependency configurations." "Go" "Command" {
                contractValidator = component "Contract Validator" "Validates module contracts" "Go"
                dependencyValidator = component "Dependency Validator" "Validates dependency graphs" "Go"
            }

            other_commands = container "Other Commands" "Miscellaneous commands including help, init, release, and CI utilities." "Go" "Command" {
                helpProvider = component "Help Provider" "Provides command help and documentation" "Go"
                initProvider = component "Init Provider" "Project initialization" "Go"
                releaseManager = component "Release Manager" "Handles CalVer releases" "Go"
                ciUtilities = component "CI Utilities" "CI-specific utilities" "Go"
            }

            render_engine = container "Render Engine" "Provides table rendering, JSON, TOML output, and custom formatters for all commands." "Go" "Infrastructure" {
                tableBuilder = component "Table Builder" "Builds formatted markdown tables" "Go"
                jsonRenderer = component "JSON Renderer" "Renders output as JSON" "Go"
                tomlRenderer = component "TOML Renderer" "Renders output as TOML" "Go"
                customFormatter = component "Custom Formatter" "Custom column and value formatters" "Go"
                structRenderer = component "Struct Renderer" "Converts structs to formatted output" "Go"
                outputWriter = component "Output Writer" "Writes formatted output to stdout" "Go"
            }

            orchestrator = container "Orchestrator" "Provides dependency-aware execution ordering, parallel execution, and progress tracking for multi-module operations." "Go" "Infrastructure" {
                executionPlanner = component "Execution Planner" "Plans module execution order" "Go"
                parallelExecutor = component "Parallel Executor" "Executes modules in parallel" "Go"
                stateTracker = component "State Tracker" "Tracks execution state" "Go"
                progressManager = component "Progress Manager" "Manages progress display" "Go"
            }

            serve_framework = container "Serve Framework" "Manages container-based services with port conflict detection and browser integration." "Go" "Infrastructure" {
                containerManager = component "Container Manager" "Manages Docker containers" "Go"
                portManager = component "Port Manager" "Detects and resolves port conflicts" "Go"
                serviceOrchestrator = component "Service Orchestrator" "Coordinates service lifecycle" "Go"
                browserManager = component "Browser Manager" "Auto-opens URLs in browser" "Go"
            }

            template_engine = container "Template Engine" "Provides prompt template rendering and variable substitution for AI interactions." "Go" "Infrastructure" {
                promptRenderer = component "Prompt Renderer" "Renders AI prompts from templates" "Go"
                variableSubstitutor = component "Variable Substitutor" "Substitutes template variables" "Go"
            }

        }

        # User interactions
        developer -> commands_system "Executes CLI commands" "CLI"

        # External system relationships
        commands_system -> repository_module "Gets file and module information" "Go package import"
        commands_system -> contracts_module "Loads module contracts" "Go package import"
        commands_system -> git_system "Performs git operations" "Go (os/exec)"
        
        # Docker relationships
        test_commands -> docker_daemon "Runs Godog test containers" "Docker API"
        docs_commands -> docker_daemon "Runs MkDocs containers" "Docker API"
        design_commands -> docker_daemon "Validates DSL via Docker" "Docker API"

        # API relationships
        commit_handler -> claude_api "Generates commit messages" "HTTPS/REST"
        design_commands -> claude_api "Generates architecture diagrams" "HTTPS/REST"
        templates_commands -> claude_api "Validates templates" "HTTPS/REST"
        specs_commands -> claude_api "Generates specifications" "HTTPS/REST"
        work_commands -> github_api "Creates pull requests" "HTTPS/REST"
        other_commands -> github_api "CI integrations" "HTTPS/REST"

        # Internal command routing
        command_registry -> commit_handler "Routes commit command" "Function calls"
        command_registry -> query_commands "Routes query commands" "Function calls"
        command_registry -> inspection_commands "Routes inspection commands" "Function calls"
        command_registry -> test_commands "Routes test commands" "Function calls"
        command_registry -> build_commands "Routes build commands" "Function calls"
        command_registry -> pipeline_commands "Routes pipeline commands" "Function calls"
        command_registry -> work_commands "Routes work commands" "Function calls"
        command_registry -> design_commands "Routes design commands" "Function calls"
        command_registry -> docs_commands "Routes docs commands" "Function calls"
        command_registry -> templates_commands "Routes template commands" "Function calls"
        command_registry -> specs_commands "Routes specs commands" "Function calls"
        command_registry -> validate_commands "Routes validate commands" "Function calls"
        command_registry -> other_commands "Routes other commands" "Function calls"

        # Render engine relationships
        commit_handler -> render_engine "Renders output" "Function calls"
        query_commands -> render_engine "Renders tables and data" "Function calls"
        inspection_commands -> render_engine "Renders command info" "Function calls"
        test_commands -> render_engine "Renders test results" "Function calls"
        build_commands -> render_engine "Renders build results" "Function calls"
        pipeline_commands -> render_engine "Renders pipeline status" "Function calls"
        work_commands -> render_engine "Renders worktree info" "Function calls"
        validate_commands -> render_engine "Renders validation results" "Function calls"
        other_commands -> render_engine "Renders output" "Function calls"

        # Orchestrator relationships
        build_commands -> orchestrator "Uses for multi-module builds" "Function calls"
        test_commands -> orchestrator "Uses for test orchestration" "Function calls"
        pipeline_commands -> orchestrator "Uses for pipeline execution" "Function calls"

        # Serve framework relationships
        design_commands -> serve_framework "Manages Structurizr server" "Function calls"
        docs_commands -> serve_framework "Manages documentation server" "Function calls"

        # Template engine relationships
        commit_handler -> template_engine "Renders commit prompts" "Function calls"
        design_commands -> template_engine "Renders design prompts" "Function calls"
        templates_commands -> template_engine "Substitutes variables" "Function calls"
        specs_commands -> template_engine "Renders spec prompts" "Function calls"
        work_commands -> template_engine "Renders PR prompts" "Function calls"

        # Component relationships

        # Command Registry components
        commandRegistrar -> commandMap "Registers command handlers" "Go function calls"
        pathMatcher -> commandMap "Looks up commands by path" "Go function calls"
        metadataExtractor -> commandMap "Stores metadata" "Go function calls"

        # Commit Handler components
        changeFilter -> moduleAssembly "Provides grouped changes" "Go function calls"
        moduleAssembly -> messageGenerator "Provides module context" "Go function calls"
        messageGenerator -> validator "Sends generated message for validation" "Go function calls"
        validator -> cleanup "Sends results for cleanup" "Go function calls"

        # Query Commands components
        gitIntegration -> fileGetter "Provides git status" "Go function calls"
        fileGetter -> moduleGetter "Retrieves module info" "Go function calls"

        # Inspection Commands components
        typeInspector -> commandInspector "Provides type metadata" "Go function calls"
        suiteInspector -> commandInspector "Provides suite metadata" "Go function calls"

        # Test Commands components
        suiteRunner -> containerRunner "Executes tests in containers" "Go function calls"
        containerRunner -> reportGenerator "Provides test results" "Go function calls"
        reportGenerator -> reportAggregator "Aggregates reports" "Go function calls"

        # Build Commands components
        buildOrchestrator -> dependencyResolver "Resolves dependencies" "Go function calls"
        buildOrchestrator -> moduleBuilder "Executes builds" "Go function calls"

        # Pipeline Commands components
        pipelineRunner -> pipelineOrchestrator "Manages state" "Go function calls"
        pipelineOrchestrator -> pipelineStatus "Updates status" "Go function calls"

        # Work Commands components
        worktreeManager -> workflowOrchestrator "Coordinates operations" "Go function calls"
        workflowOrchestrator -> branchManager "Manages branches" "Go function calls"
        branchManager -> prManager "Creates pull requests" "Go function calls"

        # Design Commands components
        designGenerator -> designValidator "Validates generated DSL" "Go function calls"
        designValidator -> validationFormatter "Formats errors" "Go function calls"

        # Docs Commands components
        docsServer -> browserIntegration "Opens documentation" "Go function calls"

        # Templates Commands components
        templateScanner -> templateApplier "Lists available templates" "Go function calls"
        templateApplier -> valueSubstitutor "Substitutes variables" "Go function calls"
        templateInstaller -> securityValidator "Validates security" "Go function calls"

        # Specs Commands components
        specsGenerator -> specsValidator "Validates generated specs" "Go function calls"
        specsValidator -> specSecurity "Validates security" "Go function calls"

        # Orchestrator components
        executionPlanner -> parallelExecutor "Provides execution plan" "Go function calls"
        parallelExecutor -> stateTracker "Updates state" "Go function calls"
        stateTracker -> progressManager "Provides state for display" "Go function calls"

        # Serve Framework components
        containerManager -> portManager "Manages ports for containers" "Go function calls"
        serviceOrchestrator -> containerManager "Manages containers" "Go function calls"
        serviceOrchestrator -> browserManager "Opens browsers" "Go function calls"

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

        component commit_handler "CommitComponents" {
            include *
            autoLayout tb
            title "Commit Command - Components"
            description "AI-powered commit message generation"
        }

        component query_commands "QueryComponents" {
            include *
            autoLayout tb
            title "Query Commands - Components"
            description "Repository data retrieval and inspection"
        }

        component test_commands "TestComponents" {
            include *
            autoLayout tb
            title "Test Commands - Components"
            description "Test execution and reporting"
        }

        component design_commands "DesignComponents" {
            include *
            autoLayout tb
            title "Design Commands - Components"
            description "Architecture diagram generation and validation"
        }

        component render_engine "RenderComponents" {
            include *
            autoLayout tb
            title "Render Engine - Components"
            description "Output formatting and rendering"
        }

        component orchestrator "OrchestratorComponents" {
            include *
            autoLayout tb
            title "Orchestrator - Components"
            description "Multi-module execution orchestration"
        }

        # Filtered views

        container commands_system "QueryAndInspection" {
            include ->command_registry->
            include ->query_commands->
            include ->inspection_commands->
            include ->render_engine->
            autoLayout lr
            title "Query and Inspection Commands"
            description "Commands for retrieving and inspecting repository data"
        }

        container commands_system "ExecutionCommands" {
            include ->command_registry->
            include ->test_commands->
            include ->build_commands->
            include ->pipeline_commands->
            include ->orchestrator->
            autoLayout lr
            title "Execution Commands"
            description "Commands for building, testing, and executing pipelines"
        }

        container commands_system "InfrastructureCommands" {
            include ->command_registry->
            include ->design_commands->
            include ->docs_commands->
            include ->serve_framework->
            autoLayout lr
            title "Infrastructure Commands"
            description "Commands for managing architecture, documentation, and services"
        }

        container commands_system "DevelopmentCommands" {
            include ->command_registry->
            include ->commit_handler->
            include ->work_commands->
            include ->templates_commands->
            include ->specs_commands->
            autoLayout lr
            title "Development Commands"
            description "Commands for development workflows and artifact generation"
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
            relationship "Go (os/exec)" {
                thickness 2
                color #7B1FA2
            }
        }

        # Theme
        theme default
    }

    configuration {
        scope softwaresystem
    }

}