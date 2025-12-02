workspace "EAC Extension" "Docker-based extension providing EAC commands for R2R CLI" {

    model {
        # External systems
        r2r_cli = softwareSystem "R2R CLI" "Containerized workflow CLI" "Consumer"
        eac_commands = softwareSystem "EAC Commands" "Go commands binary" "External"
        docker_engine = softwareSystem "Docker Engine" "Container runtime" "External"
        ghcr = softwareSystem "GitHub Container Registry" "Image registry" "External"

        # Extension system
        ext_eac = softwareSystem "EAC Extension" "Docker container packaging EAC commands for R2R CLI" {

            dockerfile = container "Dockerfile" "Multi-stage build packaging EAC commands binary" "Docker" {
                build_stage = component "Build Stage" "Compiles Go commands binary" "Docker Stage"
                runtime_stage = component "Runtime Stage" "Minimal Alpine with binary" "Docker Stage"
                entrypoint = component "Entrypoint" "Executes commands with args" "Shell"
            }
        }

        # Relationships
        r2r_cli -> ext_eac "Runs as container extension" "Docker"
        ext_eac -> eac_commands "Packages commands binary"
        dockerfile -> docker_engine "Builds container image"
        dockerfile -> ghcr "Publishes to registry"

        build_stage -> eac_commands "Compiles binary"
        runtime_stage -> build_stage "Copies binary from"
    }

    views {
        systemContext ext_eac "SystemContext" {
            include *
            autoLayout lr
            title "EAC Extension - System Context"
        }

        container ext_eac "Containers" {
            include *
            autoLayout tb
            title "EAC Extension - Container"
        }

        component dockerfile "DockerfileComponents" {
            include *
            autoLayout tb
            title "Dockerfile - Build Stages"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
