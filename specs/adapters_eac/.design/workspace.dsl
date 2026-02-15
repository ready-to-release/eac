workspace "EAC Adapter" "Recursive EAC CLI invocation through native binary or container execution modes" {

    model {
        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Pipeline and release commands invoking EAC recursively" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Tool registry and executor for routing" "Dependency"

        # External systems
        eac_binary = softwareSystem "EAC Binary" "Native eac CLI binary on host" "External"
        docker_engine = softwareSystem "Docker Engine" "Container runtime for eac-ext image" "External"

        # EAC adapter system
        eac_adapter = softwareSystem "EAC Adapter" "Allows recursive EAC CLI invocation through native binary or container execution, routing via tool registry configuration" {

            # Port
            port = container "EAC Port" "Interface and types for EAC command execution" "Go Package" {
                port_interface = component "EACPort Interface" "Execute(ctx, args, cfg) returning Result" "Go"
                exec_config = component "ExecConfig" "WorkspaceRoot, ModuleRoot, OutputDir, Writers, Env" "Go"
                result_type = component "Result" "ExitCode, Stdout, Stderr, Duration, Success()" "Go"
            }

            # Factory
            factory = container "Factory" "Resolves native vs container adapter based on tool registry" "Go Package" {
                factory_impl = component "New Factory" "Checks tool registry for eac binding, routes accordingly" "Go"
            }

            # Native Adapter
            native = container "Native Adapter" "Executes EAC commands via local binary" "Go Package" {
                native_impl = component "Native Adapter" "Runs eac binary with workspace-rooted execution" "Go"
            }

            # Container Adapter
            container_exec = container "Container Adapter" "Executes EAC commands via eac-ext Docker image" "Go Package" {
                container_impl = component "Container Adapter" "Runs eac-ext container with tool executor" "Go"
            }

            # Command Executor
            command_executor = container "Command Executor" "Convenience adapter bridging to executor pattern" "Go Package" {
                executor_adapter = component "CommandExecutorAdapter" "RunCommand convenience wrapper" "Go"
            }

            # Mock
            mock = container "Mock" "In-memory mock recording calls for testing" "Go Package" {
                mock_impl = component "MockEAC" "Records Execute calls for assertion" "Go"
            }
        }

        # Consumer relationships
        eac_cli -> eac_adapter "Invokes EAC commands recursively" "Go Import"

        # Dependency relationships
        eac_adapter -> core "Uses tool registry and executor for routing" "Go Import"

        # External relationships
        native -> eac_binary "Executes native binary" "CLI"
        container_exec -> docker_engine "Runs eac-ext container" "Docker API"

        # Internal relationships
        factory -> native "Creates when no container tool binding"
        factory -> container_exec "Creates when ToolTypeContainer detected"
        command_executor -> port "Uses EACPort interface"

        # Component relationships
        factory_impl -> port_interface "Returns EACPort implementation"
        native_impl -> exec_config "Accepts execution config"
        native_impl -> result_type "Returns execution result"
        container_impl -> exec_config "Accepts execution config"
        container_impl -> result_type "Returns execution result"
    }

    views {
        systemContext eac_adapter "SystemContext" {
            include *
            autoLayout lr
            title "EAC Adapter - System Context"
            description "Shows EAC adapter with consumers and execution targets"
        }

        container eac_adapter "Containers" {
            include *
            autoLayout tb
            title "EAC Adapter - Package Architecture"
            description "Shows EAC adapter internal structure with native and container paths"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
