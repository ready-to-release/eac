# Language-Specific Commands

Most EAC commands are language-agnostic and work across all supported languages. However, some commands are specific to certain languages.

## Go-Specific Commands

| Command                                               | Purpose                                   | Why Go-Only?                                    |
| ----------------------------------------------------- | ----------------------------------------- | ----------------------------------------------- |
| [`validate-go-tidy`](./validate/go-tidy.md)           | Validates Go module dependencies are tidy | Runs `go mod tidy` and checks for changes       |
| [`validate-dependencies`](./validate/dependencies.md) | Checks go.mod against module contracts    | Parses `go.mod` files to build dependency graph |

These commands specifically interact with Go's module system (`go.mod`) and toolchain.

---

## Language-Agnostic Commands

All other commands work across languages via **capability-based dispatch**. Commands check module capabilities and delegate to appropriate language-specific handlers.

### Build Commands

**Dispatch to language-specific handlers:**

| Module Type  | Handler                       | Builds                                              |
| ------------ | ----------------------------- | --------------------------------------------------- |
| `go`         | GoHandler                     | Go libraries and executables with cross-compilation |
| `typescript` | NpmHandler                    | npm packages via `npm install` and `tsc`            |
| `container`  | BuildxHandler                 | Docker images with multi-platform support           |
| `docs`       | MkdocsHandler                 | Documentation sites and PDFs                        |
| `static`     | NoneHandler or ScriptsHandler | No build or custom scripts                          |

## See Also

- [build](./build/build.md) - Build modules by language type
- [test](./test/test.md) - Test modules by language type
- [Command Overview](./index.md)

**Commands:**

- [`build`](./build/build.md) - Builds modules using appropriate handler
- [`show-build-summary`](./show/build-summary.md) - Shows build results
- [`show-build-times`](./show/build-times.md) - Analyzes build performance

---

### Test Commands

**Dispatch to language-specific runners:**

| Module Type  | Test Types             | Runners                               |
| ------------ | ---------------------- | ------------------------------------- |
| `go`         | `gotest`, `godog`      | GoRunner - Runs `go test` and `godog` |
| `typescript` | `mocha`, `cucumber-js` | MochaRunner, TsCucumberRunner         |

**Commands:**

- [`test`](./test/test.md) - Runs tests using appropriate runner
- [`test-suite`](./test/suite.md) - Executes test suites with filtering
- [`test-debug`](./test/debug.md) - Debugs test failures
- [`show-test-summary`](./show/test-summary.md) - Shows test results
- [`show-test-timings`](./show/test-timings.md) - Analyzes test performance

---

### Validation Commands

**Work on contracts, specifications, and code quality regardless of language:**

- [`validate`](./validate/validate.md) - Runs all validations
- [`validate-contracts`](./validate/contracts.md) - JSON schema validation
- [`validate-specs`](./validate/specs.md) - Gherkin specification validation
- [`validate-markdown`](./validate/markdown.md) - Markdown linting
- [`validate-module-files`](./validate/module-files.md) - File ownership validation
- [`validate-module-hierarchy`](./validate/module-hierarchy.md) - Dependency graph validation
- [`validate-test-tags`](./validate/test-tags.md) - Test tag contract validation

---

### Security Commands

**Scan any codebase with language-aware tools:**

- [`scan`](./scan/scan.md) - Runs security scans with `--scanner` flag
  - `--scanner vuln` - Vulnerability scanning (Trivy)
  - `--scanner sast` - Static analysis (Semgrep)
  - `--scanner secrets` - Secret detection (Trivy)
  - `--scanner iac` - Infrastructure as Code scanning (Trivy)
  - `--scanner sbom` - Software Bill of Materials generation
  - `--scanner compliance` - Compliance checking (Trivy)
- [`scan zap`](./scan/zap.md) - Dynamic application security testing (OWASP ZAP)

Security tools detect languages automatically.

---

### Release Management

**Version control and changelog generation work across all module types:**

- [`release-changelog`](./release/changelog.md) - Generates changelogs from commits
- [`release-this`](./release/this.md) - Prepares module for release
- [`release-pending`](./release/pending.md) - Checks for pending releases
- [`release-check-ci`](./release/check-ci.md) - Verifies CI status before release
- [`release-generate-module-calver`](./release/generate-module-calver.md) - Generates calendar version tags
- [`release-get-version`](./release/get-version.md) - Extracts version from changelog
- [`release-tag-pending`](./release/tag-pending.md) - Creates git tags for releases

---

### Module Management

**Introspection and discovery commands work with all module types:**

