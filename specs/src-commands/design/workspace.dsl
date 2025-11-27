workspace "src-commands Architecture" "CLI command module with design, test, pipeline, build, and commit orchestration" {
    model {
        user = person "Developer" "Uses CLI commands to manage modules, tests, and deployments"
        
        github_api = softwareSystem "GitHub API" "GitHub Actions and repository operations" "External"
        anthropic = softwareSystem "Anthropic Claude" "AI-powered generation service" "External"
        docker = softwareSystem "Docker Daemon" "Container runtime for validation and viewers" "External"
        git = softwareSystem "Git Repository" "Local version control" "External"

        src_commands = softwareSystem "src-commands Module" "CLI command framework with 19 command groups" {
            registry = container "Command Registry" "Self-registering command discovery and routing" "Go"
            orchestrator = container "Orchestrator" "Parallel execution framework for CPU-bound tasks" "Go"
            render = container "Render Engine" "Multi-format output (markdown, JSON, TOML)" "Go"
            template = container "Template Engine" "Variable substitution and template processing" "Go"

            design_cmd = container "Design Command" "Architecture documentation via Structurizr DSL" "Go" {
                create = component "Create" "Generate workspace.dsl with AI analysis" "Go"
                update = component "Update" "Update existing workspace.dsl" "Go"
                validate = component "Validate" "Validate DSL syntax via Docker" "Go"
                serve = component "Serve" "Launch Structurizr Lite viewer in Docker" "Go"
                validator = component "Validator" "Structurizr CLI integration" "Go"
                security = component "Security" "Module name validation and sanitization" "Go"
            }

            test_cmd = container "Test Command" "Module testing with Gherkin/Cucumber" "Go" {
                test_runner = component "Test Runner" "Execute go test with JSON output" "Go"
                cucumber = component "Cucumber Parser" "Parse Gherkin features and results" "Go"
                reporter = component "Reporter" "Collect and format test results" "Go"
                suite = component "Suite Selector" "Filter tests by suite name" "Go"
            }

            pipeline_cmd = container "Pipeline Command" "CI/CD orchestration with dependency order" "Go" {
                runner = component "Pipeline Runner" "Execute module pipelines" "Go"
                github_runner = component "GitHub Runner" "GitHub Actions integration" "Go"
                dependency_order = component "Dependency Resolver" "Order modules by dependencies" "Go"
                status = component "Status Checker" "Query GitHub Actions API" "Go"
            }

            build_cmd = container "Build Command" "Parallel module compilation" "Go" {
                build_runner = component "Build Runner" "Execute build commands" "Go"
                collector = component "Result Collector" "Aggregate build results" "Go"
            }

            commit_cmd = container "Commit Command" "AI-powered git operations" "Go" {
                message_gen = component "Message Generator" "Generate AI commit messages" "Go"
                validator = component "Validator" "Validate commit structure" "Go"
                verifier = component "Verifier" "Verify module contracts" "Go"
                contract_loader = component "Contract Loader" "Load module contracts" "Go"
            }

            other_cmds = container "Other Commands" "Show, Get, Specs, Templates, Help, Init, Describe" "Go"
        }

        core = softwareSystem "Core Libraries" "Shared dependencies from src/core" "Internal" {
            contracts = component "Contracts" "Module contract definitions"
            repository = component "Repository" "Git operations and module discovery"
            deps_resolver = component "Dependency Resolver" "Module dependency graph"
            system_deps = component "System Dependencies" "Docker, Git availability checks"
            ai = component "AI Client" "Anthropic API wrapper"
        }

        user -> src_commands "Invokes commands" "CLI"
        
        registry -> design_cmd "Routes to" ""
        registry -> test_cmd "Routes to" ""
        registry -> pipeline_cmd "Routes to" ""
        registry -> build_cmd "Routes to" ""
        registry -> commit_cmd "Routes to" ""
        registry -> other_cmds "Routes to" ""

        design_cmd -> orchestrator "Launches parallel tasks"
        design_cmd -> render "Formats output"
        design_cmd -> docker "Validates and serves" "Docker API"
        design_cmd -> anthropic "Generates architecture" "HTTPS/JSON"
        create -> validator "Validates DSL"
        create -> security "Validates module names"
        serve -> docker "Starts container"

        test_cmd -> orchestrator "Parallel test execution"
        test_cmd -> render "Renders test reports"
        test_cmd -> git "Detects changes" "Git API"
        test_runner -> git "Executes tests"
        cucumber -> render "Formats results"

        pipeline_cmd -> orchestrator "Executes pipelines"
        pipeline_cmd -> git "Checks module changes"
        pipeline_cmd -> docker "Runs validations"
        runner -> orchestrator "Manages execution"
        github_runner -> github_api "Queries status"
        dependency_order -> git "Loads dependencies"

        build_cmd -> orchestrator "Parallel builds"
        build_cmd -> render "Formats output"
        build_runner -> git "Executes build commands"

        commit_cmd -> anthropic "Generates messages" "HTTPS/JSON"
        commit_cmd -> render "Formats output"
        commit_cmd -> git "Gets staged changes"
        message_gen -> anthropic "Calls API"
        verifier -> git "Loads contracts"
        contract_loader -> git "Reads files"

        src_commands -> core "Uses" ""
        contracts -> repository "Provides"
        repository -> git "Reads repository"
        deps_resolver -> repository "Loads graph"
        system_deps -> docker "Checks availability"
        system_deps -> git "Checks availability"
        ai -> anthropic "Calls API"

        design_cmd -> contracts "Loads module contracts"
        test_cmd -> contracts "Loads module contracts"
        pipeline_cmd -> contracts "Loads module contracts"
        build_cmd -> contracts "Loads module contracts"
        commit_cmd -> contracts "Loads module contracts"
    }

    views {
        systemContext src_commands "SystemContext" {
            include *
            autoLayout
        }

        container src_commands "Containers" {
            include *
            autoLayout
        }

        component design_cmd "DesignCommand" {
            include *
            autoLayout
        }

        component test_cmd "TestCommand" {
            include *
            autoLayout
        }

        component pipeline_cmd "PipelineCommand" {
            include *
            autoLayout
        }

        component build_cmd "BuildCommand" {
            include *
            autoLayout
        }

        component commit_cmd "CommitCommand" {
            include *
            autoLayout
        }

        styles {
            element "Person" {
                background #08427b
                color #ffffff
                shape box
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Internal" {
                background #438dd5
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbd7
                color #000000
            }
        }
    }
}