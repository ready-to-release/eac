# Generate Module Section

Write a module section following this format (do NOT wrap in ```):

<module-name>
---------
<module-name>: <type>: <one-line description>

<1-3 paragraphs describing changes for this module>

Rules:

- First line: plain module name (NO ## prefix)
- Second line: dashes (at least 9 dashes)
- Third line: `<module-name>: <type>: <description>` (max 72 chars)
- Body: 1-3 paragraphs, wrapped at 72 characters
- Types: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`
- NO `---` separator at the end

Example:

```text
contracts
---------
contracts: feat: add commit message validation contract

Added structure.yml defining conventional commit format with scope-based
headers and Auditor-Summary requirements. Includes validation rules for
72-character line limits and format enforcement.
```

Generate the module section now (output ONLY the section, no explanations):
