# Definitions Library

Shared library for merging YAML definition files from a directory structure.

## Purpose

This library walks a directory tree, finds all `definitions.yml` files, and merges them into a single YAML structure based on their directory paths.

## Usage

```go
import "github.com/ready-to-release/eac/src/repository/definitions"

// Process all definitions in a directory
merged, err := definitions.ProcessDirectory("/path/to/repo")
if err != nil {
    log.Fatal(err)
}

// Convert to YAML bytes
output, err := yaml.Marshal(merged)
```

## Directory Structure → YAML Path

Files are merged based on their directory structure:

```
/definitions.yml           → Root level
/foo/definitions.yml       → foo.*
/foo/bar/definitions.yml   → foo.bar.*
```

## Excluded Directories

The following are automatically skipped:
- `.git`
- `node_modules`
- `.vscode`
- `.idea`
- `out`
- `*/templates/*/skeleton` (template skeleton structures)

## API

### `ProcessDirectory(rootDir string) (*yaml.Node, error)`
Convenience function that enumerates and merges all definitions.

### `EnumerateDefinitionFiles(rootDir string) ([]DefinitionFile, error)`
Finds all definitions.yml files in a directory tree.

### `MergeDefinitions(definitions []DefinitionFile) (*yaml.Node, error)`
Merges multiple definition files into a single YAML structure.

## Module

```
module github.com/ready-to-release/eac/src/repository/definitions
```

This is a standalone module that can be used by any project in the monorepo.
