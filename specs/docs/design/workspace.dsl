workspace "Docs Module Architecture" "MkDocs documentation container management for the EAC framework" {

    model {
        # External actors and systems
        developer = person "Developer" "Developer viewing and editing documentation"
        docker_daemon = softwareSystem "Docker Daemon" "Local Docker daemon for running MkDocs containers" "External"
        web_browser = softwareSystem "Web Browser" "Browser to view rendered documentation" "External"

        # Main docs system
        docs_system = softwareSystem "Docs Module" "MkDocs documentation container orchestration and management" {

            # Core containers
            container_manager = container "Container Manager" "Manages MkDocs Docker container lifecycle: build image, start/stop server, handle ports, mount volumes, stream logs, and cleanup containers." "Go (Docker SDK)" "Core" {
                imageBuilder = component "Image Builder" "Builds cli-mkdocs Docker image from Dockerfile" "Go (Docker SDK)"
                containerOrchestrator = component "Container Orchestrator" "Starts, stops, and removes MkDocs containers" "Go (Docker SDK)"
                volumeManager = component "Volume Manager" "Mounts repository docs/ directory into container" "Go (Docker SDK)"
                portMapper = component "Port Mapper" "Maps container port 8000 to host port" "Go (Docker SDK)"
                logStreamer = component "Log Streamer" "Streams container stdout/stderr to console" "Go (Docker SDK)"
            }

            browser_launcher = container "Browser Launcher" "Automatically opens web browser to MkDocs server URL after container starts successfully." "Go (os/exec)" "Integration" {
                urlGenerator = component "URL Generator" "Generates localhost URL with configured port" "Go"
                browserOpener = component "Browser Opener" "Launches system default browser with URL" "Go (os/exec)"
            }

            docker_client = container "Docker Client" "Provides Docker API client for container operations, handles authentication, manages API connections, and error handling." "Go (Docker SDK)" "Infrastructure" {
                apiClient = component "API Client" "Docker daemon API client instance" "Go (github.com/docker/docker/client)"
                connectionManager = component "Connection Manager" "Manages Docker daemon connection lifecycle" "Go"
            }

            repository_locator = container "Repository Locator" "Finds repository root directory by walking up directory tree, validates structure, and provides absolute paths for volume mounts." "Go (filepath)" "Core" {
                pathWalker = component "Path Walker" "Walks up directory tree to find root" "Go (filepath)"
                structureValidator = component "Structure Validator" "Validates src/ directory exists at root" "Go (os)"
            }

            mkdocs_container = container "MkDocs Container" "Docker container running MkDocs serve with Material theme, watches for file changes, and serves documentation on port 8000." "Docker (Python/MkDocs)" "Runtime"

        }

        # User interactions
        developer -> docs_system "Runs docs serve command" "CLI"
        developer -> web_browser "Views documentation" "HTTP"

        # External system relationships
        docs_system -> docker_daemon "Manages MkDocs containers" "Docker API"
        docker_daemon -> mkdocs_container "Runs MkDocs server" "Docker runtime"
        browser_launcher -> web_browser "Opens documentation URL" "OS default browser"

        # Internal relationships

        # Container Manager relationships
        container_manager -> docker_client "Uses Docker API for operations" "Function calls"
        container_manager -> repository_locator "Gets repository root for volume mounts" "Function calls"
        container_manager -> mkdocs_container "Manages container lifecycle" "Docker API"

        # Browser Launcher relationships
        browser_launcher -> container_manager "Waits for container startup" "Function calls"

        # Repository Locator relationships
        repository_locator -> container_manager "Provides absolute paths" "Function calls"

        # Component relationships

        # Container Manager components
        imageBuilder -> apiClient "Builds Docker image from Dockerfile" "Docker SDK calls"
        volumeManager -> pathWalker "Gets repository root path" "Function calls"
        containerOrchestrator -> portMapper "Configures port mappings" "Function calls"
        containerOrchestrator -> volumeManager "Configures volume mounts" "Function calls"
        containerOrchestrator -> apiClient "Creates and starts containers" "Docker SDK calls"
        logStreamer -> apiClient "Attaches to container logs" "Docker SDK calls"

        # Browser Launcher components
        urlGenerator -> browserOpener "Provides URL to open" "Function calls"

        # Docker Client components
        connectionManager -> apiClient "Provides configured client" "Function calls"

        # Repository Locator components
        pathWalker -> structureValidator "Validates found directory" "Function calls"
    }

    views {
        # System Context view
        systemContext docs_system "SystemContext" {
            include *
            autoLayout lr
            title "Docs Module - System Context"
            description "Shows the Docs module in context with external systems and users"
        }

        # Container view
        container docs_system "Containers" {
            include *
            autoLayout tb
            title "Docs Module - Container Architecture"
            description "Shows the 5 core containers and runtime container"
        }

        # Component views for key containers

        component container_manager "ContainerManagerComponents" {
            include *
            autoLayout tb
            title "Container Manager - Components"
            description "Docker container lifecycle management components"
        }

        component browser_launcher "BrowserComponents" {
            include *
            autoLayout tb
            title "Browser Launcher - Components"
            description "Automatic browser opening components"
        }

        component docker_client "ClientComponents" {
            include *
            autoLayout tb
            title "Docker Client - Components"
            description "Docker API client management components"
        }

        component repository_locator "LocatorComponents" {
            include *
            autoLayout tb
            title "Repository Locator - Components"
            description "Repository root discovery components"
        }

        # Filtered views

        container docs_system "CoreContainers" {
            include ->container_manager->
            include ->repository_locator->
            include ->docker_client->
            autoLayout lr
            title "Core Containers"
            description "Core containers for Docker orchestration"
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
            element "Runtime" {
                background #1976D2
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
            relationship "HTTP" {
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
