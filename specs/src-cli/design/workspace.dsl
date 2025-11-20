workspace "src-cli Architecture" "CLI application with AI-driven architecture documentation generation" {
    model {
        developer = person "Developer" "Uses CLI to generate and validate architecture"
        anthropic_api = softwareSystem "Anthropic API" "Provides LLM for architecture generation" "External"
        docker_engine = softwareSystem "Docker Engine" "Runs validation and viewer containers" "External"
        file_system = softwareSystem "File System" "Stores workspace DSL and configuration" "External"
        git_repo = softwareSystem "Git Repository" "Manages version control and module paths" "External"

        cli = softwareSystem "src-cli System" "CLI tool for architecture design generation and validation" {
            command_router = container "Command Router" "Routes design subcommands" "Go"
            
            create_service = container "Design Create" "Generates Structurizr DSL using AI" "Go" {
                config_parser = component "Config Parser" "Parses CLI arguments into DesignConfig" "Go"
                module_validator = component "Module Validator" "Verifies module exists in specs directory" "Go"
                docker_checker = component "Docker Checker" "Verifies Docker availability" "Go"
                prompt_builder = component "Prompt Builder" "Constructs AI prompt with contracts" "Go"
                ai_generator = component "AI Generator" "Calls Anthropic API for DSL generation" "Go"
                composite_validator = component "Composite Validator" "Runs quick and full validation phases" "Go"
                output_writer = component "Output Writer" "Writes DSL to workspace.dsl file" "Go"
            }
            
            validate_service = container "Design Validate" "Validates workspace DSL files" "Go" {
                args_parser = component "Arguments Parser" "Extracts module and command flags" "Go"
                validator_factory = component "Validator Factory" "Creates StructurizrValidator instance" "Go"
                docker_executor = component "Docker Executor" "Runs Docker validation command" "Go"
                output_parser = component "Output Parser" "Parses Docker CLI validation results" "Go"
                formatter = component "Validation Formatter" "Formats results for console and JSON" "Go"
                json_writer = component "JSON Writer" "Writes validation results to file" "Go"
            }
            
            serve_service = container "Design Serve" "Launches interactive Structurizr Lite viewer" "Go" {
                workspace_locator = component "Workspace Locator" "Finds workspace DSL file" "Go"
                container_manager = component "Container Manager" "Manages Structurizr Lite container lifecycle" "Go"
                browser_launcher = component "Browser Launcher" "Opens browser to viewer URL" "Go"
            }
            
            contract_system = container "Contract System" "Manages architectural contracts and prompts" "Go" {
                contract_loader = component "Contract Loader" "Loads design contracts from disk" "Go"
                anti_corruption = component "Anti-Corruption Rules" "Validates AI output against rules" "Go"
                retry_handler = component "Retry Handler" "Manages generation retry logic" "Go"
            }
            
            repository_system = container "Repository System" "Detects and resolves module paths" "Go" {
                root_detector = component "Root Detector" "Locates Git repository root" "Go"
                path_resolver = component "Path Resolver" "Maps module names to filesystem paths" "Go"
            }
            
            ai_system = container "AI System" "Manages LLM provider integration" "Go" {
                executor = component "AI Executor" "Communicates with LLM provider" "Go"
                provider_registry = component "Provider Registry" "Manages pluggable AI providers" "Go"
            }
        }

        developer -> command_router "Executes CLI commands"
        
        command_router -> create_service "Routes to design create"
        command_router -> validate_service "Routes to design validate"
        command_router -> serve_service "Routes to design serve"
        
        create_service -> contract_system "Loads contracts and builds prompt"
        create_service -> repository_system "Resolves module paths"
        create_service -> ai_system "Requests architecture generation"
        create_service -> validate_service "Validates generated DSL"
        create_service -> file_system "Writes workspace.dsl output"
        
        validate_service -> repository_system "Resolves workspace file location"
        validate_service -> docker_engine "Runs validation in container" "Docker API"
        validate_service -> file_system "Writes validation results"
        
        serve_service -> repository_system "Locates workspace directory"
        serve_service -> docker_engine "Manages viewer container" "Docker API"
        serve_service -> file_system "Reads workspace.dsl"
        
        contract_system -> file_system "Reads contract files and templates"
        
        repository_system -> git_repo "Detects repository boundaries"
        repository_system -> file_system "Resolves module paths"
        
        ai_system -> anthropic_api "Sends generation requests" "HTTPS"
    }

    views {
        systemContext cli "SystemContext" {
            include *
            autoLayout
        }

        container cli "Containers" {
            include *
            autoLayout
        }

        component create_service "CreateService" {
            include *
            autoLayout
        }

        component validate_service "ValidateService" {
            include *
            autoLayout
        }

        component serve_service "ServeService" {
            include *
            autoLayout
        }

        component contract_system "ContractSystem" {
            include *
            autoLayout
        }

        component ai_system "AISystem" {
            include *
            autoLayout
        }

        styles {
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Person" {
                background #08427b
                color #ffffff
                shape box
            }
        }
    }
}