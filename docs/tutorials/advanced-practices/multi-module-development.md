# Multi-Module Development

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Estimated time:** 45 minutes
**Prerequisites:** [Your First Module](../getting-started/first-module.md), [Building and Testing Changes](../core-workflows/building-and-testing.md)

## Planned Content

This tutorial teaches you how to work effectively across multiple modules, managing dependencies, understanding build order, and maintaining module contracts.

### What You'll Learn

- Understand module dependency graphs
- View and navigate module dependencies
- Define module contracts with dependencies
- Determine correct build order
- Handle circular dependencies
- Validate module hierarchy
- Work across module boundaries
- Integration testing between modules

### Tutorial Structure

1. **Understanding module dependencies**
   - Dependency types: build, test, runtime
   - Direct vs. transitive dependencies
   - View dependencies: `r2r show dependencies <module>`
   - Visualize dependency graph

2. **Module contracts**
   - Define dependencies in `.r2r/eac/modules.yml`
   - Dependency syntax and semantics
   - Contract validation
   - Breaking vs. non-breaking changes

3. **Build order and execution**
   - Topological sort of dependency graph
   - Get execution order: `r2r get execution-order <modules>`
   - Build dependencies: `r2r get build-deps <module>`
   - Parallel vs. sequential builds

4. **Example: Multi-module feature**
   - Scenario: Add authentication to API
   - Modules affected: auth-lib, api-gateway, user-service
   - Define dependency chain
   - Build in correct order
   - Integration testing

5. **Working across module boundaries**
   - Identify module ownership: `r2r show files`
   - Understand API contracts between modules
   - Make compatible changes
   - Version module interfaces

6. **Validating dependencies**
   - Check contracts: `r2r validate dependencies`
   - Verify hierarchy: `r2r validate module-hierarchy`
   - Go module validation: `r2r validate go-tidy`
   - Detect circular dependencies

7. **Handling dependency changes**
   - Update contracts when dependencies change
   - Rebuild dependent modules
   - Integration test suite
   - Contract testing patterns

8. **Circular dependencies**
   - Why circular dependencies are problematic
   - Detecting circular dependencies
   - Refactoring to break cycles
   - Dependency inversion principle

### Example: Multi-Module Feature

The tutorial implements authentication across three modules:

**auth-lib** (base library):
- Provides JWT token validation
- No dependencies on other modules

**api-gateway** (depends on auth-lib):
- Uses auth-lib for authentication
- Routes requests to services

**user-service** (depends on auth-lib):
- Uses auth-lib to verify requests
- Manages user data

**Dependency graph:**
```text
auth-lib
├── api-gateway
└── user-service
```

**Build order:**
1. Build auth-lib first
2. Build api-gateway and user-service in parallel
3. Integration test all three together

### Commands Demonstrated

```bash
# View module dependencies
r2r show dependencies api-gateway

# Get build order for multiple modules
r2r get execution-order auth-lib api-gateway user-service

# Build in dependency order
r2r build auth-lib
r2r build api-gateway user-service

# Validate dependency contracts
r2r validate dependencies

# Check for circular dependencies
r2r validate module-hierarchy
```

### Key Concepts Covered

- Dependency graphs and topological sorting
- Module contracts and versioning
- Build order determination
- Cross-module integration testing
- Circular dependency detection
- Contract-based development

### Module Contract Example

```yaml
modules:
  - name: api-gateway
    type: go-app
    root: go/api-gateway
    dependencies:
      - auth-lib
    artifacts:
      - bin/api-gateway

  - name: auth-lib
    type: go-lib
    root: go/auth-lib
    dependencies: []
    artifacts: []
```

### Best Practices

- Keep dependency graphs shallow (avoid deep chains)
- Make dependencies explicit in contracts
- Avoid circular dependencies
- Use interface contracts between modules
- Integration test at module boundaries
- Version module APIs semantically
- Document breaking changes

### Common Patterns

**Layered architecture:**
```text
infrastructure (base)
├── domain (business logic)
│   └── application (use cases)
│       └── api (HTTP/gRPC)
```

**Shared library pattern:**
```text
common-lib (shared utilities)
├── service-a
├── service-b
└── service-c
```

**Microservices pattern:**
```text
(independent modules, communication via APIs)
service-a  service-b  service-c
```

### Integration Testing

Testing across module boundaries:

1. **Contract tests**: Verify API contracts
2. **Integration tests**: Test module interactions
3. **End-to-end tests**: Full system scenarios

Use test level tags:
- L0-L1: Unit tests (within module)
- L2: Component tests (module as a whole)
- L3: Integration tests (multiple modules)
- L4: System tests (full system)

### Next Steps

After completing this tutorial, you'll be able to manage complex multi-module systems. Continue to [Creating Custom Commands](./creating-custom-commands.md) to learn how to extend the r2r CLI.

{{ diataxis_footer() }}
