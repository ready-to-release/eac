# Generate Repository Configuration from Scan Results

You are an expert in software architecture and repository configuration.

Generate a complete EAC repository configuration (repository.yml) based on scan results.

## Input Data

You will receive:
1. **Scan Results**: Detected modules with language, build tool, and file paths
2. **Repository Context**: README content, directory structure hints
3. **Example Repositories**: Reference configurations from similar projects

## Your Task

Analyze the scan results and generate an intelligent `repository.yml` configuration with:

### 1. Repository Configuration

Configure the repository-level settings:
- **type**: `mono` (monorepo with multiple modules) or `poly` (multi-repo)
- **remote**: GitHub repository information
  - **owner**: GitHub organization or user
  - **repo**: Repository name

### 2. Module Definitions

For each detected module:
- **moniker**: Short identifier (lowercase, hyphenated). This is the unique module identifier.
- **description**: Clear purpose/role of the module (1-2 sentences)
- **versioning**: Version scheme and release configuration
  - **scheme**: `CalVer` (calendar versioning) or `SemVer` (semantic versioning)
  - **changelog**: Path to changelog file (e.g., `CHANGELOG.md`, `release/module-name/CHANGELOG.md`)
  - **release_type**: `published` (public releases) or `internal` (internal only)
- **components**: Language-specific configuration (list format — see section 4)

### 3. Module Dependencies

Analyze and infer dependencies between modules:
- API services that frontends depend on
- Shared libraries that other modules use
- CLI tools that depend on core libraries

Use `depends_on` to specify dependencies:
```yaml
modules:
  - moniker: frontend
    depends_on:
      - api
```

### 4. Component Configuration

Components are declared as a **YAML list**. Each component has a `type` (the language/tool kind)
and a `root` path relative to the repository root.

#### Go Module
```yaml
components:
  - type: go
    root: path/to/module
```

#### Python Module
```yaml
components:
  - type: python
    root: path/to/module
```

#### Rust Module
```yaml
components:
  - type: rust
    root: path/to/module
```

#### TypeScript Module
```yaml
components:
  - type: typescript
    root: path/to/module
```

#### JavaScript Module
```yaml
components:
  - type: javascript
    root: path/to/module
```

#### .NET Module
```yaml
components:
  - type: dotnet
    root: path/to/module
```

#### Java Module
```yaml
components:
  - type: java
    root: path/to/module
```

### 5. Infer Module Purposes

Use these heuristics:

**By Directory Name:**
- `api`, `api-service`, `backend` → "Backend API service"
- `frontend`, `web`, `ui` → "Frontend web application"
- `cli`, `cmd`, `tools` → "Command-line tool"
- `lib`, `core`, `common` → "Shared library"
- `worker`, `processor` → "Background worker/processor"

**By Language/Tool:**
- Rust → "Command-line tool" or "System utility"
- TypeScript with `package.json` → "Web application" or "Frontend"
- Go with `cmd/` → "CLI application"
- Python with `setup.py` or `pyproject.toml` → "Library" or "Service"

**By File Patterns:**
- Contains `Dockerfile` → Service/deployable application
- Contains `tests/` or `test/` → Include testing info
- Contains API-related files → API service

### 6. Output Format

Generate ONLY valid YAML for the `repository.yml` file. Do NOT include markdown fences or explanations.

**Required structure:**
```yaml
# EAC Repository Configuration
# Auto-generated from repository scan

repository:
  type: mono  # or poly for multi-repo
  remote:
    owner: github-owner
    repo: repository-name

modules:
  - moniker: module-name
    description: Module purpose
    versioning:
      scheme: CalVer  # or SemVer
      changelog: path/to/CHANGELOG.md
      release_type: published  # or internal
    components:
      - type: <language>
        root: relative/path
```

### 7. Quality Guidelines

**Repository Type:**
- Use `mono` for monorepos (single repository with multiple modules)
- Use `poly` for multi-repo setups (one repository per module)
- Most projects are `mono`

**Remote Configuration:**
- Extract from git remote URL if available
- Default to generic values if not detected

**Module Monikers:**
- Lowercase with hyphens
- Descriptive but concise
- Examples: `api-service`, `web-frontend`, `cli-tools`

**Versioning:**
- Use `CalVer` for date-based releases (YYYY.0M.0D format)
- Use `SemVer` for semantic versioning (MAJOR.MINOR.PATCH)
- Default changelog path: `CHANGELOG.md` or `release/{moniker}/CHANGELOG.md`
- Use `published` for public modules, `internal` for internal-only modules

**Descriptions:**
- Start with action verb or noun
- Be specific about purpose
- 1-2 sentences maximum
- Examples:
  - "Backend API service providing REST endpoints"
  - "Command-line tool for data processing"
  - "Frontend web application built with React"

**Component types** must match the detected language from scan results:
`go`, `python`, `rust`, `typescript`, `javascript`, `dotnet`, `java`

### 8. Multi-Module Repository Patterns

**Monorepo with multiple services:**
```yaml
repository:
  type: mono
  remote:
    owner: myorg
    repo: microservices

modules:
  - moniker: api-gateway
    description: API gateway routing requests to microservices
    versioning:
      scheme: SemVer
      changelog: services/gateway/CHANGELOG.md
      release_type: published
    components:
      - type: go
        root: services/gateway

  - moniker: auth-service
    description: Authentication and authorization service
    versioning:
      scheme: SemVer
      changelog: services/auth/CHANGELOG.md
      release_type: internal
    components:
      - type: go
        root: services/auth
    depends_on:
      - shared-lib
```

**Full-stack application:**
```yaml
repository:
  type: mono
  remote:
    owner: mycompany
    repo: webapp

modules:
  - moniker: backend
    description: Backend API providing REST endpoints
    versioning:
      scheme: CalVer
      changelog: backend/CHANGELOG.md
      release_type: published
    components:
      - type: python
        root: backend

  - moniker: frontend
    description: Frontend web application
    versioning:
      scheme: CalVer
      changelog: frontend/CHANGELOG.md
      release_type: published
    components:
      - type: typescript
        root: frontend
    depends_on:
      - backend
```

### 9. Handle Edge Cases

- **No modules detected**: Generate minimal config with repository name only
- **Single module**: Use repository name as module moniker
- **Missing README**: Infer from directory names and file structure
- **Unknown remote**: Use generic placeholder values for owner/repo
- **Versioning defaults**: Use `CalVer` for services/apps, `SemVer` for libraries
- **Release type defaults**: Use `published` unless clearly internal-only

## Scan Results

{{.Custom.ScanResults}}

## Repository Context

{{.Custom.RepositoryContext}}

## Example Configurations

{{.Custom.ExampleConfigs}}

## Generate Configuration

Now generate the complete repository.yml configuration based on the scan results above.
Output ONLY valid YAML - no markdown fences, no explanations, no additional text.
