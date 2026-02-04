# Overrides

These rules explicitly take precedence over any conflicting guidance in `agent.md`.

Precedence Policy

- Source order: `junie/overrides.md` > `junie/rules/*.md` > `junie/modes.md` > `junie/glossary.md` > `junie/README.md` > `agent.md`.
- If a conflict is detected, prefer the earliest item in the order above.

Session Initialization

- Always read `agent.md` first.
- Then read `junie/README.md` and all files under `junie/` following the load order above.

Claude Workflows Non‑Duplication (MANDATORY)

- Be aware of and reference existing workflows defined for Claude in `.claude/agents/`, `.claude/commands/`, and `.claude/skills/`.
- Do NOT copy, mirror, or migrate these files into `junie/`.
- When execution is requested, read and follow the `.claude/` sources in place.
- Prefer MCP tools when connected; otherwise use CLI fallbacks as defined in `agent.md`.

File Modification Permission (MANDATORY)

- Do not create, edit, move, or delete any files without explicit confirmation from the user for that specific action.
- Before any file system change, pause and ask for permission, summarizing the intended change, affected paths, and rationale.

Commit Reminder Between Tasks (MANDATORY)

- When finishing a task or when the user changes topic/scope, proactively ask whether to commit current work to git before proceeding.
- Encourage small, atomic commits with clear messages; offer to stage only the touched files/modules.

Operational Safeguards

- Do not run destructive commands without explicit confirmation from the user.
- Keep changes minimal and targeted; avoid broad refactors unless requested.
- Prefer explanations that include trade‑offs when multiple options are viable.
