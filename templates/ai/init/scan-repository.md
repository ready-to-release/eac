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
- **moniker**: Short identifier (lowercase, hyphenated)
- **name**: Full descriptive name (title case)
- **description**: Clear purpose/role of the module (1-2 sentences)
- **versioning**: Version scheme and release configuration
  - **scheme**: `CalVer` (calendar versioning) or `SemVer` (semantic versioning)
  - **changelog**: Path to changelog file (e.g., `CHANGELOG.md`, `release/module-name/CHANGELOG.md`)
  - **release_type**: `published` (public releases) or `internal` (internal only)
- **components**: Language-specific configuration

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

For each component type, generate appropriate settings:

#### Go Module
```yaml
components:
  go:
    root: path/to/module
    type: service  # or library, cli-tool
```

#### Python Module
```yaml
components:
  python:
    root: path/to/module
    type: service  # or library, cli-tool
```

#### Rust Module
```yaml
components:
  rust:
    root: path/to/module
    type: binary  # or library
```

#### TypeScript Module
```yaml
components:
  typescript:
    root: path/to/module
    type: app  # or library
```

#### .NET Module
```yaml
components:
  dotnet:
    root: path/to/module
    type: webapi  # or library, console
```

#### Java Module
```yaml
components:
  java:
    root: path/to/module
    type: service  # or library
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
- Rust with `main.rs` → "Command-line tool" or "System utility"
- TypeScript with `package.json` → "Web application" or "Frontend"
- Go with `cmd/` → "CLI application"
- Python with `setup.py` → "Library" or "Service"

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
    name: Descriptive Module Name
    description: Module purpose
    versioning:
      scheme: CalVer  # or SemVer
      changelog: path/to/CHANGELOG.md
      release_type: published  # or internal
    components:
      <language>:
        root: relative/path
        type: component-kind
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

**Module Names:**
- Title case, descriptive
- Full name of the module
- Examples: "API Service", "Web Frontend", "CLI Tools"

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

**Component Types:**
- Match actual project structure
- Use standard types: `service`, `library`, `cli-tool`, `app`, `binary`, `webapi`, `console`

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
    name: API Gateway
    description: API gateway routing requests to microservices
    versioning:
      scheme: SemVer
      changelog: services/gateway/CHANGELOG.md
      release_type: published
    components:
      go:
        root: services/gateway
        type: service

  - moniker: auth-service
    name: Auth Service
    description: Authentication and authorization service
    versioning:
      scheme: SemVer
      changelog: services/auth/CHANGELOG.md
      release_type: internal
    components:
      go:
        root: services/auth
        type: service
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
    name: Backend API
    description: Backend API providing REST endpoints
    versioning:
      scheme: CalVer
      changelog: backend/CHANGELOG.md
      release_type: published
    components:
      python:
        root: backend
        type: service

  - moniker: frontend
    name: Frontend Web App
    description: Frontend web application
    versioning:
      scheme: CalVer
      changelog: frontend/CHANGELOG.md
      release_type: published
    components:
      typescript:
        root: frontend
        type: app
    depends_on:
      - backend
```

### 9. Handle Edge Cases

- **No modules detected**: Generate minimal config with repository name only
- **Single module**: Use repository name as module moniker
- **Ambiguous types**: Default to `service` for services, `library` for libraries
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
