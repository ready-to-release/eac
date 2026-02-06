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
   - If exists, init will automatically re-initialize and merge configs

4. **Run migration**:
   - Execute: `eac init --scan --ai-provider claude-api`
   - Preserves user customizations during re-init

5. **Display results**:
   - Show generated files: repository.yml, books.yml, environments.yml
   - List detected modules with languages
   - Provide next steps: review config, verify modules, commit to git

## Example Repository Catalog

### Go Projects
- **cobra** - `https://github.com/spf13/cobra` - CLI library (single module)
- **hugo** - `https://github.com/gohugoio/hugo` - Static site generator (monorepo)
- **traefik** - `https://github.com/traefik/traefik` - Cloud native reverse proxy (multi-module)

### Python Projects
- **flask** - `https://github.com/pallets/flask` - Web framework (single module)
- **django** - `https://github.com/django/django` - Web framework (monorepo structure)
- **requests** - `https://github.com/psf/requests` - HTTP library (simple structure)

### Rust Projects
- **ripgrep** - `https://github.com/BurntSushi/ripgrep` - Fast grep tool (Cargo workspace)
- **tokio** - `https://github.com/tokio-rs/tokio` - Async runtime (complex monorepo)
- **serde** - `https://github.com/serde-rs/serde` - Serialization framework (multi-crate)

### TypeScript/JavaScript Projects
- **vscode** - `https://github.com/microsoft/vscode` - Code editor (large monorepo)
- **react** - `https://github.com/facebook/react` - UI library (monorepo with multiple packages)
- **next.js** - `https://github.com/vercel/next.js` - React framework (monorepo)

### .NET Projects
- **roslyn** - `https://github.com/dotnet/roslyn` - C# compiler (solution with multiple projects)
- **aspnetcore** - `https://github.com/dotnet/aspnetcore` - ASP.NET Core framework

### Java Projects
- **spring-boot** - `https://github.com/spring-projects/spring-boot` - Application framework (Gradle multi-module)
- **elasticsearch** - `https://github.com/elastic/elasticsearch` - Search engine (Gradle build)

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

- Smart re-initialization: Init automatically detects and merges existing configs
- User edits preserved: Module names, versioning, and custom dependencies are kept
- AI enhancement (--ai-provider) is optional but recommended
- Large repos may take longer to scan
