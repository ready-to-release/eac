# Intelligent Module Detection Guide

This document explains how the commit message generator automatically detects which module a file belongs to based on its path structure.

## Overview

The module detection system uses **pattern-based path analysis** to extract module names from file paths, eliminating the need for hardcoded mappings.

## Detection Flow

```mermaid
flowchart TD
    Start([File Path]) --> Normalize[Normalize Path Separators]
    Normalize --> CheckAuto{Starts with<br/>automation/?}

    CheckAuto -->|Yes| ExtractAuto[Extract automation/<module-name>]
    ExtractAuto --> ReturnAuto([Return: module-name])

    CheckAuto -->|No| CheckContainer{Starts with<br/>containers/?}
    CheckContainer -->|Yes| ExtractContainer[Extract containers/<module-name>]
    ExtractContainer --> ReturnContainer([Return: module-name])

    CheckContainer -->|No| CheckMCP{Starts with<br/>src/mcp/?}
    CheckMCP -->|Yes| ExtractMCP[Extract src/mcp/<service>]
    ExtractMCP --> PrefixMCP[Prefix with 'mcp-']
    PrefixMCP --> ReturnMCP([Return: mcp-service])

    CheckMCP -->|No| CheckVSCode{Starts with<br/>.vscode/extensions/?}
    CheckVSCode -->|Yes| ReturnVSCode([Return: vscode-ext])

    CheckVSCode -->|No| CheckContracts{Starts with<br/>contracts/?}
    CheckContracts -->|Yes| ExtractContracts[Extract contracts/<name>]
    ExtractContracts --> PrefixContracts[Prefix with 'contracts-']
    PrefixContracts --> ReturnContracts([Return: contracts-name])

    CheckContracts -->|No| CheckDocs{Starts with<br/>docs/?}
    CheckDocs -->|Yes| ReturnDocs([Return: docs])

    CheckDocs -->|No| CheckClaude{Starts with<br/>.claude/?}
    CheckClaude -->|Yes| ReturnClaude([Return: claude-config])

    CheckClaude -->|No| CheckVSConfig{Starts with<br/>.vscode/?}
    CheckVSConfig -->|Yes| ReturnVSConfig([Return: vscode-config])

    CheckVSConfig -->|No| CheckRoot{Root file?}
    CheckRoot -->|.md| ReturnRootDocs([Return: docs])
    CheckRoot -->|Config file| ReturnRepoConfig([Return: repo-config])
    CheckRoot -->|Other| ReturnUnknown([Return: unknown])

    style ReturnAuto fill:#90EE90
    style ReturnContainer fill:#90EE90
    style ReturnMCP fill:#87CEEB
    style ReturnVSCode fill:#FFB6C1
    style ReturnContracts fill:#DDA0DD
    style ReturnDocs fill:#F0E68C
    style ReturnClaude fill:#FFA07A
    style ReturnVSConfig fill:#FFA07A
    style ReturnRootDocs fill:#F0E68C
    style ReturnRepoConfig fill:#D3D3D3
    style ReturnUnknown fill:#FF6B6B
```

## Module Type Classification

```mermaid
graph TB
    subgraph Deployable["Deployable Units (Versioned)"]
        MCP["MCP Servers<br/>src/mcp/*/<br/>→ mcp-*"]
        VSCodeExt["VSCode Extension<br/>.vscode/extensions/*/<br/>→ vscode-ext"]
    end

    subgraph Infrastructure["Infrastructure (Type-Prefixed)"]
        AutoShell["Shell Scripts<br/>automation/sh-*/<br/>→ sh-*"]
        AutoPwsh["PowerShell<br/>automation/pwsh-*/<br/>→ pwsh-*"]
        AutoPy["Python<br/>automation/py-*/<br/>→ py-*"]
        Containers["Containers<br/>containers/*/<br/>→ container-name"]
    end

    subgraph Supporting["Supporting Modules"]
        Docs["Documentation<br/>docs/*/, *.md<br/>→ docs"]
        Contracts["Contracts<br/>contracts/*/<version>/<br/>→ contracts-*"]
        Claude["Claude Config<br/>.claude/<br/>→ claude-config"]
        VSCode["VSCode Config<br/>.vscode/<br/>→ vscode-config"]
        Repo["Repo Config<br/>root configs<br/>→ repo-config"]
    end

    style MCP fill:#87CEEB
    style VSCodeExt fill:#FFB6C1
    style AutoShell fill:#90EE90
    style AutoPwsh fill:#90EE90
    style AutoPy fill:#90EE90
    style Containers fill:#98FB98
    style Docs fill:#F0E68C
    style Contracts fill:#DDA0DD
    style Claude fill:#FFA07A
    style VSCode fill:#FFA07A
    style Repo fill:#D3D3D3
```

