# docker-eac

The `docker-eac` module wraps the Docker SDK to implement `ContainerPort` for container lifecycle management.

## System Context

Shows how the Docker adapter provides container operations to CLI commands.

<!-- structurizr:adapters:docker:SystemContext -->

## Container Architecture

High-level view of the Docker adapter packages.

<!-- structurizr:adapters:docker:Containers -->

## Component Architecture

### Container Adapter Components

ContainerPort implementation with retry logic and exponential backoff.

<!-- structurizr:adapters:docker:ContainerAdapterComponents -->

### Docker Client Components

Abstracted Docker SDK client with real, mock, and factory variants.

<!-- structurizr:adapters:docker:DockerClientComponents -->

### Serve Components

Long-running container management with port reservation and browser launching.

<!-- structurizr:adapters:docker:ServeComponents -->

## Design File

- **Location**: `specs/docker-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module docker-eac`

## Key Responsibilities

| Package   | Purpose                                              |
| --------- | ---------------------------------------------------- |
| client    | Docker SDK abstraction with testable interface        |
| container | ContainerPort adapter with retry and mount mapping    |
| serve     | Long-running container lifecycle with port management |
| scan      | OWASP ZAP DAST scanning via Docker                   |
| dind      | Docker-in-Docker path translation for CI              |
