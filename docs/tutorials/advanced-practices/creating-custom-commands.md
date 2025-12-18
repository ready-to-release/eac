# Creating Custom Commands

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Module](../getting-started/first-module.md), Go programming knowledge

## Planned Content

This tutorial teaches you how to extend the r2r CLI with custom commands, integrate with MCP servers, and automate project-specific workflows.

### What You'll Learn

- Understand the r2r CLI extension architecture
- Create custom commands in Go
- Register commands with the CLI
- Access repository contracts and configuration
- Implement command logic and validation
- Write tests for custom commands
- Package commands as extensions
- Integrate with MCP protocol

### Tutorial Structure

1. **Understanding CLI architecture**
   - Command structure and registration
   - Extension points and hooks
   - Command context and dependencies
   - Configuration and contracts access

2. **Creating your first command**
   - Command structure (Cobra-based)
   - Implement `run-checks` command
   - Parse flags and arguments
   - Access repository configuration

3. **Implementing command logic**
   - Read module contracts
   - Execute validations
   - Format output (table, JSON, markdown)
   - Handle errors gracefully

4. **Testing custom commands**
   - Unit tests for command logic
   - Integration tests with real contracts
   - Mock dependencies
   - Test command output

5. **Advanced command patterns**
   - Commands with subcommands
   - Interactive commands (prompts)
   - Long-running operations
   - Progress reporting

6. **Packaging as extension**
   - Extension metadata
   - Distribution (binary, source, container)
   - Installation and discovery
   - Versioning extensions

7. **MCP server integration**
   - Implement MCP protocol
   - Expose commands via MCP
   - Tool definitions and schemas
   - Connect to Claude Code / IDEs

### Example: Custom `run-checks` Command

The tutorial creates a custom command for pre-deployment checks:

**Command usage:**

```bash
r2r run-checks --module api-gateway --env production
```

**Implementation structure:**

```text
go/eac-extensions/
├── commands/
│   └── checks/
│       ├── run_checks.go       # Command implementation
│       ├── run_checks_test.go  # Unit tests
│       └── validators.go       # Check logic
├── main.go                     # Extension entry point
└── go.mod
```

**Command implementation:**

```go
package checks

import (
    "github.com/spf13/cobra"
    "github.com/ready-to-release/eac/pkg/contracts"
)

func NewRunChecksCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "run-checks",
        Short: "Run pre-deployment checks",
        RunE:  runChecks,
    }

    cmd.Flags().String("module", "", "Module to check")
    cmd.Flags().String("env", "", "Target environment")

    return cmd
}

func runChecks(cmd *cobra.Command, args []string) error {
    module := cmd.Flag("module").Value.String()
    env := cmd.Flag("env").Value.String()

    // Load contracts
    repo, err := contracts.LoadRepository()
    if err != nil {
        return err
    }

    // Run checks
    checks := []Check{
        CheckTestsPassing(module),
        CheckArtifactsExist(module),
        CheckSecurityScans(module),
        CheckEnvironmentReady(env),
    }

    // Execute and report
    results := ExecuteChecks(checks)
    PrintResults(results)

    return nil
}
```

### Key Concepts Covered

- Cobra command framework
- r2r contract system
- Command registration and discovery
- Testing CLI commands
- Extension packaging
- MCP protocol integration

### Command Development Workflow

1. **Design command interface**
   - Define command name and flags
   - Plan command output
   - Consider error cases

2. **Implement command logic**
   - Create command struct
   - Implement RunE function
   - Access contracts and config

3. **Write tests**
   - Unit test command logic
   - Integration test with real data
   - Test error handling

4. **Register command**
   - Add to extension registry
   - Build and install extension

5. **Document command**
   - Add help text
   - Create usage examples
   - Document in reference section

### Testing Custom Commands

```go
func TestRunChecks(t *testing.T) {
    tests := []struct {
        name    string
        module  string
        env     string
        wantErr bool
    }{
        {
            name:    "valid module and env",
            module:  "api-gateway",
            env:     "staging",
            wantErr: false,
        },
        {
            name:    "invalid module",
            module:  "nonexistent",
            env:     "staging",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewRunChecksCommand()
            cmd.SetArgs([]string{
                "--module", tt.module,
                "--env", tt.env,
            })

            err := cmd.Execute()
            if (err != nil) != tt.wantErr {
                t.Errorf("want error %v, got %v", tt.wantErr, err)
            }
        })
    }
}
```

### MCP Integration

Expose custom commands via MCP server:

```go
// Implement MCP protocol
type CommandsMCPServer struct {
    commands []*cobra.Command
}

func (s *CommandsMCPServer) ListTools() []Tool {
    tools := []Tool{}
    for _, cmd := range s.commands {
        tools = append(tools, Tool{
            Name:        cmd.Use,
            Description: cmd.Short,
            Schema:      generateSchema(cmd),
        })
    }
    return tools
}

func (s *CommandsMCPServer) ExecuteTool(name string, args map[string]interface{}) (string, error) {
    cmd := findCommand(s.commands, name)
    return executeCommand(cmd, args)
}
```

### Use Cases for Custom Commands

- **Project-specific workflows**: `r2r deploy-to-k8s`
- **Quality checks**: `r2r run-checks`
- **Report generation**: `r2r generate-compliance-report`
- **Integration with tools**: `r2r sync-jira`
- **Automation**: `r2r batch-update-deps`

### Best Practices

- Follow Cobra conventions
- Provide comprehensive help text
- Validate inputs early
- Use structured output (JSON, table)
- Write tests for all commands
- Document in reference section
- Version extensions semantically
- Handle errors gracefully

### Distribution Options

1. **Source code**: Users build locally
2. **Binary**: Pre-built for platforms
3. **Container**: Docker image
4. **Plugin**: Dynamic loading

### Next Steps

After completing this tutorial, you've mastered advanced r2r CLI practices. Explore [Specialized Topics](../specialized-topics/) for deep dives into BDD techniques, architecture documentation, and security scanning.
