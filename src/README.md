# Source Libraries

This directory contains reusable Go libraries for the project.

## Available Libraries

### 1. Contracts (`src/contracts`)

Library for loading and managing module contracts from YAML files.

**Purpose:** Type-safe contract loading and validation

**Features:**

- Load contracts from disk
- Parse YAML contract files
- Type-safe contract structures
- Error handling with wrapped errors

**Module:** `github.com/ready-to-release/eac/src/contracts`

### 2. Contracts/Modules (`src/contracts/modules`)

Module-specific contract handling with glob pattern matching.

**Purpose:** Module contract registry and file matching

**Features:**

- Load all module contracts from `contracts/modules/{version}/*.yml`
- Registry pattern for O(1) lookups
- Full glob pattern support (`**/*.ext`, `?`, `[abc]`, etc.)
- Dependency graph analysis
- File pattern matching

**Module:** `github.com/ready-to-release/eac/src/contracts/modules`

### 3. Repository (`src/repository`)

Git repository operations and file listing.

**Purpose:** Git-aware file operations

**Features:**

- Find repository root from any directory
- List all repository files (tracked/untracked)
- Respect `.gitignore` rules
- Cross-platform path handling

**Module:** `github.com/ready-to-release/eac/src/repository`

---

## Calling Library Functions from Shell

### Prerequisites

- **Go 1.24.4+** installed
- Git repository (for repository library)

### Quick Start - Call Any Function

Create a simple Go file that calls the function, then run it:

```bash
# Example: Call repository.GetRepositoryRoot()
cd out

# Create wrapper file
cat > call-function.go << 'EOF'
package main
import ("fmt"; "github.com/ready-to-release/eac/src/repository")
func main() {
    root, err := repository.GetRepositoryRoot("")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(root)
}
EOF

# Run it
go run call-function.go
```

### One-Liner Method

Use `go run -` to execute inline Go code:

```bash
cd out

# Call any function directly
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/repository")
func main() {
    root, _ := repository.GetRepositoryRoot("")
    fmt.Println(root)
}' | go run -
```

---

## Library-Specific Instructions

### Contracts Library

**Location:** `src/contracts/`

#### How to Call: `contracts.NewLoader()`

```bash
cd out

# Create wrapper
cat > load-contract.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/contracts"
)
func main() {
    loader := contracts.NewLoader("..")
    var contract contracts.BaseContract
    err := loader.LoadYAML("contracts/modules/0.1.0/src-cli.yml", &contract)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Loaded: %s\n", contract.GetName())
    fmt.Printf("Type: %s\n", contract.GetType())
    fmt.Printf("Root: %s\n", contract.GetRoot())
}
EOF

# Run it
go run load-contract.go
```

#### One-Liner

```bash
cd out
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/contracts")
func main() {
    loader := contracts.NewLoader("..")
    var c contracts.BaseContract
    loader.LoadYAML("contracts/modules/0.1.0/src-cli.yml", &c)
    fmt.Println(c.GetName())
}' | go run -
```

#### Run Tests (Optional)

```bash
cd src/contracts
go test -v
```

**Dependencies:** `gopkg.in/yaml.v3`

---

### Contracts/Modules Library

**Location:** `src/contracts/modules/`

#### How to Call: `modules.LoadFromWorkspace()`

```bash
cd out

# Create wrapper
cat > load-modules.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/contracts/modules"
)
func main() {
    registry, err := modules.LoadFromWorkspace("..", "0.1.0")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Loaded %d modules\n", registry.Count())

    // List all modules
    for _, moniker := range registry.AllMonikers() {
        fmt.Printf("  - %s\n", moniker)
    }
}
EOF

# Run it
go run load-modules.go
```

#### How to Call: `module.MatchesFile()`

```bash
cd out

cat > test-pattern.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/contracts/modules"
)
func main() {
    registry, _ := modules.LoadFromWorkspace("..", "0.1.0")
    module, _ := registry.Get("src-mcp-vscode")

    // Test file matching
    testFiles := []string{
        "src/mcp/vscode/main.go",
        "src/mcp/vscode/test/unit.go",
        "docs/README.md",
    }

    for _, file := range testFiles {
        matches := module.MatchesFile(file)
        fmt.Printf("%v: %s\n", matches, file)
    }

    // Show patterns
    fmt.Println("\nPatterns:", module.GetGlobPatterns())
}
EOF

go run test-pattern.go
```