- [`show-modules`](./show/modules.md) - Lists all modules
- [`show-component-types`](./show/component-types.md) - Lists module types
- [`show-dependencies`](./show/dependencies.md) - Shows dependency graph
- [`show-files`](./show/files.md) - Shows file ownership
- [`get-modules`](./get/modules.md) - Returns module data as JSON
- [`get-dependencies`](./get/dependencies.md) - Returns dependencies as JSON

---

### AI Workflows

**AI commands generate language-appropriate outputs:**

- [`create-spec`](./create/spec.md) - Generates Gherkin specifications
- [`create-design`](./create/design.md) - Generates Structurizr architecture diagrams
- [`create-commit-message`](./create/commit-message.md) - Generates commit messages
- [`create-pr`](./create/pr.md) - Generates pull request descriptions
- [`create-squash-message`](./create/squash-message.md) - Generates squash commit messages

AI providers adapt output to module context.

---

### Workspace Management

**Git worktree operations work regardless of module language:**

- [`work-create`](./work/create.md) - Creates feature workspace
- [`work-commit`](./work/commit.md) - Commits with AI-generated messages
- [`work-pull`](./work/pull.md) - Syncs workspace with main
- [`work-merge`](./work/merge.md) - Merges workspace changes
- [`work-remove`](./work/remove.md) - Removes workspace
- [`show-workspaces`](./show/workspaces.md) - Lists all workspaces

---

### Documentation

**Documentation generation supports multiple formats:**

- [`serve-docs`](./serve/docs.md) - Starts MkDocs server
- [`serve-design`](./serve/design.md) - Starts Structurizr Lite server
- [`templates-install`](./templates/index.md) - Installs project templates
- [`templates-install-docs`](./templates/install-docs.md) - Installs documentation templates
- [`templates-install-ai`](./templates/install-ai.md) - Installs AI prompt templates

---

## How Capability-Based Dispatch Works

EAC uses a **capability matching system** to route commands to appropriate handlers:

### 1. Modules Declare Capabilities

In `.eac/repository.yml`:

```yaml
- moniker: my-service
  type: go
  capabilities: [go_module, cross_compile]
```

### 2. Handlers Register Capabilities

In handler implementation:

```go
type GoHandler struct {
    Capabilities []string
}

func (h *GoHandler) GetCapabilities() []string {
    return []string{"go_module", "cross_compile"}
}
```

### 3. Commands Dispatch to Matching Handler

```go
handler := GetHandlerForModule(module)
err := handler.Build(ctx, module)
```

### 4. Handler Executes Language-Specific Logic

The handler runs the appropriate build tools (`go build`, `npm install`, `docker buildx`, etc.).

---

## Adding Language Support

To add support for a new language:

1. **Create a handler** - Implement the `Builder` or `Runner` interface
2. **Register capabilities** - Define what the handler can do
3. **Define component type** - Create default file patterns and settings
4. **Register handler** - Add to handler registry in `init()`

See [Component Types Reference](../architecture/component-kinds.md) for detailed instructions.

---

## Language Support Summary

<!-- markdownlint-disable MD060 -->

| Language       | Native Support | Build     | Test                  | Notes                                |
| -------------- | -------------- | --------- | --------------------- | ------------------------------------ |
| **Go**         | ✅ Yes         | ✅ Full   | ✅ gotest, godog      | Complete toolchain integration       |
| **TypeScript** | ✅ Yes         | ✅ Full   | ✅ mocha, cucumber-js | npm and tsc support                  |
| **JavaScript** | ✅ Yes         | ✅ npm    | ✅ mocha, cucumber-js | Via TypeScript module type           |
| **Docker**     | ✅ Yes         | ✅ buildx | -                     | Language-agnostic containerization   |
| **Python**     | ⚠️ Container    | ⚠️ Custom  | ⚠️ Custom              | Use `container` type with Dockerfile |
| **Rust**       | ⚠️ Container    | ⚠️ Custom  | ⚠️ Custom              | Use `container` type with Dockerfile |
| **Java**       | ⚠️ Container    | ⚠️ Custom  | ⚠️ Custom              | Use `container` type with Dockerfile |

<!-- markdownlint-enable MD060 -->

**Legend:**

- ✅ **Full** - Native handler with comprehensive support
- ⚠️ **Custom** - Requires Dockerfile or custom scripts
- **-** - Not applicable or not supported

---

## Related Documentation

- [Component Types Reference](../architecture/component-kinds.md) - Detailed component type specifications
- [Architecture](../architecture/index.md) - System architecture and component design
- [Build Command](./build/build.md) - Build command reference
- [Test Command](./test/test.md) - Test command reference
