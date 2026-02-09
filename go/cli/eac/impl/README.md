# eac Command Implementations

All CLI command implementations for the `eac` binary. Each subdirectory corresponds to a
top-level command or command group registered with the CLI framework.

Module moniker: `eac` | Dependencies: `clibase`, `core`, adapters

## Command Index

| Command Group | Package | Purpose |
| --- | --- | --- |
| `eac build` | [build/](./build/) | Compile modules with parallel, dependency-ordered execution |
| `eac create` | [create/](./create/) | Generate artifacts using AI (commit messages, specs, designs, risk docs) |
| `eac describe` | [describe/](./describe/) | Get structured command information |
| `eac design` | [design/](./design/) | Architecture documentation commands using Structurizr DSL |
| `eac docs` | [docs/](./docs/) | Documentation generation and synchronization |
| `eac drawio` | [drawio/](./drawio/) | Create and manage draw.io diagram files |
| `eac extension` | [extension/](./extension/) | Extension metadata output for clie CLI integration |
| `eac get` | [get/](./get/) | Retrieve repository data in structured formats (YAML, JSON, TOML) |
| `eac help` | [help/](./help/) | Display help information for commands |
| `eac init` | [init/](./init/) | Initialize EAC project configuration |
| `eac lint` | [lint/](./lint/) | Lint modules using the command framework |
| `eac list` | [list/](./list/) | List available commands and resources |
| `eac pipeline` | [pipeline/](./pipeline/) | CI pipeline orchestration (await, dispatch, monitor) |
| `eac release` | [release/](./release/) | Release lifecycle management (await-deps, tagging, changelog) |
| `eac scan` | [scan/](./scan/) | Security scanning using the command framework |
| `eac serve` | [serve/](./serve/) | Start server for module build output |
| `eac show` | [show/](./show/) | Display repository information in human-readable format |
| `eac specs` | [specs/](./specs/) | BDD specification management |
| `eac templates` | [templates/](./templates/) | Install templates without value replacements |
| `eac test` | [test/](./test/) | Run tests with parallel orchestration and result merging |
| `eac update` | [update/](./update/) | Update generated artifacts (docs, evidence, caches, designs) |
| `eac validate` | [validate/](./validate/) | Validate contracts, configuration, dependencies, and specs |
| `eac work` | [work/](./work/) | Working branch management (commit, pull, merge, create) |

## Shared Infrastructure

| Package | Purpose |
| --- | --- |
| [internal/](./internal/) | Shared helpers for GET and SHOW commands (artifact resolution, formatting) |

## Architecture Notes

Commands follow a registration pattern: each package registers itself with the
`registry` during init and implements the command lifecycle defined by `cmdframework`.
Complex commands (build, test, lint, scan) use the orchestrator for parallel execution.
Parent commands like `get`, `show`, `create`, `update`, and `validate` serve as grouping
entry points that delegate to subcommand packages.