#### One-Liner

```bash
cd out
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/contracts/modules")
func main() {
    r, _ := modules.LoadFromWorkspace("..", "0.1.0")
    fmt.Printf("Loaded %d modules\n", r.Count())
}' | go run -
```

#### Run Tests (Optional)

```bash
cd src/contracts/modules
go test -v
```

**Dependencies:** `gopkg.in/yaml.v3`, `github.com/gobwas/glob`

**Supported Glob Patterns:**

- `**/*.ext` - All files with extension
- `src/**/test/*.go` - Multiple `**` segments
- `file?.go` - Single char wildcard
- `file[abc].go` - Character classes

---

### Repository Library

**Location:** `src/repository/`

#### How to Call: `repository.GetRepositoryRoot()`

```bash
cd out

# Create wrapper
cat > get-root.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/repository"
)
func main() {
    root, err := repository.GetRepositoryRoot("")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(root)
}
EOF

# Run it
go run get-root.go
```

#### How to Call: `repository.GetRepositoryFiles()`

```bash
cd out

cat > get-files.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/repository"
)
func main() {
    files, err := repository.GetRepositoryFiles(true, false, "")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Printf("Found %d files\n", len(files))

    // Show first 10
    for i, file := range files {
        if i >= 10 { break }
        fmt.Printf("  %s\n", file.Path)
    }
}
EOF

go run get-files.go
```

#### How to Call: `repository.IsGitRepository()`

```bash
cd out

cat > check-repo.go << 'EOF'
package main
import (
    "fmt"
    "github.com/ready-to-release/eac/src/repository"
)
func main() {
    isRepo := repository.IsGitRepository("..")
    fmt.Printf("Is git repository: %v\n", isRepo)
}
EOF

go run check-repo.go
```

#### One-Liners

```bash
cd out

# Get repository root
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/repository")
func main() {
    root, _ := repository.GetRepositoryRoot("")
    fmt.Println(root)
}' | go run -

# Count files
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/repository")
func main() {
    files, _ := repository.GetRepositoryFiles(true, false, "")
    fmt.Printf("Found %d files\n", len(files))
}' | go run -

# Check if directory is a git repo
echo 'package main
import ("fmt"; "github.com/ready-to-release/eac/src/repository")
func main() {
    fmt.Println(repository.IsGitRepository(".."))
}' | go run -
```

#### Run Tests (Optional)

```bash
cd src/repository
go test -v
```

**Dependencies:** None (standard library only)

---

## Running Demo Programs

### Demo Programs Location

`/out/` directory contains various demo programs:

```bash
cd out

# Demo: Repository file listing
go run demo-get-repository-files.go

# Demo: Contract loading
go run test-contracts-loader.go

# Demo: Repository operations
go run test-repository.go

# Demo: Simple repository usage
go run test-repository-simple.go

# Demo: Pattern matching validation
go run test-actual-contract-patterns.go

# Demo: Fixed glob patterns
go run demo-fixed-patterns.go
```

### Create Your Own Test Program

**Template:**

```go
package main

import (
    "fmt"
    "github.com/ready-to-release/eac/src/contracts/modules"
    "github.com/ready-to-release/eac/src/repository"
)

func main() {
    // Your test code here

    // Example: Load contracts
    registry, err := modules.LoadFromWorkspace("..", "0.1.0")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Loaded %d modules\n", registry.Count())

    // Example: Get repository files
    files, err := repository.GetRepositoryFiles(true, false, "..")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Found %d files\n", len(files))
}
```

**Save as:** `out/my-test.go`

**Run:**

```bash
cd out
go run my-test.go
```

---

## Running All Tests

### Test All Libraries

```bash
# From project root
go test ./src/contracts/... -v
go test ./src/repository/... -v

# Or test everything
go test ./src/... -v
```

### Test with Coverage

```bash
# Individual library
cd src/contracts
go test -cover

# All libraries with detailed coverage
go test ./src/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Summary

```bash
# Quick test summary
go test ./src/...

