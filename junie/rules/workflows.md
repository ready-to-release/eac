# Workflows

Session Start

1. Read `agent.md`.
2. Read `junie/README.md`.
3. Read all `junie/rules/*.md`.
4. Read `junie/modes.md` and `junie/glossary.md`.
5. Apply precedence from `junie/overrides.md`.

Mode Selection

- Prefer minimal mode needed:
  - Quick Q&A → [CHAT]
  - Repo analysis/advice without edits → [ADVANCED_CHAT]
  - Run tests/app or short checks → [RUN_VERIFY]
  - Single, trivial edit (≤3 steps, one file) → [FAST_CODE]
  - Anything larger/multi‑file or needs investigation → [CODE]
  - Environment/setup tasks → [SETUP]

External Workflows Awareness & Non‑duplication

- Source of truth for Claude workflows lives under `.claude/`:
  - Agents: `.claude/agents/*.md`
  - Commands (slash workflows): `.claude/commands/*.md`
  - Skills (multi‑step workflows): `.claude/skills/*.md`
- Junie must not copy, mirror, or migrate these files into `junie/`.
- When a user asks to run or follow an existing workflow defined for Claude:
  - Read the relevant `.claude/` file(s) in place and follow their guidance.
  - Prefer MCP servers when available; otherwise use CLI fallbacks as defined in `agent.md`.
  - For discovery, list or reference available items by reading directory contents or documented indexes; do not duplicate their content in responses beyond what’s necessary to execute the workflow.
- Workflows that do not require an upstream LLM:
  - Prefer MCP tools (`mcp__commands__*`, `mcp__github__*`) when CONNECTED.
  - If NOT CONNECTED, use documented CLI fallbacks (e.g., `go run ./go/cli/eac <command>`), per `agent.md`.
  - Keep execution non‑interactive and scoped.

Task Switching Protocol

- When a task is completed or the user changes topic/scope, pause to confirm whether to commit current work before proceeding.
- Recommend small, atomic commits with descriptive messages.
- Offer to stage only the relevant files/modules.

Verification

- When changes are made, run the smallest sufficient verification (tests/build) required by the change risk.
- Never submit with failing builds/tests without explicit user approval.
