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

### 1. Module Definitions

For each detected module:
- **moniker**: Short identifier (lowercase, hyphenated)
- **description**: Clear purpose/role of the module (1-2 sentences)
- **components**: Language-specific configuration

### 2. Module Dependencies

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

### 3. Component Configuration

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

### 4. Infer Module Purposes

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

### 5. Output Format

Generate ONLY valid YAML for the `repository.yml` file. Do NOT include markdown fences or explanations.

**Required structure:**
```yaml
# EAC Repository Configuration
# Auto-generated from repository scan

repository:
  name: project-name
  description: Brief project description

modules:
  - moniker: module-name
    description: Module purpose
    components:
      <language>:
        root: relative/path
        type: component-type
```

### 6. Quality Guidelines

**Module Monikers:**
- Lowercase with hyphens
- Descriptive but concise
- Examples: `api-service`, `web-frontend`, `cli-tools`

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

### 7. Multi-Module Repository Patterns

**Monorepo with multiple services:**
```yaml
modules:
  - moniker: api-gateway
    description: API gateway routing requests to microservices
    components:
      go:
        root: services/gateway
        type: service

  - moniker: auth-service
    description: Authentication and authorization service
    components:
      go:
        root: services/auth
        type: service
    depends_on:
      - shared-lib
```

**Full-stack application:**
```yaml
modules:
  - moniker: backend
    description: Backend API providing REST endpoints
    components:
      python:
        root: backend
        type: service

  - moniker: frontend
    description: Frontend web application
    components:
      typescript:
        root: frontend
        type: app
    depends_on:
      - backend
```

### 8. Handle Edge Cases

- **No modules detected**: Generate minimal config with repository name only
- **Single module**: Use repository name as module moniker
- **Ambiguous types**: Default to `service` for services, `library` for libraries
- **Missing README**: Infer from directory names and file structure

## Scan Results

{{.Custom.ScanResults}}

## Repository Context

{{.Custom.RepositoryContext}}

## Example Configurations

{{.Custom.ExampleConfigs}}

## Generate Configuration

Now generate the complete repository.yml configuration based on the scan results above.
Output ONLY valid YAML - no markdown fences, no explanations, no additional text.
