# ext-eac

The `ext-eac` module defines the Docker extension image that packages the EAC tooling for use with r2r-cli. It provides a containerized environment with all dependencies pre-installed.

## System Context

Shows how ext-eac packages EAC tooling for containerized execution.

<!-- structurizr:ext-eac:SystemContext -->

## Container Architecture

High-level view of the extension image structure.

<!-- structurizr:ext-eac:Containers -->

## Dockerfile Components

The layers and components within the Docker image.

<!-- structurizr:ext-eac:DockerfileComponents -->

## Design File

- **Location**: `specs/ext-eac/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module ext-eac`

## Image Contents

The ext-eac Docker image includes:

| Component | Purpose |
|-----------|---------|
| Go Runtime | For executing eac-commands |
| Docker CLI | For Docker-in-Docker operations |
| Git | For repository operations |
| PlantUML | For diagram rendering |
| MkDocs | For documentation building |
| Pandoc | For PDF generation |

## Usage

```bash
# Run EAC commands via r2r
r2r eac build docs
r2r eac test eac-commands
r2r eac validate

# Or use the image directly
docker run ghcr.io/ready-to-release/ext-eac:latest build docs
```

## Release Process

The ext-eac image is released to GitHub Container Registry:

```bash
# Tag format: ghcr.io/ready-to-release/ext-eac:<version>
r2r eac release ext-eac
```
