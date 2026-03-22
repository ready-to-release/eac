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
eac init

# Initialize with project name
eac init --name my-project
```

### Verifying Initialization

```bash
# Check configuration
eac show config

# Validate setup
eac validate
```

## See Also

- [Getting Started Guide](../../../tutorials/getting-started/quick-start.md)
- [show config](../commands/show/config.md) - View configuration
- [validate](../commands/validate/validate.md) - Validate configuration
