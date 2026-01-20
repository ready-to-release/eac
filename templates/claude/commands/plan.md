---
name: plan
description: Create implementation plan using MCP-powered discovery
model: claude-sonnet-4-5
thinking: extended
color: green
---

# Planning Command

Create a detailed implementation plan using MCP tools to discover project structure.

## What This Does

1. **Discover** current system state using MCP tools
2. **Analyze** dependencies and impacts
3. **Design** proposed changes
4. **Document** implementation plan
5. **Save** plan to `out/<feature-name>-plan.md`

## MCP Tools Used

This command leverages EAC MCP tools for discovery:

```bash
get-modules              # Understand module structure
get-dependencies <mod>   # Map dependency relationships
show-modules             # View module metadata
get-files-by-module <mod> # Find relevant source files
validate-contracts       # Check current state
```

## When to Use

- Before starting any non-trivial feature
- Planning refactoring work
- Evaluating architectural changes
- Documenting technical decisions

## Workflow

### Step 1: Understand Request
Clarify requirements and constraints.

### Step 2: MCP Discovery
```bash
# Discover structure
get-modules

# Map dependencies for affected modules
get-dependencies <module-name>

# Find relevant files
get-files-by-module <module-name>

# Check current contracts
validate-contracts
```

### Step 3: Design Solution
Based on MCP discovery, design changes that:
- Minimize impact on existing modules
- Maintain clean boundaries
- Follow established patterns

### Step 4: Save Plan
Write plan to `out/<feature-name>-plan.md`

## Output Format

```markdown
## Feature: <name>

### Current State (from MCP Discovery)
- Module structure (output of `get-modules`)
- Dependencies (output of `get-dependencies`)
- Key interfaces
- Relevant files (from `get-files-by-module`)

### Proposed Changes
- Modified modules
- New interfaces
- Dependency updates

### Implementation Steps
1. Step-by-step migration
2. MCP commands to verify each step
3. Testing strategy
4. Rollback plan

### MCP Verification Commands
```bash
# Commands to verify implementation
validate-module-hierarchy
build <module>
test <module>
validate-contracts
```

### Risks & Mitigations
- Identified risks
- Mitigation strategies
```

**Output Location**: `out/<feature-name>-plan.md`

## Example Usage

**User**: "Plan adding authentication middleware"

**MCP Discovery**:
```bash
get-modules                    # Find API and auth modules
get-dependencies api-module    # See what depends on API
get-files-by-module api-module # Locate API source files
```

**Output**: `out/authentication-middleware-plan.md` with:
- Current module structure (from MCP)
- Dependency impact analysis (from MCP)
- Proposed changes
- Implementation steps with MCP verification commands
