workspace "R2R CLI Architecture" "Ready to Release CLI - Enterprise-grade automation framework with containerized extensions" {

    model {
        # External actors and systems
        user = person "Developer" "Developer using the R2R CLI to run containerized extensions and manage workflows"
        ghcr = softwareSystem "GitHub Container Registry" "Hosts Docker container images for CLI extensions" "External"
        docker_daemon = softwareSystem "Docker Daemon" "Local Docker daemon that runs extension containers" "External"

        # Main CLI system
        r2r_cli = softwareSystem "R2R CLI System" "Enterprise-grade automation framework with containerized extensions" {
            # Core containers

            cli_application = container "CLI Application" "Main entry point and command router. Parses CLI arguments, routes commands to handlers, coordinates operations across containers, and manages execution flow." "Go (Cobra)" "Core" {
                rootCmd = component "Root Command" "Cobra root command with global flags and subcommand registration" "Go"
                commandHandlers = component "Command Handlers" "Handlers for run, install, list, config commands" "Go"
                errorHandler = component "Error Handler" "Centralized error handling and user-friendly messages" "Go"
            }

            configuration_manager = container "Configuration Manager" "Loads base and override YAML configs, validates schema, checks pinned versions for CI enforcement, merges configuration layers, and provides config to other containers." "Go (Viper)" "Core" {
                configLoader = component "Config Loader" "Loads r2r.yaml and override files using Viper" "Go"
                configMerger = component "Config Merger" "Merges base and override configurations" "Go"
                pinnedVersionChecker = component "Pinned Version Checker" "Enforces version pinning in CI environments" "Go"
            }

            docker_orchestrator = container "Docker Orchestrator" "Manages complete container lifecycle: creation, execution, volume mounting (workspace, caches), network configuration, cleanup, resource limits (CPU, memory), progress tracking, and output streaming." "Go (Docker SDK)" "Core" {
                containerManager = component "Container Manager" "Creates, starts, stops, and removes containers" "Go (Docker SDK)"
                volumeManager = component "Volume Manager" "Manages workspace and cache volume mounts" "Go (Docker SDK)"
                streamHandler = component "Stream Handler" "Streams container output to user console" "Go (Docker SDK)"
                resourceLimiter = component "Resource Limiter" "Applies CPU and memory limits to containers" "Go (Docker SDK)"
            }

            extension_manager = container "Extension Manager" "Manages extension installation from GHCR, metadata retrieval, local extension development (load_local mode), and versioning/updates." "Go" "Core" {
                installer = component "Extension Installer" "Pulls extension images from registries" "Go"
                localLoader = component "Local Extension Loader" "Supports local extension development mode" "Go"
                versionResolver = component "Version Resolver" "Resolves extension versions (latest, tags, local)" "Go"
            }

            github_registry_client = container "GitHub Registry Client" "Provides GHCR API integration, tag listing, latest version detection, authentication for private registries, and extension metadata fetching." "Go (HTTP)" "Integration" {
                apiClient = component "GHCR API Client" "HTTP client for GitHub Container Registry API" "Go (net/http)"
                authenticator = component "Registry Authenticator" "Handles GHCR authentication tokens" "Go"
                tagResolver = component "Tag Resolver" "Lists and resolves extension tags (sha- pattern for latest)" "Go"
            }

            validation_engine = container "Validation Engine" "Validates YAML config structure, extension definitions (required fields, formats), Docker image references, environment variables, resource limits, and port/volume mappings." "Go" "Core"

            logger = container "Logger" "Provides structured JSON logging, log levels (debug/info/warn/error/fatal), context propagation (command, operation ID), and console/file output with field-based metadata." "Go (Zerolog)" "Infrastructure"

            registry_cache = container "Registry Cache" "Caches extension version info, available tags, latest stable versions (sha- tags), with timestamp-based expiration (configurable TTL, default 5 minutes)." "Go (JSON file)" "Infrastructure" {
                cacheStore = component "Cache Store" "Persists cache to JSON file on disk" "Go (JSON)"
                ttlManager = component "TTL Manager" "Manages cache expiration and cleanup" "Go"
            }

            command_parser = container "Command Parser" "Handles subcommand detection, extension name extraction, Viper flags vs container arguments separation, bash redirect pollution filtering, and argument boundary detection." "Go" "Core" {
                argBoundaryDetector = component "Argument Boundary Detector" "Detects where CLI args end and container args begin" "Go"
                flagFilter = component "Flag Filter" "Filters out Viper flags from container arguments" "Go"
            }

            # Documentation
            !docs docs
            !decisions decisions
        }

        # User interactions
        user -> r2r_cli "Runs CLI commands" "CLI"

        # External system relationships
        r2r_cli -> docker_daemon "Creates and manages extension containers" "Docker API"
        github_registry_client -> ghcr "Fetches extension images and metadata" "HTTPS/REST"
        docker_orchestrator -> docker_daemon "Manages container lifecycle" "Docker SDK"

        # CLI Application relationships (orchestration layer)
        cli_application -> command_parser "Parses command-line arguments and detects boundaries" "Function calls"
        cli_application -> configuration_manager "Loads and retrieves configuration" "Function calls"
        cli_application -> logger "Logs commands, operations, and errors" "Function calls"
        cli_application -> docker_orchestrator "Executes extensions in containers" "Function calls"
        cli_application -> extension_manager "Installs and manages extensions" "Function calls"

        # Configuration Manager relationships
        configuration_manager -> validation_engine "Validates configuration schema and values" "Function calls"
        configuration_manager -> registry_cache "Checks for cached pinned versions" "Function calls"
        configuration_manager -> github_registry_client "Fetches latest extension versions for validation" "Function calls"
        configuration_manager -> logger "Logs configuration loading and validation" "Function calls"

        # Extension Manager relationships
        extension_manager -> github_registry_client "Downloads extension container images" "Function calls"
        extension_manager -> registry_cache "Updates cache with extension metadata" "Function calls"
        extension_manager -> logger "Logs installation and management operations" "Function calls"
        extension_manager -> validation_engine "Validates extension metadata" "Function calls"

        # Docker Orchestrator relationships
        docker_orchestrator -> logger "Logs container operations and output" "Function calls"
        docker_orchestrator -> configuration_manager "Retrieves extension configuration" "Function calls"

        # GitHub Registry Client relationships
        github_registry_client -> logger "Logs registry API calls and responses" "Function calls"
        github_registry_client -> registry_cache "Reads and writes tag information" "Function calls"

        # Validation Engine relationships
        validation_engine -> logger "Logs validation errors and warnings" "Function calls"

        # Command Parser relationships
        command_parser -> logger "Logs parsing operations and detected boundaries" "Function calls"

        # Component-level relationships

        # CLI Application components
        rootCmd -> commandHandlers "Routes commands to handlers" "Go function calls"
        commandHandlers -> errorHandler "Handles command errors" "Go function calls"

        # Configuration Manager components
        configLoader -> configMerger "Passes loaded configs for merging" "Go function calls"
        configMerger -> pinnedVersionChecker "Validates pinned versions in merged config" "Go function calls"

        # Docker Orchestrator components
        containerManager -> volumeManager "Sets up volumes before container creation" "Go function calls"
        containerManager -> resourceLimiter "Applies resource limits to containers" "Go function calls"
        containerManager -> streamHandler "Streams container output" "Go function calls"

        # Extension Manager components
        versionResolver -> installer "Resolves version before installation" "Go function calls"
        versionResolver -> localLoader "Loads local extensions in dev mode" "Go function calls"

        # GitHub Registry Client components
        authenticator -> apiClient "Provides auth tokens for API calls" "Go function calls"
        apiClient -> tagResolver "Fetches tags for resolution" "Go function calls"

        # Registry Cache components
        cacheStore -> ttlManager "Checks TTL before reading cache" "Go function calls"

        # Command Parser components
        argBoundaryDetector -> flagFilter "Identifies flags to filter" "Go function calls"
    }

    views {
        # System Context view - shows external systems and actors
        systemContext r2r_cli "SystemContext" {
            include *
            autoLayout lr
            title "R2R CLI - System Context"
            description "Shows the R2R CLI system in context with external systems and users"
        }

        # Container view - shows all containers within the CLI system
        container r2r_cli "Containers" {
            include *
            autoLayout lr
            title "R2R CLI - Container Architecture"
            description "Shows the 9 core containers and their relationships"
        }

        # Component views for key containers

        component cli_application "CLIComponents" {
            include *
            autoLayout tb
            title "CLI Application - Components"
            description "Command routing and orchestration components"
        }

        component configuration_manager "ConfigurationComponents" {
            include *
            autoLayout tb
            title "Configuration Manager - Components"
            description "Configuration loading, merging, and validation components"
        }

        component docker_orchestrator "DockerComponents" {
            include *
            autoLayout tb
            title "Docker Orchestrator - Components"
            description "Container lifecycle management components"
        }

        component extension_manager "ExtensionComponents" {
            include *
            autoLayout tb
            title "Extension Manager - Components"
            description "Extension installation and management components"
        }

        component github_registry_client "RegistryComponents" {
            include *
            autoLayout tb
            title "GitHub Registry Client - Components"
            description "GHCR API integration components"
        }

        component registry_cache "CacheComponents" {
            include *
            autoLayout tb
            title "Registry Cache - Components"
            description "Cache storage and TTL management components"
        }

        component command_parser "ParserComponents" {
            include *
            autoLayout tb
            title "Command Parser - Components"
            description "Argument parsing and boundary detection components"
        }

        # Filtered views for specific concerns

        container r2r_cli "CoreContainers" {
            include ->cli_application->
            include ->configuration_manager->
            include ->docker_orchestrator->
            include ->command_parser->
            autoLayout lr
            title "Core Containers"
            description "Core containers responsible for command execution flow"
        }

        container r2r_cli "ExtensionManagement" {
            include ->extension_manager->
            include ->github_registry_client->
            include ->registry_cache->
            include ->validation_engine->
            autoLayout lr
            title "Extension Management"
            description "Containers responsible for extension installation and management"
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
            element "Infrastructure" {
                background #F57C00
                color #ffffff
            }
            element "Integration" {
                background #5E35B1
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
