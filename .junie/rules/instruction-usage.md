# Instruction & Tool Usage Guidance

## Core Requirement
Before executing any project-specific commands or maintenance tasks, Junie MUST:
1.  **Source the environment:** Run `source ./importer.sh` in the terminal. This silences pedantic Cgo warnings on macOS and provides the `run` and `eac` aliases.
2.  **Consult specialized instructions** located in the following directories:
    - `.claude/agents/`: Role-specific guidance (Architect, Debugger, Test Engineer, etc.).
    - `.claude/commands/`: Command-specific workflows (Implement, Plan, Test, Review, etc.).
    - `.claude/skills/`: Specialized skill instructions (Refactor-safe, Feature workflow, etc.).
    - `docs/how-to-guides/eac/commands/`: Comprehensive manual for the `eac` CLI tool.

## Tool Usage (`eac` / `clie`)
- The primary tool for repository maintenance is the `eac` CLI, which is available via the `eac` alias after sourcing `importer.sh`.
- **Validation:** Always use `eac validate` for pre-commit or health checks.
- **Documentation & Assets:** Use `eac update docs` to sync documentation assets, including command references and diagram caches.
- **Architecture:** Use `eac create design` or `eac update design` for Structurizr diagrams.

## Diagram Handling (`draw.io` / Structurizr)
- Follow instructions in `.claude/commands/drawio.md` and `.claude/skills/drawio-editor.md` when handling diagrams.
- If BDD tests report missing `drawio` cache entries, the standard resolution is to run `eac update docs`.

## Maintenance Workflow
- If the repository state is "unhealthy" (failing repository specs), prioritize using the dedicated `eac update` commands over manual fixes to ensure consistency with the toolchain's expectations.
