workspace "CLIE CLI Architecture" "Enterprise-grade containerized workflow execution CLI" {
    model {
        developer = person "Developer" "Uses CLIE CLI to execute containerized development workflows"
        cicd_system = person "CI/CD System" "Automated build and test pipelines" "Automation"

        docker_engine = softwareSystem "Docker Engine" "Container runtime and orchestration" "External"
        github_registry = softwareSystem "GitHub Container Registry" "Docker image registry" "External"
        git_repo = softwareSystem "Git Repository" "Version control system" "External"

        # Dependencies (from module contracts)
        eac_core = softwareSystem "EAC Core" "Core domain libraries for repository operations" "Dependency"

        clie_cli = softwareSystem "CLIE CLI" "Command-line interface for containerized workflow execution" {

            // Command Layer - Cobra-based command structure
            command_layer = container "Command Layer" "Cobra-based command routing and execution" "Go/Cobra" {
                root_cmd = component "Root Command" "Main entry point with logging, argument filtering, version validation" "Go"
                run_cmd = component "Run Command" "Executes extensions in containers with TTY/non-TTY support" "Go"
                interactive_cmd = component "Interactive Command" "Starts extension containers with shell access" "Go"
                install_cmd = component "Install Command" "Pulls images and manages extension installation" "Go"
                init_cmd = component "Init Command" "Creates .clie/clie.yml configuration" "Go"
                version_cmd = component "Version Command" "Displays version information" "Go"
                extension_aliases = component "Extension Aliases" "Registers extension name aliases for convenience" "Go"
                verify_cmd = component "Verify Command" "Validates extension configurations" "Go"
                validate_cmd = component "Validate Command" "Validates configuration schema" "Go"
                update_cmd = component "Update Command" "Updates the clie CLI binary to the latest version from GitHub releases" "Go"
                cleanup_cmd = component "Cleanup Command" "Removes stopped containers" "Go"
                list_cmd = component "List Command" "Lists available extensions" "Go"
                metadata_cmd = component "Metadata Command" "Extracts extension metadata" "Go"
            }

            // Configuration Management
            config_system = container "Configuration System" "Multi-layer configuration management" "Go/Viper" {
                config_loader = component "Config Loader" "Loads .clie/clie.yml configuration" "Go"
                repo_finder = component "Repository Finder" "Locates git repository root" "Go"
                extension_validator = component "Extension Validator" "Validates pinned SHA tags" "Go"
                config_merger = component "Config Merger" "Merges base + local + personal + dev configs" "Go"
            }

            // Command Parser - EBNF-based parsing
            parser_system = container "Command Parser" "EBNF-based command-line argument parser" "Go" {
                ebnf_parser = component "EBNF Parser" "Parses arguments per EBNF grammar" "Go"
                boundary_detector = component "Boundary Detector" "Identifies Viper vs container arg boundary" "Go"
                global_flag_parser = component "Global Flag Parser" "Extracts --clie-debug, --clie-quiet flags" "Go"
                extension_resolver = component "Extension Resolver" "Resolves extension names and subcommands" "Go"
            }

            // Docker Orchestration
            docker_system = container "Docker Orchestration" "Container lifecycle management" "Go/Docker SDK" {
                container_host = component "Container Host" "Manages container lifecycle operations" "Go"
                container_creator = component "Container Creator" "Creates container configs with volumes" "Go"
                container_runner = component "Container Runner" "Starts containers and manages execution" "Go"
                attach_manager = component "Attach Manager" "Handles stdin/stdout/stderr attachment" "Go"
                tty_handler = component "TTY Handler" "Manages TTY vs non-TTY modes" "Go"
                signal_handler = component "Signal Handler" "Handles SIGINT/SIGTERM for graceful shutdown" "Go"
                cleanup_manager = component "Cleanup Manager" "Cleans up child containers (Docker-in-Docker)" "Go"
                ansi_filter = component "ANSI Filter" "Filters problematic escape sequences" "Go"
                image_inspector = component "Image Inspector" "Inspects Docker images for metadata" "Go"
                volume_mapper = component "Volume Mapper" "Maps repository root to /workspace" "Go"
            }

            // Extension Management
            extension_system = container "Extension Management" "Extension installation and lifecycle" "Go" {
                installer = component "Extension Installer" "Ensures images exist locally" "Go"
                pull_policy_handler = component "Pull Policy Handler" "Handles Always/IfNotPresent/Never policies" "Go"
                registry_client = component "Registry Client" "Interacts with GitHub Container Registry" "Go"
                sha_resolver = component "SHA Resolver" "Discovers latest SHA tags for pinning" "Go"
                local_loader = component "Local Loader" "Loads local development images" "Go"
            }

            // Validation System
            validator_system = container "Validation System" "Configuration and command validation" "Go" {
                command_validator = component "Command Validator" "Validates commands against EBNF" "Go"
                schema_validator = component "Schema Validator" "Validates .clie/clie.yml against JSON schema" "Go"
                contract_validator = component "Contract Validator" "Validates against contracts" "Go"
            }

            // Logging and Monitoring
            logging_system = container "Logging System" "Structured logging and monitoring" "Go/Zerolog" {
                logger = component "Logger" "Structured logging with context" "Go"
                context_manager = component "Context Manager" "Manages command, component, operation_id context" "Go"
                level_controller = component "Level Controller" "Controls debug/info/warn/error/fatal levels" "Go"
                output_formatter = component "Output Formatter" "Formats console and file output" "Go"
            }

            // Terminal Interface
            terminal_system = container "Terminal System" "Cross-platform terminal handling" "Go" {
                terminal_detector = component "Terminal Detector" "Detects terminal capabilities" "Go"
                unix_handler = component "Unix Handler" "Unix-specific terminal operations" "Go"
                windows_handler = component "Windows Handler" "Windows-specific terminal operations" "Go"
                console_api = component "Console API" "Windows console API integration" "Go"
                spinner = component "Spinner" "Progress indicators for long operations" "Go"
            }

            // GitHub Integration
            github_system = container "GitHub Integration" "GitHub Container Registry integration" "Go" {
                registry_api = component "Registry API Client" "HTTP client for GHCR" "Go"
                tag_fetcher = component "Tag Fetcher" "Fetches available image tags" "Go"
                auth_handler = component "Auth Handler" "Handles GITHUB_TOKEN authentication" "Go"
                cache_manager = component "Cache Manager" "Caches registry responses" "Go"
            }

            // Cache System
            cache_system = container "Cache System" "Caching for registry responses and metadata" "Go" {
                registry_cache = component "Registry Cache" "Caches GitHub registry API responses" "Go"
                metadata_cache = component "Metadata Cache" "Caches extension metadata" "Go"
            }

            // Session Management
            session_system = container "Session System" "Manages CLI session state" "Go" {
                session_manager = component "Session Manager" "Tracks session state across commands" "Go"
            }

            // TUI System
            tui_system = container "TUI System" "Terminal UI components" "Go" {
                tui_spinner = component "TUI Spinner" "Progress indicators for long operations" "Go"
                progress_bar = component "Progress Bar" "Visual progress for downloads" "Go"
            }

            // Version Management
            version_system = container "Version Management" "Build version tracking" "Go" {
                version_info = component "Version Info" "Stores version, commit, timestamp" "Go"
                version_validator = component "Version Validator" "Validates version constraints" "Go"
            }

            // Environment Constants
            envconsts = container "Environment Constants" "Isolated environment variable constants for clie CLI (duplicated from core for module isolation)" "Go" {
                app_consts = component "App Constants" "CLIE_CONFIG, CLIE_FIXED_REDIRECT, CLIE_SKIP_PIN_WARNING" "Go"
                debug_consts = component "Debug Constants" "CLIE_DEBUG, CLIE_LOG_LEVEL, CLIE_VERBOSE_LOG" "Go"
                docker_consts = component "Docker Constants" "CLIE_HOST_REPOROOT, CLIE_DOCKER_MODE, CLIE_DOCKER_HOST" "Go"
                test_consts = component "Test Constants" "CLIE_TESTING, CLIE_CHECK_TAGS, CLIE_ORIGINAL_ARGS" "Go"
            }
        }

        // User Interactions
        developer -> clie_cli "Executes commands via CLI"
        developer -> command_layer "Executes CLI commands"
        cicd_system -> clie_cli "Runs automated workflows"

        // Dependency relationships (from module contracts)
        clie_cli -> eac_core "Uses repository utilities" "Go Import"

        // Command Layer relationships
        root_cmd -> run_cmd "Routes to run command"
        root_cmd -> interactive_cmd "Routes to interactive command"
        root_cmd -> install_cmd "Routes to install command"
        root_cmd -> init_cmd "Routes to init command"
        root_cmd -> version_cmd "Routes to version command"
        root_cmd -> verify_cmd "Routes to verify command"
        root_cmd -> validate_cmd "Routes to validate command"
        root_cmd -> update_cmd "Routes to update command"
        root_cmd -> cleanup_cmd "Routes to cleanup command"
        root_cmd -> list_cmd "Routes to list command"
        root_cmd -> metadata_cmd "Routes to metadata command"
        root_cmd -> logging_system "Initializes logging"
        root_cmd -> parser_system "Validates command structure"

        // Run Command Flow
        run_cmd -> config_system "Loads configuration"
        run_cmd -> parser_system "Parses arguments"
        run_cmd -> extension_system "Ensures image exists"
        run_cmd -> docker_system "Creates and runs container"
        run_cmd -> logging_system "Logs execution"

        // Interactive Command Flow
        interactive_cmd -> config_system "Loads configuration"
        interactive_cmd -> extension_system "Validates extension"
        interactive_cmd -> docker_system "Starts interactive container"

        // Install Command Flow
        install_cmd -> config_system "Reads/updates config"
        install_cmd -> github_system "Queries registry for tags"
        install_cmd -> extension_system "Installs images"

        // Configuration System relationships
        config_system -> git_repo "Finds repository root"
        config_system -> validator_system "Validates schema"

        // Parser System relationships
        parser_system -> validator_system "Validates against EBNF"

        // Docker System relationships
        docker_system -> docker_engine "Manages containers" "Docker API"
        docker_system -> logging_system "Logs operations"
        docker_system -> terminal_system "Handles terminal I/O"

        // Extension System relationships
        extension_system -> github_system "Fetches image metadata"
        extension_system -> docker_engine "Pulls images" "Docker API"
        extension_system -> config_system "Updates configuration"

        // GitHub System relationships
        github_system -> github_registry "Queries registry" "HTTPS"
        github_system -> cache_system "Caches registry responses"

        // Cache System relationships
        extension_system -> cache_system "Caches metadata"
        install_cmd -> cache_system "Uses cached registry data"

        // Session System relationships
        command_layer -> session_system "Manages session state"

        // TUI System relationships
        install_cmd -> tui_system "Shows download progress"
        run_cmd -> tui_system "Shows spinner during setup"
        docker_system -> tui_system "Shows pull progress"

        // Environment constants relationships
        config_system -> envconsts "Uses config env vars"
        docker_system -> envconsts "Uses Docker env vars"
        logging_system -> envconsts "Uses debug env vars"
        command_layer -> envconsts "Uses testing env vars"

        // Cross-cutting concerns
        command_layer -> logging_system "Logs all operations"
        docker_system -> terminal_system "Streams I/O"
    }

    views {
        systemContext clie_cli "SystemContext" {
            include *
            autoLayout
        }

        container clie_cli "Containers" {
            include *
            autoLayout lr
        }

        component config_system "ConfigurationSystem" {
            include *
            autoLayout
        }

        component parser_system "ParserSystem" {
            include *
            autoLayout
        }

        component docker_system "DockerOrchestration" {
            include *
            autoLayout
        }

        component extension_system "ExtensionManagement" {
            include *
            autoLayout
        }

        component validator_system "ValidationSystem" {
            include *
            autoLayout
        }

        component logging_system "LoggingSystem" {
            include *
            autoLayout
        }

        component terminal_system "TerminalSystem" {
            include *
            autoLayout
        }

        component github_system "GitHubIntegration" {
            include *
            autoLayout
        }

        component cache_system "CacheSystem" {
            include *
            autoLayout
        }

        component session_system "SessionSystem" {
            include *
            autoLayout
        }

        component tui_system "TUISystem" {
            include *
            autoLayout
        }

        dynamic clie_cli "RunWorkflow" "Container execution workflow (clie run pwsh script.ps1)" {
            developer -> command_layer "1. Execute: clie run pwsh script.ps1"
            command_layer -> parser_system "2. Parse arguments"
            command_layer -> config_system "3. Load configuration"
            command_layer -> extension_system "4. Ensure image exists"
            extension_system -> github_system "5. Check for updates"
            github_system -> github_registry "6. Query registry"
            extension_system -> docker_engine "7. Pull image if needed"
            command_layer -> docker_system "8. Create container"
            docker_system -> docker_engine "9. Start container"
            docker_system -> terminal_system "10. Attach I/O"
            docker_system -> docker_engine "11. Wait for completion"
            docker_system -> command_layer "12. Return exit code"
            autoLayout
        }

        dynamic clie_cli "InstallWorkflow" "Extension installation workflow (clie install pwsh)" {
            developer -> command_layer "1. Execute: clie install pwsh"
            command_layer -> github_system "2. Query registry for latest SHA"
            github_system -> github_registry "3. Fetch available tags"
            command_layer -> config_system "4. Update config with SHA pin"
            command_layer -> extension_system "5. Install image"
            extension_system -> docker_engine "6. Pull image"
            extension_system -> command_layer "7. Report success"
            autoLayout
        }

        theme default
    }
}