## Pattern Examples

### Pattern 1: Automation Modules

```mermaid
graph LR
    subgraph "automation/ directory"
        A1["sh-vscode/"]
        A2["pwsh-build/"]
        A3["py-test/"]
    end

    A1 --> M1["Module: sh-vscode<br/>Type: Shell"]
    A2 --> M2["Module: pwsh-build<br/>Type: PowerShell"]
    A3 --> M3["Module: py-test<br/>Type: Python"]

    style A1 fill:#90EE90
    style A2 fill:#90EE90
    style A3 fill:#90EE90
    style M1 fill:#E0FFE0
    style M2 fill:#E0FFE0
    style M3 fill:#E0FFE0
```

### Examples

| File Path                         | Detected Module | Type Inferred      |
| --------------------------------- | --------------- | ------------------ |
| `automation/sh-vscode/install.sh` | `sh-vscode`     | Shell (sh-)        |
| `automation/pwsh-build/build.ps1` | `pwsh-build`    | PowerShell (pwsh-) |
| `automation/py-test/test.py`      | `py-test`       | Python (py-)       |

### Pattern 2: Container Modules

```mermaid
graph LR
    subgraph "containers/ directory"
        C1["mkdocs/"]
        C2["nginx-proxy/"]
        C3["postgres/"]
    end

    C1 --> CM1["Module: mkdocs"]
    C2 --> CM2["Module: nginx-proxy"]
    C3 --> CM3["Module: postgres"]

    style C1 fill:#98FB98
    style C2 fill:#98FB98
    style C3 fill:#98FB98
    style CM1 fill:#E0FFE0
    style CM2 fill:#E0FFE0
    style CM3 fill:#E0FFE0
```

### Examples

| File Path                           | Detected Module |
| ----------------------------------- | --------------- |
| `containers/mkdocs/Dockerfile`      | `mkdocs`        |
| `containers/nginx-proxy/nginx.conf` | `nginx-proxy`   |
| `containers/postgres/init.sql`      | `postgres`      |

### Pattern 3: MCP Server Modules

```mermaid
graph LR
    subgraph "src/mcp/ directory"
        S1["pwsh/"]
        S2["docs/"]
        S3["github/"]
        S4["vscode/"]
    end

    S1 --> SM1["Module: mcp-pwsh"]
    S2 --> SM2["Module: mcp-docs"]
    S3 --> SM3["Module: mcp-github"]
    S4 --> SM4["Module: mcp-vscode"]

    style S1 fill:#87CEEB
    style S2 fill:#87CEEB
    style S3 fill:#87CEEB
    style S4 fill:#87CEEB
    style SM1 fill:#D4E8F0
    style SM2 fill:#D4E8F0
    style SM3 fill:#D4E8F0
    style SM4 fill:#D4E8F0
```

### Examples

| File Path                | Detected Module |
| ------------------------ | --------------- |
| `src/mcp/pwsh/main.go`   | `mcp-pwsh`      |
| `src/mcp/docs/server.go` | `mcp-docs`      |
| `src/mcp/github/api.go`  | `mcp-github`    |
| `src/mcp/vscode/main.go` | `mcp-vscode`    |

### Pattern 4: Contract Modules

