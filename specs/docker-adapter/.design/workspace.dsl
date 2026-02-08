workspace "Docker Adapter" "Docker SDK wrapper implementing ContainerPort for container lifecycle management" {

    model {
        # External actors
        developer = person "Developer" "Runs commands that use Docker containers"

        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Build, test, scan, serve commands using containers" "Dependent"
        clibase = softwareSystem "Clibase" "Orchestrator dispatching container work units" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Tool registry and executor" "Dependency"
        container_runtime_contract = softwareSystem "Container Runtime Contract" "ContainerPort interface definition" "Dependency"

        # External systems
        docker_engine = softwareSystem "Docker Engine" "Container runtime daemon" "External"
        ghcr = softwareSystem "GitHub Container Registry" "Docker image registry" "External"

        # Docker adapter system
        docker_adapter = softwareSystem "Docker Adapter" "Wraps Docker SDK to implement ContainerPort for builds, tests, scans, and serve operations" {

            # Container Adapter
            container_adapter = container "Container Adapter" "Implements ContainerPort with retry logic and exponential backoff" "Go Package" {
                adapter_impl = component "ContainerAdapter" "Execute, Build, Pull, ImageExists, IsAvailable, Close" "Go"
                retry_logic = component "Retry Logic" "Exponential backoff (3 retries, 100ms-2s) for transient errors" "Go"
                mount_mapper = component "Mount Mapper" "Maps host paths to container volumes" "Go"
                env_builder = component "Env Builder" "Constructs container environment variables" "Go"
            }

            # Docker Client
            docker_client = container "Docker Client" "Abstract Docker SDK operations behind testable interface" "Go Package" {
                client_interface = component "DockerClient Interface" "13-method abstraction over Docker SDK" "Go"
                real_client = component "RealDockerClient" "Production wrapper around docker/client.Client" "Go"
                mock_client = component "SimpleMockDockerClient" "BDD-friendly mock with JSON state persistence" "Go"
                client_factory = component "Client Factory" "Selects real or mock based on environment" "Go"
            }

            # Serve
            serve = container "Serve" "Long-running container management for docs and design servers" "Go Package" {
                serve_manager = component "Serve Manager" "Start, stop, list long-running containers" "Go"
                port_manager = component "Port Manager" "Atomic port reservation with TTL cleanup" "Go"
                browser_launcher = component "Browser Launcher" "Platform-aware browser opening" "Go"
            }

            # Run
            run = container "Run" "One-shot container execution" "Go Package" {
                run_container = component "RunContainer" "Single container execution with output capture" "Go"
            }

            # Image Operations
            image_ops = container "Image Operations" "Parallel image ensure with concurrency control" "Go Package" {
                image_ensure = component "Image Ensure" "Parallel build-or-pull with semaphore limiting" "Go"
            }

            # Scan
            scan = container "Security Scanning" "OWASP ZAP DAST scanning via Docker containers" "Go Package" {
                zap_scanner = component "ZAP Scanner" "OWASP ZAP container execution and result parsing" "Go"
            }

            # DinD Utilities
            dind = container "DinD Utilities" "Docker-in-Docker path translation for CI" "Go Package" {
                path_translator = component "Path Translator" "Translates host paths for nested Docker" "Go"
            }

            # Global
            global = container "Global Singleton" "Lazy-initialized global container adapter" "Go Package" {
                global_instance = component "GlobalContainer" "Lazy singleton with sync.Once initialization" "Go"
            }
        }

        # Actor relationships
        developer -> docker_adapter "Uses via CLI commands" "CLI"

        # Consumer relationships
        eac_cli -> docker_adapter "Runs containers for build, test, scan, serve" "Go Import"
        clibase -> docker_adapter "Dispatches container work units" "Go Import"

        # Dependency relationships
        docker_adapter -> core "Uses tool registry and executor" "Go Import"
        docker_adapter -> container_runtime_contract "Implements ContainerPort" "Go Import"

        # External relationships
        docker_client -> docker_engine "Manages containers and images" "Docker API"
        image_ops -> docker_engine "Pulls and builds images" "Docker API"
        serve -> docker_engine "Manages long-running containers" "Docker API"
        image_ops -> ghcr "Pulls container images" "HTTPS"

        # Internal relationships
        container_adapter -> docker_client "Uses client for Docker operations"
        container_adapter -> dind "Translates paths in DinD mode"
        serve -> docker_client "Uses client for serve operations"
        run -> docker_client "Uses client for one-shot execution"
        image_ops -> docker_client "Uses client for image operations"
        scan -> docker_client "Uses client for ZAP containers"
        global -> container_adapter "Creates singleton adapter"

        # Component relationships
        adapter_impl -> retry_logic "Retries transient failures"
        adapter_impl -> mount_mapper "Maps volume mounts"
        adapter_impl -> env_builder "Builds environment"
        client_factory -> real_client "Creates for production"
        client_factory -> mock_client "Creates for testing"
        serve_manager -> port_manager "Reserves ports"
        serve_manager -> browser_launcher "Opens browser after start"
    }

    views {
        systemContext docker_adapter "SystemContext" {
            include *
            autoLayout lr
            title "Docker Adapter - System Context"
            description "Shows Docker adapter with consumers, dependencies, and external systems"
        }

        container docker_adapter "Containers" {
            include *
            autoLayout tb
            title "Docker Adapter - Package Architecture"
            description "Shows Docker adapter internal structure"
        }

        component container_adapter "ContainerAdapterComponents" {
            include *
            autoLayout tb
            title "Container Adapter - Components"
            description "ContainerPort implementation with retry logic"
        }

        component docker_client "DockerClientComponents" {
            include *
            autoLayout tb
            title "Docker Client - Components"
            description "Abstracted Docker SDK client"
        }

        component serve "ServeComponents" {
            include *
            autoLayout tb
            title "Serve - Components"
            description "Long-running container management"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
