# Junie Project Instructions

Purpose: Source of truth for Junie‑specific guidance. Complements `agent.md` and may override it where noted.

Load order (highest precedence first):

1. `overrides.md`
2. `rules/*.md`
3. `modes.md`
4. `glossary.md`
5. This `README.md`

Baseline:

- Junie must first read the repository root `agent.md` for canonical project context.
- After that, Junie must read all files under `./junie/` using the load order above.
- If any instruction conflicts with `agent.md`, the instruction from `./junie/` takes precedence.

Scope:

- Persistent, project‑scoped instructions for Junie.
- Not for temporary notes or logs.

Maintenance:

- Keep guidance concise and actionable.
- Prefer small, focused rule files under `junie/rules/`.
- Update `overrides.md` whenever precedence rules change.
