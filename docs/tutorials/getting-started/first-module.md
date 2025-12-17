# Your First Module
**Status:** Placeholder - Content coming soon

**Prerequisites:** [Quick Start Guide](./quick-start.md), Go 1.21+ installed

## Planned Content

This tutorial will teach you how to create a complete Go module from scratch in an r2r monorepo.

### What You'll Learn

- Create a simple Go module (`hello-world-service`)
- Write table-driven unit tests (`*_test.go`)
- Define the module contract in `.r2r/eac/repository.yml`
- Understand module types, dependencies, and artifacts
- Build and test with `r2r build` and `r2r test`
- Verify module registration with `r2r show modules`

### Tutorial Structure

1. **Set up module directory structure**
   - Create `go/hello-world-service/` directory
   - Initialize Go module with `go mod init`
   - Understand Go module organization

2. **Write the implementation**
   - Create `service.go` with a simple greeting function
   - Follow Go idiomatic conventions
   - Keep it simple and focused

3. **Write unit tests**
   - Create `service_test.go`
   - Implement table-driven tests
   - Run tests with `go test`

4. **Register the module**
   - Add module contract to `.r2r/eac/repository.yml`
   - Define module type, dependencies, and build configuration
   - Understand contract schema

5. **Build and test with r2r**
   - Build: `r2r build hello-world-service`
   - Test: `r2r test hello-world-service`
   - Inspect artifacts: `r2r show artifacts hello-world-service`

6. **Understand module metadata**
   - Module types (go-app, go-lib, go-tool, etc.)
   - Dependencies between modules
   - Build artifacts and their purpose

### Example Module

The tutorial will create a simple greeting service:

```go
package greet

func Greet(name string) string {
    if name == "" {
        return "Hello, Guest!"
    }
    return fmt.Sprintf("Hello, %s!", name)
}
```

With comprehensive table-driven tests and proper error handling.

### Key Concepts Covered

- Go module structure in monorepo
- Module contracts and registration
- Table-driven testing pattern
- Build artifacts and dependencies
- r2r build and test commands

### Next Steps

After completing this tutorial, you'll understand module fundamentals. Continue to [Understanding Test Suites](./understanding-test-suites.md) to learn about test levels and when to use each.