```mermaid
graph LR
    subgraph "contracts/ directory"
        CT2["deployable-units/0.1.0/"]
        CT3["schemas/1.0.0/"]
    end

    CT2 --> CTM2["Module: contracts-deployable-units"]
    CT3 --> CTM3["Module: contracts-schemas"]

    style CT1 fill:#DDA0DD
    style CT2 fill:#DDA0DD
    style CT3 fill:#DDA0DD
    style CTM1 fill:#F0E0F8
    style CTM2 fill:#F0E0F8
    style CTM3 fill:#F0E0F8
```

### Examples

| File Path                                       | Detected Module              |
| ----------------------------------------------- | ---------------------------- |
| `contracts/deployable-units/0.1.0/mcp-pwsh.yml` | `contracts-deployable-units` |

## File-to-Module Mapping

```mermaid
graph TD
    File[File Change Detected] --> Parse[Parse Git Status]
    Parse --> Normalize[Normalize Status<br/>added/modified/deleted/renamed]
    Normalize --> Extract[Extract File Path]
    Extract --> Detect[determineFileModule]

    Detect --> Result[FileChange Object]

    Result --> Field1["Status: normalized"]
    Result --> Field2["FilePath: path/to/file"]
    Result --> Field3["Module: detected-module"]

    Field1 --> Table[Markdown Table]
    Field2 --> Table
    Field3 --> Table

    Table --> Commit[Commit Message]

    style File fill:#FFE4B5
    style Result fill:#98FB98
    style Table fill:#87CEEB
    style Commit fill:#FFB6C1
```

## Module Categories

```mermaid
mindmap
  root((Modules))
    Deployable
      mcp-pwsh
      mcp-docs
      mcp-github
      mcp-vscode
      vscode-ext
    Infrastructure
      sh-vscode
      pwsh-build
      mkdocs
      nginx-proxy
    Configuration
      claude-config
      vscode-config
      repo-config
    Documentation
      docs
    Contracts
      contracts-repository
      contracts-deployable-units
```

## Commit Message Assembly

```mermaid
sequenceDiagram
    participant Git as Git Working Tree
    participant Parser as Status Parser
    participant Detector as Module Detector
    participant Formatter as Table Formatter
    participant Claude as Claude AI
    participant Output as Commit Message

    Git->>Parser: git status --porcelain
    Parser->>Parser: Parse XY status codes
    Parser->>Parser: Normalize to 4 categories

    loop For each file
        Parser->>Detector: File path
        Detector->>Detector: Pattern matching
        Detector-->>Parser: Module name
    end

    Parser->>Formatter: FileChange objects
    Formatter->>Formatter: Calculate column widths
    Formatter-->>Claude: Formatted markdown table

    Note over Claude: Analyze changes<br/>Generate semantic commits<br/>Apply 50/72 rule

    Claude->>Output: # Revision SHA<br/>Summary<br/>Table<br/>Module sections
```

## Example: Multi-Module Commit

```mermaid
graph TB
    subgraph "Changed Files"
        F1["automation/pwsh-build/build.ps1"]
        F2["containers/mkdocs/Dockerfile"]
        F3["src/mcp/vscode/main.go"]
        F4["docs/guide/module-detection.md"]
    end

    F1 --> M1["pwsh-build"]
    F2 --> M2["mkdocs"]
    F3 --> M3["mcp-vscode"]
    F4 --> M4["docs"]

    M1 --> CM["Commit Message"]
    M2 --> CM
    M3 --> CM
    M4 --> CM

    CM --> S1["pwsh-build: feat: ...]
    CM --> S2["mkdocs: chore: ...]
    CM --> S3["mcp-vscode: feat: ...]
    CM --> S4["docs: docs: ...]

    style F1 fill:#FFE4B5
    style F2 fill:#FFE4B5
    style F3 fill:#FFE4B5
    style F4 fill:#FFE4B5
    style M1 fill:#90EE90
    style M2 fill:#98FB98
    style M3 fill:#87CEEB
    style M4 fill:#F0E68C
    style CM fill:#FFB6C1
```

**Result:**

