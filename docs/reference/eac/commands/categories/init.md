# Init Commands

## Overview

Init commands set up EAC project configuration and initialize repository structure.

**Key Characteristics**:

- Creates required configuration files
- Sets up directory structure
- Initializes contracts and schemas

**When to use**: When setting up a new EAC-managed repository or adding EAC to an existing project.

## All Init Commands

<!-- book:category-commands init -->

## Common Workflows

### Initializing a New Project

```bash
# Initialize with defaults
r2r eac init

# Initialize with project name
r2r eac init --name my-project
```

### Verifying Initialization

```bash
# Check configuration
r2r eac show config

# Validate setup
r2r eac validate
```

## See Also

- [Getting Started Guide](../../../../tutorials/getting-started/quick-start.md)
- [show config](../show/config.md) - View configuration
- [validate](../validate/validate.md) - Validate configuration
