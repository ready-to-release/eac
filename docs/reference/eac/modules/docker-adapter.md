# docker-adapter

The `docker-adapter` module wraps the Docker SDK to implement `ContainerPort` for container lifecycle management.

## System Context

Shows how the Docker adapter provides container operations to CLI commands.

<!-- structurizr:docker-adapter:SystemContext -->

## Container Architecture

High-level view of the Docker adapter packages.

<!-- structurizr:docker-adapter:Containers -->

## Component Architecture

### Container Adapter Components

ContainerPort implementation with retry logic and exponential backoff.

<!-- structurizr:docker-adapter:ContainerAdapterComponents -->

### Docker Client Components

Abstracted Docker SDK client with real, mock, and factory variants.

<!-- structurizr:docker-adapter:DockerClientComponents -->

### Serve Components

Long-running container management with port reservation and browser launching.

<!-- structurizr:docker-adapter:ServeComponents -->

## Design File

- **Location**: `specs/docker-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module docker-adapter`

## Key Responsibilities

| Package   | Purpose                                              |
| --------- | ---------------------------------------------------- |
| client    | Docker SDK abstraction with testable interface        |
| container | ContainerPort adapter with retry and mount mapping    |
| serve     | Long-running container lifecycle with port management |
| scan      | OWASP ZAP DAST scanning via Docker                   |
| dind      | Docker-in-Docker path translation for CI              |
