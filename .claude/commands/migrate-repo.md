# Migrate Repo

```text
description: "Migrate a repository to EAC framework with auto-detection"
```

You help users migrate repositories to the EAC (Enterprise Application Compliance) framework.

## Process

1. **Determine mode**:
   - Ask user: Showcase mode (demo on popular repos) or Custom mode (current repo)?
   - For showcase: Present catalog of popular repos (flask, ripgrep, cobra, TypeScript, etc.)

2. **Clone repository** (showcase mode only):
   - Clone to `/c/source/ready-to-release/<repo-name>`
   - Navigate into cloned directory

3. **Check existing config**:
   - Look for `.eac/repository.yml`
   - If exists and not --force, ask user for confirmation

4. **Run migration**:
   - Execute: `eac init --scan --ai-provider claude-api [--force]`
   - Use --force for showcase mode (avoid permission issues)
   - For custom mode, only use --force if user confirms

5. **Display results**:
   - Show generated files: repository.yml, books.yml, environments.yml
   - List detected modules with languages
   - Provide next steps: review config, verify modules, commit to git

## Scanner Capabilities

Auto-detects:
- **Languages**: Go, Python, Rust, TypeScript, JavaScript, .NET, Java
- **Module boundaries**: Package manager files (go.mod, package.json, Cargo.toml, pom.xml, etc.)
- **Multi-module repos**: Monorepo detection

## Next Steps Template

After migration:
1. Review: `cat .eac/repository.yml`
2. Verify: `eac show modules`
3. Test build: `eac build <module-name>`
4. Commit: `git add .eac/ && git commit -m "Add EAC configuration"`

## Error Handling

- No modules detected → Explain no package files found
- AI provider error → Falls back to rule-based generation automatically
- Clone failure → Check URL and internet connection

## Important Notes

- Showcase mode: Always use --force to avoid permission prompts
- Custom mode: Ask before overwriting existing config
- AI enhancement (--ai-provider) is optional but recommended
- Large repos may take longer to scan
