# TypeScript Module Setup

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Module](../getting-started/first-module.md), Node.js and npm installed

## Planned Content

This tutorial teaches you how to set up and work with TypeScript/npm modules in the r2r monorepo.

### What You'll Learn

- Create TypeScript modules in the monorepo
- Configure `package.json` and `tsconfig.json`
- Register npm modules in `.r2r/eac/repository.yml`
- Build TypeScript modules with `r2r build`
- Test TypeScript modules with `r2r test`
- Manage npm dependencies between modules
- Integrate with Go modules in the same repo

### Tutorial Structure

1. **Understanding npm module support**
   - Module types: npm-app, npm-lib, npm-tool
   - Build process: TypeScript → JavaScript
   - Artifact generation
   - Integration with monorepo

2. **Creating a TypeScript module**
   - Directory structure: `go/my-ts-app/`
   - Initialize npm: `npm init`
   - Install TypeScript: `npm install -D typescript`
   - Configure `tsconfig.json`

3. **Module contract configuration**
   - Add to `.r2r/eac/repository.yml`
   - Define module type (npm-app, npm-lib)
   - Specify dependencies
   - Configure build settings

4. **Writing TypeScript code**
   - Create `src/` directory
   - Write TypeScript implementation
   - Follow TypeScript best practices
   - Type definitions and interfaces

5. **Testing TypeScript modules**
   - Install test framework (Jest, Vitest)
   - Write unit tests
   - Configure test scripts
   - Run tests: `r2r test <module>`

6. **Building TypeScript modules**
   - Build command: `r2r build <module>`
   - Compiles TypeScript to JavaScript
   - Generates artifacts
   - Understand build output

7. **Managing dependencies**
   - Internal dependencies (other modules)
   - External dependencies (npm packages)
   - Lock files (package-lock.json)
   - Dependency validation

8. **Mixed Go and TypeScript monorepo**
   - Go modules and npm modules together
   - Cross-module dependencies
   - Build order considerations
   - Integration testing

### Example: TypeScript Service Module

The tutorial creates a TypeScript service:

**Directory structure:**

```text
go/api-client/
├── src/
│   ├── index.ts
│   ├── client.ts
│   └── types.ts
├── tests/
│   └── client.test.ts
├── package.json
├── tsconfig.json
└── README.md
```

**Module contract (`.r2r/eac/repository.yml`):**

```yaml
modules:
  - name: api-client
    type: npm-lib
    root: go/api-client
    dependencies: []
    artifacts:
      - dist/index.js
      - dist/index.d.ts
```

**Package.json:**

```json
{
  "name": "@r2r/api-client",
  "version": "1.0.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint src"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "jest": "^29.0.0",
    "@types/jest": "^29.0.0"
  }
}
```

**TypeScript config:**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "declaration": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

### Key Concepts Covered

- TypeScript module configuration
- npm module types in monorepo
- Building and testing TypeScript
- Module dependencies
- Mixed language monorepo

### Building and Testing

```bash
# Register module (done once)
# Edit .r2r/eac/repository.yml

# Install dependencies
cd go/api-client
npm install

# Build module
r2r eac build api-client
# Runs: npm run build

# Test module
r2r eac test api-client
# Runs: npm test

# View artifacts
r2r eac show artifacts api-client
```

### npm Module Types

**npm-app:**

- Standalone application
- Produces executable artifact
- Example: CLI tool, web server

**npm-lib:**

- Shared library
- Consumed by other modules
- Example: API client, utilities

**npm-tool:**

- Build/development tool
- Used during build process
- Example: Code generator, linter

### Best Practices

- Use TypeScript strict mode
- Write comprehensive type definitions
- Include `.d.ts` files in artifacts
- Version lock dependencies (package-lock.json)
- Use consistent tsconfig across modules
- Write unit tests for all code
- Lint TypeScript code (ESLint)
- Keep modules focused and small

### Testing TypeScript Modules

**Jest configuration:**

```json
{
  "preset": "ts-jest",
  "testEnvironment": "node",
  "roots": ["<rootDir>/tests"],
  "testMatch": ["**/*.test.ts"]
}
```

**Example test:**

```typescript
import { APIClient } from '../src/client';

describe('APIClient', () => {
  it('should make GET requests', async () => {
    const client = new APIClient('https://api.example.com');
    const result = await client.get('/users');
    expect(result).toBeDefined();
  });
});
```

### Mixed Monorepo Example

Repository with both Go and TypeScript:

```text
go/
├── eac-commands/        # Go CLI
├── eac-web-ui/          # TypeScript React app
└── api-client/          # TypeScript library
```

**Dependencies:**

- `eac-web-ui` depends on `api-client` (TypeScript → TypeScript)
- `eac-commands` is independent (Go)

**Build order:**

1. Build `api-client` (TypeScript lib)
2. Build `eac-web-ui` (depends on api-client)
3. Build `eac-commands` (independent)

### Common Issues

**Issue: Module not found**

- Solution: Run `npm install` in module directory

**Issue: Build fails**

- Solution: Check `tsconfig.json` configuration

**Issue: Tests not running**

- Solution: Verify test script in `package.json`

**Issue: Artifact not generated**

- Solution: Check `outDir` in `tsconfig.json` matches artifacts in contract

### Integration with CI/CD

```yaml
# .github/workflows/ci.yml
jobs:
  test-typescript:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Build TypeScript modules
        run: r2r eac build api-client

      - name: Test TypeScript modules
        run: r2r eac test api-client
```

### Next Steps

After completing this tutorial, you can work with TypeScript modules in the monorepo. Explore other specialized topics based on your needs: [Effective BDD Scenarios](./effective-bdd-scenarios.md), [Architecture Documentation](./architecture-documentation.md), or [Security Scanning](./security-scanning-workflow.md).