# Expected output:
# ok      github.com/ready-to-release/eac/src/contracts          0.385s
# ok      github.com/ready-to-release/eac/src/contracts/modules  0.284s
# ok      github.com/ready-to-release/eac/src/repository         1.234s
```

---

## Building Executables

### Build Demo Programs

```bash
cd out

# Build specific demo
go build -o demo.exe demo-get-repository-files.go

# Run built executable
./demo.exe -stats
```

### Build with Dependencies

```bash
# Ensure dependencies are downloaded
cd out
go mod tidy

# Build
go build -o my-tool.exe my-program.go

# Run
./my-tool.exe
```

---

## Troubleshooting

### Missing Dependencies

**Error:** `cannot find package`

**Solution:**

```bash
cd src/contracts
go mod tidy
go get github.com/gobwas/glob
```

### Module Path Issues

**Error:** `module declares its path as X but was required as Y`

**Solution:**

```bash
# Check go.mod has correct module path
cat src/contracts/go.mod

# Should show:
# module github.com/ready-to-release/eac/src/contracts
```

### Test Failures

**Error:** Tests fail due to missing files

**Solution:**

```bash
# Make sure you're in a git repository
git status

# Repository tests require git to be initialized
git init  # If not already initialized
```

### Import Issues

**Error:** `could not import package`

**Solution:**

```bash
# From the directory using the library
go mod tidy

# Add replace directive if needed (for local development)
go mod edit -replace github.com/ready-to-release/eac/src/contracts=../src/contracts
```

---

## Library Versions

| Library           | Module Path                                             | Version | Dependencies  |
| ----------------- | ------------------------------------------------------- | ------- | ------------- |
| contracts         | `github.com/ready-to-release/eac/src/contracts`         | 0.0.0   | yaml.v3       |
| contracts/modules | `github.com/ready-to-release/eac/src/contracts/modules` | 0.0.0   | yaml.v3, glob |
| repository        | `github.com/ready-to-release/eac/src/repository`        | 0.0.0   | None          |

---

## Development Workflow

### Adding New Tests

```bash
# 1. Navigate to library
cd src/contracts

# 2. Create or edit test file
# Example: loader_test.go

# 3. Run tests
go test -v -run YourTestName

# 4. Verify coverage
go test -cover
```

### Updating Dependencies

```bash
# Update all dependencies
cd src/contracts
go get -u ./...
go mod tidy

# Update specific dependency
go get -u gopkg.in/yaml.v3
```

### Cross-Library Usage

```bash
# If one library imports another, use replace directive
cd out
cat go.mod

# Should contain:
# replace github.com/ready-to-release/eac/src/contracts => ../src/contracts
# replace github.com/ready-to-release/eac/src/repository => ../src/repository
```

---

## Quick Reference

### Essential Commands

```bash
# Run all tests
go test ./src/... -v

# Test with coverage
go test ./src/... -cover

# Run specific demo
cd out && go run demo-get-repository-files.go

# Build executable
cd out && go build -o tool.exe my-program.go

# Update dependencies
cd src/contracts && go mod tidy
```

### Directory Structure

```
src/
├── contracts/           # Base contract library
│   ├── go.mod
│   ├── loader.go
│   ├── types.go
│   ├── errors.go
│   └── *_test.go
├── contracts/modules/   # Module contract library
│   ├── go.mod
│   ├── loader.go
│   ├── types.go
│   ├── registry.go
│   └── *_test.go
└── repository/          # Repository operations
    ├── go.mod
    ├── repository.go
    └── repository_test.go
```

---

## Additional Resources

### Documentation

- **Glob patterns:** `/out/GLOB-PATTERNS-REFERENCE.md`
- **Migration report:** `/out/MIGRATION-SUMMARY.md`
- **Library summaries:** `/out/LIBRARY-SUMMARY.md`

### Example Programs

- `/out/demo-get-repository-files.go` - Repository file listing
- `/out/test-contracts-loader.go` - Contract loading
- `/out/test-repository-simple.go` - Simple repository usage
- `/out/demo-fixed-patterns.go` - Glob pattern examples

### Test Programs

- `/out/test-actual-contract-patterns.go` - Pattern validation
- `/out/test-complex-patterns.go` - Complex pattern testing

---

**Last Updated:** 2025-11-03
**Go Version:** 1.24.4+
**Status:** Production Ready
