# eac-ext

The `eac-ext` module defines the Docker extension image that packages the EAC tooling for use with clie-cli.

It provides a containerized environment with all dependencies pre-installed.

## System Context

Shows how eac-ext packages EAC tooling for containerized execution.

<!-- structurizr:eac-ext:SystemContext -->

## Container Architecture

High-level view of the extension image structure.

<!-- structurizr:eac-ext:Containers -->

## Dockerfile Components

The layers and components within the Docker image.

<!-- structurizr:eac-ext:DockerfileComponents -->

## Design File

- **Location**: `specs/eac-ext/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module eac-ext`

## Image Contents

The eac-ext Docker image includes:

| Component  | Purpose                         |
| ---------- | ------------------------------- |
| Go Runtime | For executing eac-commands      |
| Docker CLI | For Docker-in-Docker operations |
| Git        | For repository operations       |
| PlantUML   | For diagram rendering           |
| MkDocs     | For documentation building      |
| Pandoc     | For PDF generation              |

## Usage

```bash
# Run EAC commands via clie
eac build docs
eac test eac-commands
eac validate

# Or use the image directly
docker run ghcr.io/ready-to-release/eac-ext:latest build docs
```

## Release Process

The eac-ext image is released to GitHub Container Registry:

```bash
# Tag format: ghcr.io/ready-to-release/eac-ext:<version>
eac release eac-ext
```
