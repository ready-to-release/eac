# Command Categories

{{ page_breadcrumb() }}

Browse EAC commands organized by function. Each category groups commands that serve a similar purpose.

## All Categories

### create (7 commands)

**Purpose**: AI-powered content generation

Generate commit messages, specifications, architecture diagrams, and other content using AI assistance.

[Browse create commands →](./create.md)

---

### get (18 commands)

**Purpose**: JSON output for automation

Retrieve repository data in structured JSON format for CI/CD pipelines, build scripts, and automation tools.

[Browse get commands →](./get.md)

---

### show (17 commands)

**Purpose**: Human-readable output

Display repository information in formatted tables and text for interactive terminal use.

[Browse show commands →](./show.md)

---

### validate (20 commands)

**Purpose**: Contract and quality validation

Ensure repository consistency, enforce contracts, and verify quality standards.

[Browse validate commands →](./validate.md)

---

### work (6 commands)

**Purpose**: Workspace management

Manage parallel development workspaces using git worktrees for isolated feature development.

[Browse work commands →](./work.md)

---

### test (4 commands)

**Purpose**: Testing and test suite management

Run tests, manage test suites, and analyze test results.

[Browse test commands →](./test.md)

---

### build (1 command)

**Purpose**: Module building

Build modules in dependency order.

[Browse build command →](./other.md#build)

---

### pipeline (6 commands)

**Purpose**: CI/CD orchestration

Orchestrate CI/CD workflows, monitor build status, and generate diagnostics.

[Browse pipeline commands →](./pipeline.md)

---

### release (8 commands)

**Purpose**: Release management

Manage the release process from changelog generation through tagging and validation.

[Browse release commands →](./release.md)

---

### scan (8 commands)

**Purpose**: Security scanning

Perform security scans and generate compliance evidence.

[Browse scan commands →](./scan.md)

---

### serve (2 commands)

**Purpose**: Local development servers

Start local servers for documentation and architecture visualization.

[Browse serve commands →](./serve.md)

---

### templates (7 commands)

**Purpose**: Template management

Manage project templates for consistent documentation and specification creation.

[Browse templates commands →](./templates.md)

---

### update (1 command)

**Purpose**: Update operations

Update existing artifacts like architecture diagrams.

[Browse update command →](./update.md)

---

### other (3 commands)

**Purpose**: Utility commands

Commands that don't fit into verb-based categories: help, init, and extension metadata.

[Browse other commands →](./other.md)

---

## Category Quick Reference

| Category      | Commands | Primary Use             | Output Format  |
| ------------- | -------- | ----------------------- | -------------- |
| **create**    | 7        | AI content generation   | Varies         |
| **get**       | 18       | Automation & scripting  | JSON           |
| **show**      | 17       | Interactive exploration | Formatted text |
| **validate**  | 20       | Quality gates           | Pass/fail      |
| **work**      | 6        | Workspace management    | Varies         |
| **test**      | 4        | Testing                 | Test results   |
| **build**     | 1        | Building                | Artifacts      |
| **pipeline**  | 6        | CI/CD                   | Varies         |
| **release**   | 8        | Release management      | Varies         |
| **scan**      | 8        | Security                | Reports        |
| **serve**     | 2        | Local servers           | Process        |
| **templates** | 7        | Templates               | Varies         |
| **update**    | 1        | Updates                 | Varies         |
| **other**     | 3        | Utilities               | Varies         |

## Finding Commands by Task

### Development Tasks

**I want to...**

- **...see what modules exist**: [show modules](../show/modules.md)
- **...check dependencies**: [show dependencies](../show/dependencies.md) or [get dependencies](../get/dependencies.md)
- **...create a workspace**: [work create](../work/create.md)
- **...write specifications**: [create spec](../create/spec.md)
- **...generate diagrams**: [create design](../create/design.md)

### Quality & Validation

**I want to...**

- **...validate contracts**: [validate contracts](../validate/contracts.md)
- **...check dependencies**: [validate dependencies](../validate/dependencies.md)
- **...validate specs**: [validate specs](../validate/specs.md)
- **...scan for secrets**: [scan secrets](../scan/secrets.md)
- **...check for vulnerabilities**: [scan vuln](../scan/vuln.md)

### Building & Testing

**I want to...**

- **...build a module**: [build](../other/build.md)
- **...run tests**: [test](../test/test.md)
- **...run a test suite**: [test suite](../test/suite.md)
- **...see test results**: [show test-summary](../show/test-summary.md)
- **...get build order**: [get execution order](../get/execution-order.md)

### Committing & PRs

**I want to...**

- **...commit changes**: [work commit](../work/commit.md) or [create commit-message](../create/commit-message.md)
- **...create a PR**: [create pr](../create/pr.md)
- **...see changed files**: [show files-changed](../show/files-changed.md)
- **...get changed modules**: [get changed-modules](../get/changed-modules.md)

### Release & Deployment

**I want to...**

- **...check for release changes**: [release pending](../release/pending.md)
- **...generate changelog**: [release changelog](../release/changelog.md)
- **...create a release**: [release this](../release/this.md)
- **...check CI status**: [release check-ci](../release/check-ci.md)
- **...run pipelines**: [pipeline run](../pipeline/run.md)

### Documentation

**I want to...**

- **...serve docs locally**: [serve docs](../serve/docs.md)
- **...view architecture diagrams**: [serve design](../serve/design.md)
- **...manage templates**: [templates commands](./templates.md)

## Category Patterns

### Information Retrieval (get/show)

Most information commands come in pairs:

| get (JSON)         | show (Formatted)    | Information      |
| ------------------ | ------------------- | ---------------- |
| `get modules`      | `show modules`      | Module contracts |
| `get dependencies` | `show dependencies` | Dependency graph |
| `get files`        | `show files`        | File ownership   |
| `get config`       | `show config`       | Configuration    |
| `get tests`        | `show tests`        | Test information |

**Rule**: Use `get` for automation, `show` for interactive use.

### Quality Gates (validate/scan/test)

Quality commands all validate different aspects:

- **validate**: Repository structure and contracts
- **scan**: Security issues and compliance
- **test**: Functional correctness

**Use in**: Pre-commit hooks, CI pipelines, release gates

### Workflow (work/pipeline/release)

Workflow commands manage different stages:

- **work**: Local development (git worktrees)
- **pipeline**: CI/CD execution
- **release**: Version management

## See Also

- [Command Reference Index](../index.md) - Main command reference
- [Command Taxonomy](../overview/command-taxonomy.md) - How commands are organized
- [Naming Conventions](../overview/naming-conventions.md) - Command naming rules
- [Output Formats](../overview/output-formats.md) - JSON vs formatted output

{{ diataxis_footer() }}