```text
# Revision abc123...

This commit adds automated build tooling, updates container images,
enhances the VSCode MCP server, and documents the module detection
system. These changes improve developer productivity and system
maintainability across the mono-repository.

| Status   | File                                    | Module      |
| -------- | --------------------------------------- | ----------- |
| added    | automation/pwsh-build/build.ps1         | pwsh-build  |
| modified | containers/mkdocs/Dockerfile            | mkdocs      |
| modified | src/mcp/vscode/main.go                  | mcp-vscode  |
| added    | docs/guide/module-detection.md          | docs        |

---

pwsh-build: feat: add CI/CD build automation

Implements automated build pipeline with
artifact generation and test execution.

---

mkdocs: chore: update base image to Python 3.11

Updates base image for improved performance
and security patches.

---

mcp-vscode: feat: add intelligent module detection

Implements path-based module extraction for
automatic module identification from file paths.

---

docs: docs: add module detection guide

Documents the intelligent module detection
system with visual flowcharts and examples.
```

## Benefits

```mermaid
graph LR
    A[Intelligent Detection] --> B1[Automatic Discovery]
    A --> B2[Type Inference]
    A --> B3[Scalability]
    A --> B4[Precision]

    B1 --> C1[No code changes<br/>for new modules]
    B2 --> C2[sh-/pwsh-/py-<br/>prefixes show type]
    B3 --> C3[Grows naturally<br/>with repository]
    B4 --> C4[Specific commit<br/>prefixes]

    style A fill:#FFB6C1
    style B1 fill:#90EE90
    style B2 fill:#87CEEB
    style B3 fill:#DDA0DD
    style B4 fill:#F0E68C
```

### Before vs After

```mermaid
graph TB
    subgraph "Before: Hardcoded"
        B1["automation/sh-vscode/install.sh"]
        B1 --> BM1["infra"]

        style B1 fill:#FFE4B5
        style BM1 fill:#FF6B6B
    end

    subgraph "After: Intelligent"
        A1["automation/sh-vscode/install.sh"]
        A1 --> AM1["sh-vscode"]
        AM1 --> AT1["Type: Shell<br/>Module: sh-vscode"]

        style A1 fill:#FFE4B5
        style AM1 fill:#90EE90
        style AT1 fill:#E0FFE0
    end
```

**Commit Comparison:**

❌ **Before (Generic):**

```text
infra: feat: add installation script
```

✅ **After (Specific):**

```text
sh-vscode: feat: add automated installation

Implements shell script for automated
VSCode extension installation and setup.
```

## Implementation Reference

- **Source Code:** `src/mcp/vscode/main.go:499-594`
- **Tests:** `src/mcp/vscode/module_test.go`
- **Documentation:** [Repository Layout](../reference/trunk/repository-layout.md)

## Adding New Modules

```mermaid
flowchart LR
    Create[Create Directory] --> Auto{Module Type?}

    Auto -->|Automation| Shell["automation/<prefix>-<name>/"]
    Auto -->|Container| Container["containers/<name>/"]
    Auto -->|MCP Server| MCP["src/mcp/<service>/"]
    Auto -->|Extension| Ext[".vscode/extensions/<name>/"]

    Shell --> Detect1[Automatically detected]
    Container --> Detect2[Automatically detected]
    MCP --> Detect3[Automatically detected]
    Ext --> Detect4[Automatically detected]

    Detect1 --> Use[Use in commits]
    Detect2 --> Use
    Detect3 --> Use
    Detect4 --> Use

    style Create fill:#FFB6C1
    style Shell fill:#90EE90
    style Container fill:#98FB98
    style MCP fill:#87CEEB
    style Ext fill:#FFB6C1
    style Use fill:#F0E68C
```

### Examples

```bash
# 1. Create directory
mkdir -p automation/py-test

# 2. Add files
echo "#!/usr/bin/env python3" > automation/py-test/run_tests.py

# 3. Commit (module automatically detected as "py-test")
git add automation/py-test/
# Click robot button in VSCode
# Generated commit will use: "py-test: feat: ..."
```

No code changes required! 🎉
