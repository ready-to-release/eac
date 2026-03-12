# EAC Architecture

## Viewing Architecture Diagrams

All architecture diagrams referenced in this document can be viewed interactively using Structurizr:

```bash
eac serve design
# Opens http://localhost:8080
```

**Design files:**

- **eac**: [specs/eac/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac/.design/)
- **core**: [specs/core/.design/](https://github.com/ready-to-release/eac/tree/main/specs/core/.design/)
- **eac-mcp-server**: [specs/eac-mcp-server/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-mcp-server/.design/)
- **eac-ext**: [specs/eac-ext/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-ext/.design/)

See [Viewing Architecture](./viewing-diagrams.md) for detailed instructions.

---

## System Context

```mermaid
graph TB
    Dev[Developer] -->|eac| EAC[EAC CLI]
    LLM[LLM Tools] -->|MCP| MCP[MCP Server]

    Dev -.->|optional| CLI[CLIE CLI]
    CLI -->|Docker| Ext[eac-ext Container]
    Ext --> Commands

    EAC --> Commands[eac-commands]
    MCP -->|Direct| Commands

    Commands --> Core[core]
    Commands --> Specs[godog-eac]

    Core --> Contracts[YAML Contracts]
    Core --> Modules[Go Modules]
    Specs --> Features[BDD Specs]

    style EAC fill:#e1ffe1
    style CLI fill:#f5f5f5
    style Ext fill:#ffe1e1
    style Core fill:#e1ffe1
```

**EAC operates in three execution modes**:

1. **Standalone CLI** (recommended): Developer runs `eac <command>` directly on host machine -- native Go performance, no Docker required
2. **Containerized** (via CLIE): Developer runs `eac <command>` → CLIE CLI launches eac-ext Docker container → Command executes in isolated environment
3. **MCP Server**: LLM tools connect via MCP protocol → eac-mcp-commands exposes tools → Commands execute directly

---

## EAC Components

EAC consists of 28 modules across six groups: **Core Framework** (core, clibase, contracts), **CLI Tools** (eac, eac-ext, eac-mcp-server), **Commands** (7 domain components), **Adapters** (16 tool integrations), **OCI Tools** (12 container images), and **Supporting** (docs, templates, bundle, installers, etc.).

See [Modules Reference](../modules/index.md) for the complete module catalog and per-module documentation.

The EAC extension container packages the compiled commands binary:

```text
eac-ext:latest
├── /app/out/tools/eac    # Compiled commands binary
├── /var/task             # Mounted repository (volume)
└── /usr/local/bin        # System dependencies (git, gh, etc.)
```

---

## Architecture Patterns

### Contract-Driven Architecture

All repository structure is defined in YAML contracts validated against JSON schemas. Contracts are loaded by `core` at runtime and validated before any operation, ensuring type safety with clear error messages and schema-based documentation.

See [Contracts System](./contracts.md) for the complete contract field reference.

### Module System

Modules are the fundamental unit of organization in EAC repositories. Each module:

- Has a unique moniker (e.g., `eac-commands`, `clie`)
- Owns specific files (defined in `modules.yml`)
- Declares dependencies on other modules
- Has a module type that determines build/test/release behavior

**Module boundaries** are enforced:

- Files can only belong to one module
- Dependencies must be explicitly declared
- Circular dependencies are detected and rejected

See [Module System](./modules.md) for module organization patterns.

### Dependency Management

Modules form a directed acyclic graph (DAG) based on declared dependencies.

EAC uses this graph to:

- **Topological sorting**: Determine build order (dependencies before dependents)
- **Parallel execution**: Build independent modules concurrently
- **Change detection**: Rebuild only changed modules and dependents
- **CI optimization**: Dispatch workflows only for affected modules

All module dependencies are declared via a single `depends_on` field that covers
build, test, and deployment relationships.

See [Dependency System](./dependencies.md) for graph algorithms and caching strategies.

---

## Command Organization

**Commands organized by category**:

| Category       | Example Commands                                                     |
| -------------- | -------------------------------------------------------------------- |
| **Discovery**  | `show-modules`, `show-dependencies`, `get-files`                     |
| **Build**      | `build`, `get-artifacts`, `validate-artifacts`                       |
| **Test**       | `test`, `test-suite`, `test-debug`, `show-test-summary`              |
| **Validation** | `validate-contracts`, `validate-dependencies`, `validate-specs`      |
| **Release**    | `release-changelog`, `release-this`, `release-pending`               |
| **Security**   | `scan`, `scan-zap`, `validate-risk-catalog`                          |
| **AI**         | `get-commit-message`, `create-spec`, `create-design`, `create-pr` |
| **CI/CD**      | `pipeline-run`, `pipeline-wait`, `get-changed-modules-ci`            |

All commands are implemented in `eac` following a consistent pattern:

```go
// All commands follow this pattern
type BuildHandler struct {
    repo *Repository  // From core
}

func (h *BuildHandler) Execute(args []string) error {
    // 1. Load contracts from .eac/
    // 2. Resolve dependencies using eac-core
    // 3. Execute build logic
    // 4. Verify artifacts and update cache
    return nil
}
```



---

## AI Integration

**Purpose**: AI provider integrations for automated workflows

**Location**: Integrated within eac-commands at `go/adapters/ai/`

### Supported Providers

| Provider      | Models                     | Use Cases                            |
| ------------- | -------------------------- | ------------------------------------ |
| **Anthropic** | Claude Opus, Sonnet, Haiku | Commit messages, specs, designs, PRs |
| **OpenAI**    | GPT-4, GPT-3.5             | Commit messages, specs, designs, PRs |

### Configuration

**File**: `.eac/ai-config.yml`

```yaml
provider: anthropic
model: claude-sonnet-4
api_key_env: ANTHROPIC_API_KEY
```

### AI-Powered Commands

| Command                 | AI Task                                     | Output          |
| ----------------------- | ------------------------------------------- | --------------- |
| `get-commit-message` | Analyze git diff, generate semantic message | Commit message  |
| `create-spec`           | Convert natural language to Gherkin         | `.feature` file |
| `create-design`         | Generate Structurizr DSL from description   | `workspace.dsl` |
| `create-pr`             | Analyze commits, generate PR description    | GitHub PR body  |

**Implementation**:

- Tool handler: `go/adapters/ai/toolhandler/handler.go`
- Configuration loading: `go/adapters/ai/config_loader.go`

AI commands use a **retry strategy** with exponential backoff for rate limiting and transient errors.

---

## Configuration Management

### Configuration Hierarchy

**Precedence** (highest to lowest):

1. **Personal config** (`.personal.yml`, not in Git) - User-specific overrides
2. **User config** (`.yml` files in `.eac/`) - Team shared settings
3. **System defaults** (`contracts/eac-core/0.1.0/defaults/`) - Default configurations
4. **Hardcoded defaults** (in eac-core) - Fallback values

This hierarchy allows:

- Teams to share conventions in repository
- Individuals to override without affecting others
- Types to provide sensible defaults
- System to never fail on missing config

---

## Technology Stack

### EAC-Specific Technologies

| Component         | Technology                | Purpose                     |
| ----------------- | ------------------------- | --------------------------- |
| **CLI Framework** | Cobra                     | Command parsing and routing |
| **Configuration** | Viper                     | YAML/ENV config loading     |
| **BDD Testing**   | Godog (Cucumber for Go)   | Gherkin feature execution   |
| **JSON Schema**   | gojsonschema              | Contract validation         |
| **AI Providers**  | Anthropic SDK, OpenAI SDK | LLM integrations            |
| **MCP Server**    | JSON-RPC over stdio       | LLM tool exposure           |

### External Tools

| Tool                | Purpose                              |
| ------------------- | ------------------------------------ |
| **Git**             | Version control and change detection |
| **GitHub CLI (gh)** | GitHub API access                    |
| **MkDocs**          | Documentation generation             |
| **Structurizr**     | C4 model architecture diagrams       |
| **Trivy**           | Vulnerability scanning, SBOM         |
| **Semgrep**         | Static analysis (SAST)               |
| **OWASP ZAP**       | Dynamic analysis (DAST)              |

Container technologies (Docker, Docker SDK) are used by the [CLIE extension host](https://ready-to-release.github.io/eac/reference/clie/architecture/) for containerized execution.

---

## Security

**EAC Security Features**:

- **Pre-commit validation**: Validate contracts, specs, and code before committing
- **Security scanning**: Integrated Trivy (vulnerabilities), Semgrep (SAST), ZAP (DAST)
- **OSCAL compliance**: Automated evidence generation in OSCAL 1.1.2 format
- **Secret detection**: Prevent accidental commit of secrets
- **Dependency validation**: Verify module dependencies match declarations

**Evidence collection**:

- **Test results**: BDD feature execution results in structured format
- **Scan results**: Vulnerability, SAST, DAST findings
- **OSCAL documents**: Assessment results, catalogs, profiles
- **Artifacts**: All evidence stored in `out/evidence/` for audit

**Container-level security** (when running via CLIE):

Isolation, non-root execution, and network restrictions are provided by the [CLIE extension host](https://ready-to-release.github.io/eac/reference/clie/architecture/#security-model).

---

## Performance and Scalability

### Parallel Execution

- **Build**: Topological sort enables parallel module builds
- **Test**: Test suites run in parallel by default
- **CI**: GitHub Actions matrix strategy for cross-platform testing

### Caching

The cache system uses a **2D taxonomy** (Level x Type) to classify all caches:

- **Level**: `local` (developer machine) or `remote` (network/CI)
- **Type**: `registry`, `state`, `asset`, `layer`, `work`, or `ci`

Key cache types:

- **`local:state`** — UoW manifests in `out/` with input/output hashes
- **`local:asset`** — Rendered diagrams (Mermaid, Structurizr) with content-based invalidation
- **`local:layer`** — Docker BuildKit layer cache
- **`remote:ci`** — CI build status (last successful GitHub Actions run per module)

Any cache can be bypassed with `--skip-cache=<spec>` (e.g., `--skip-cache=remote:ci`).

See [Cache System](./cache-system.md) for the full 2D taxonomy and CI cache architecture.

### Change Detection

EAC uses **UoW manifest-based change detection** for local builds and **CI cache checking** for remote CI dispatch:

1. **Input hashing**: SHA256 of all source files matched by component patterns
2. **Manifest comparison**: Compare current input hash against stored UoW manifest
3. **Cross-context invalidation**: Rebuild tests when builds produce new output
4. **Dependency propagation**: Rebuild dependents of changed modules
5. **CI cache**: Skip CI dispatch when a module's last successful run matches HEAD SHA

See [Cache System](./cache-system.md) for the detection algorithm.

### Build Execution

All commands (build, test, lint, scan) are executed through a unified
orchestrator that manages parallel UoW execution with capacity-aware
scheduling.

See [Build Execution System](./build-execution.md) for the full
architecture.

---

## Optional: CLIE Extension Host

EAC can optionally run inside the [CLIE CLI framework](https://ready-to-release.github.io/eac/reference/clie/architecture/) for containerized execution:

- **CLIE provides**: Container orchestration, git discovery, volume mounting, configuration loading
- **EAC provides**: Commands, contracts, modules, AI integration, security scanning

See [Running EAC via CLIE](./cli-integration.md) for details on the CLIE/EAC integration.

---

## Related Documentation

### Architecture

- [CLIE CLI Architecture](https://ready-to-release.github.io/eac/reference/clie/architecture/) - Framework overview and container model
- [Modules Reference](../modules/index.md) - Module documentation and C4 diagrams
- [Contracts System](./contracts.md) - YAML contract specification
- [Dependency System](./dependencies.md) - Module dependency graph
- [Component Types](./component-kinds.md) - Component type reference
- [Repository Layout](./repository-layout.md) - File organization conventions
- [CLI Integration](./cli-integration.md) - CLIE ↔ EAC integration details
- [Build Execution System](./build-execution.md) - UoW orchestration and parallel scheduling
- [Cache System](./cache-system.md) - Incremental builds and input hashing
- [Component Resolution](./component-resolution.md) - How contracts become executable UoWs
- [Tool System](./tool-system.md) - Tool composition, executor modes, and container configuration

### How-To Guides

- [Building Modules](../../../how-to-guides/eac/commands/build-test-validate/build-single-module.md)
- [Running Tests](../../../how-to-guides/eac/commands/build-test-validate/run-tests-for-module.md)
- [Creating Specifications](../../../how-to-guides/eac/commands/documentation/create-specifications.md)
- [Generating Architecture Diagrams](../../../how-to-guides/eac/commands/documentation/generate-architecture-diagrams.md)

### Reference

- [EAC Commands](../commands/index.md) - Complete command reference
- [Decision Records](../../../architecture/decisions/index.md) - Architectural decisions
