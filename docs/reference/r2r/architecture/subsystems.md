# Subsystems

Detailed component architecture for each R2R CLI subsystem.

## Configuration System

Multi-layer configuration management supporting base, local, personal, and dev configs.

<!-- structurizr:r2r-cli:ConfigurationSystem -->

**Components:**

- **Config Loader** - Loads `.r2r/r2r-cli.yml` configuration
- **Repository Finder** - Locates git repository root
- **Extension Validator** - Validates pinned SHA tags
- **Config Merger** - Merges base + local + personal + dev configs

---

## Parser System

EBNF-based command-line argument parsing with boundary detection.

<!-- structurizr:r2r-cli:ParserSystem -->

**Components:**

- **EBNF Parser** - Parses arguments per EBNF grammar
- **Boundary Detector** - Identifies Viper vs container arg boundary
- **Global Flag Parser** - Extracts `--r2r-debug`, `--r2r-quiet` flags
- **Extension Resolver** - Resolves extension names and subcommands

---

## Docker Orchestration

Container lifecycle management and execution with TTY support.

<!-- structurizr:r2r-cli:DockerOrchestration -->

**Components:**

- **Container Host** - Manages container lifecycle operations
- **Container Creator** - Creates container configs with volumes
- **Container Runner** - Starts containers and manages execution
- **Attach Manager** - Handles stdin/stdout/stderr attachment
- **TTY Handler** - Manages TTY vs non-TTY modes
- **Signal Handler** - Handles SIGINT/SIGTERM for graceful shutdown
- **Cleanup Manager** - Cleans up child containers (Docker-in-Docker)
- **ANSI Filter** - Filters problematic escape sequences
- **Image Inspector** - Inspects Docker images for metadata
- **Volume Mapper** - Maps repository root to `/workspace`

---

## Extension Management

Extension installation and lifecycle management with pull policies.

<!-- structurizr:r2r-cli:ExtensionManagement -->

**Components:**

- **Extension Installer** - Ensures images exist locally
- **Pull Policy Handler** - Handles Always/IfNotPresent/Never policies
- **Registry Client** - Interacts with GitHub Container Registry
- **SHA Resolver** - Discovers latest SHA tags for pinning
- **Local Loader** - Loads local development images

---

## Validation System

Configuration and command validation against schemas and contracts.

<!-- structurizr:r2r-cli:ValidationSystem -->

**Components:**

- **Command Validator** - Validates commands against EBNF
- **Schema Validator** - Validates `.r2r/r2r-cli.yml` against JSON schema
- **Contract Validator** - Validates against contracts

---

## Logging System

Structured logging with context management using Zerolog.

<!-- structurizr:r2r-cli:LoggingSystem -->

**Components:**

- **Logger** - Structured logging with context
- **Context Manager** - Manages command, component, operation_id context
- **Level Controller** - Controls debug/info/warn/error/fatal levels
- **Output Formatter** - Formats console and file output

---

## Terminal System

Cross-platform terminal handling for Unix and Windows.

<!-- structurizr:r2r-cli:TerminalSystem -->

**Components:**

- **Terminal Detector** - Detects terminal capabilities
- **Unix Handler** - Unix-specific terminal operations
- **Windows Handler** - Windows-specific terminal operations
- **Console API** - Windows console API integration
- **Spinner** - Progress indicators for long operations

---

## GitHub Integration

GitHub Container Registry integration for image management.

<!-- structurizr:r2r-cli:GitHubIntegration -->

**Components:**

- **Registry API Client** - HTTP client for GHCR
- **Tag Fetcher** - Fetches available image tags
- **Auth Handler** - Handles GITHUB_TOKEN authentication
- **Cache Manager** - Caches registry responses

---

## Cache System

Caching for registry responses and extension metadata.

<!-- structurizr:r2r-cli:CacheSystem -->

**Components:**

- **Registry Cache** - Caches GitHub registry API responses
- **Metadata Cache** - Caches extension metadata

---

## Session System

CLI session state management across commands.

<!-- structurizr:r2r-cli:SessionSystem -->

**Components:**

- **Session Manager** - Tracks session state across commands

---

## TUI System

Terminal UI components for visual feedback.

<!-- structurizr:r2r-cli:TUISystem -->

**Components:**

- **TUI Spinner** - Progress indicators for long operations
- **Progress Bar** - Visual progress for downloads

## See Also

- [System Context](./system-context.md) - High-level architecture
- [Workflows](./workflows.md) - Dynamic workflow diagrams
