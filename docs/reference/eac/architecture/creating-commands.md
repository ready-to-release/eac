# Creating CLI Commands

This guide is for developers contributing new commands to the EAC CLI in `go/cli/eac/impl/`.

## Overview

The EAC CLI uses a structured comment-based system to define command metadata, help text, and flags.

This guide provides templates and best practices for creating new commands that integrate seamlessly with the CLI's help system,
command registry, and shell completion.

## Command File Header Template

Every command file must include structured comment headers that define help metadata.

Copy this template and fill in the placeholders:

```go
// Command: <parent> <subcommand>
// Short: <One sentence describing what this command does>
// Long: <First paragraph explaining the purpose and what problem it solves>
// Long: <Second paragraph describing the behavior and how it works>
// Long: <Third paragraph describing inputs, outputs, and important details>
// Flag.<flagname>: type=<type>, default=<value>, shorthand=<letter>, usage=<Educational description of what this flag does, when to use it, and why>, required=<true|false>, completion=<val1,val2,...>
package mypackage
```

### Header Field Descriptions

#### `Command:` (Required)

The command name as users will type it, with spaces separating parent and subcommand.

**Examples:**

- `build` - Standalone command
- `create spec` - Subcommand under create
- `show modules` - Subcommand under show

#### `Short:` (Required)

A single sentence (50-80 characters) that concisely describes what the command does.

This appears in command listings.

**Best Practices:**

- Start with a verb (e.g., "Generate", "Display", "Validate")
- Be specific about what the command operates on
- Avoid generic phrases like "Manages things" - be concrete

**Good Examples:**

- `Generate AI-powered commit messages from staged changes`
- `Display all module contracts in a human-readable table`
- `Validate Gherkin specifications against quality contracts`

**Bad Examples:**

- ❌ `Handles commits` (too vague)
- ❌ `A command for creating specifications` (unnecessary words)
- ❌ `This command validates files` (don't say "this command")

#### `Long:` (Required, Multiple Lines)

Detailed multi-paragraph description explaining the command thoroughly.

Each `// Long:` line becomes a paragraph in the help output.

**Structure:**

1. **First paragraph**: What does this command do? What problem does it solve?
2. **Second paragraph**: How does it work? What's the behavior?
3. **Third paragraph**: What inputs does it accept? What outputs does it produce?
4. **Additional paragraphs** (optional): Default behaviors, validation requirements, special notes

**Best Practices:**

- Write in present tense
- Be specific about formats, locations, and behaviors
- Mention default behaviors explicitly
- Explain validation or error conditions
- Reference related commands if relevant

**Example:**

```go
// Long: The create spec command uses AI to transform natural language feature descriptions into
// Long: properly formatted Gherkin specifications following Rule/Scenario patterns. The generated specifications
// Long: include Feature, Rule, and Scenario blocks with appropriate tags and structure.
// Long: All specifications are validated against the specification contract to ensure they meet quality standards.
// Long: The command automatically saves the specification to the specs/ directory, organized by module.
// Long: Use --debug to inspect intermediate outputs and understand how the AI generates specifications.
```

#### `Flag.<name>:` (Optional, One per Flag)

Structured flag definition with key=value attributes.

**Format:**

```go
// Flag.<flagname>: type=<type>, default=<value>, shorthand=<letter>, usage=<description>, required=<true|false>, completion=<values>
```

**Attributes:**

- `type` (required): `bool`, `string`, `int`, `float`
- `default` (optional): Default value as string (e.g., `false`, `text`, `0`)
- `shorthand` (optional): Single letter for short flag (e.g., `d` for `-d`)
- `usage` (required): Educational description (can contain commas)
- `required` (optional): `true` or `false` (default: `false`)
- `completion` (optional): Comma-separated list of valid values

**Usage Field Guidelines:**
The usage field is the most important part - make it educational and specific:

- Explain WHAT the flag does
- Explain WHEN to use it
- Explain WHY it's useful
- Mention any side effects or important details
- Provide examples of valid values if applicable

**Good Flag Examples:**

```go
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs (context, prompts, AI responses) to the 'out' directory for troubleshooting and analysis
// Flag.module: type=string, shorthand=m, usage=Target module for the specification (e.g., eac-commands, eac-core). If not provided, the module will be inferred from the description
// Flag.format: type=string, shorthand=f, default=text, completion=text,json, usage=Output format for validation results (text for human-readable, json for machine-readable)
// Flag.quiet: type=bool, shorthand=q, default=false, usage=Suppress success messages and show only validation errors and warnings
```

**Bad Flag Examples:**

```go
// ❌ Flag.debug: type=bool, usage=Debug mode
// (Too vague - what does debug mode do?)

// ❌ Flag.output: type=string, usage=Output path
// (Missing details - what goes in this path? What format? What's the default?)

// ❌ Flag.verbose: type=bool, usage=More output
// (Not educational - more output of what? When would I use this?)
```

## Complete File Template

Here's a complete template with placeholder explanations:

```go
// Command: <command-name> or <parent> <subcommand>
// Short: <One sentence describing what this command does>
// Long: <Purpose: What does this do and what problem does it solve?>
// Long: <Behavior: How does it work? What's the workflow?>
// Long: <I/O: What inputs? What outputs? What format?>
// Long: <Details: Default behaviors, validation, special notes>
// Flag.flagname1: type=bool, shorthand=f, default=false, usage=<Educational description: what, when, why>
// Flag.flagname2: type=string, shorthand=s, usage=<Educational description with details and examples>, required=true, completion=option1,option2
package mypackage

import (
    "fmt"
    "os"
    "strings"

    "github.com/ready-to-release/eac/go/cli/eac/internal/registry"
    "github.com/ready-to-release/eac/go/core/repository"
)

func init() {
    registry.Register(MyCommand)
}

// Config holds configuration for this command
type Config struct {
    // Add fields for each flag
    FlagName1 bool
    FlagName2 string
    // Add other configuration fields
}

// MyCommand is the main entry point for the command
func MyCommand() int {
    // Parse configuration
    config, err := parseConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return 1
    }

    // Get repository root if needed
    workspaceRoot, err := repository.GetRepositoryRoot("")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
        return 1
    }

    // Implement command logic here
    // ...

    return 0
}

// parseConfig parses command line arguments into configuration
func parseConfig() (*Config, error) {
    config := &Config{}

    args := os.Args[2:] // Skip program name and command name
    for i := 0; i < len(args); i++ {
        arg := args[i]

        switch arg {
        case "--flagname1", "-f":
            config.FlagName1 = true

        case "--flagname2", "-s":
            if i+1 < len(args) {
                config.FlagName2 = args[i+1]
                i++
            } else {
                return nil, fmt.Errorf("--flagname2 requires a value")
            }

        default:
            if !strings.HasPrefix(arg, "-") {
                // Handle positional arguments
            } else {
                return nil, fmt.Errorf("unknown flag: %s", arg)
            }
        }
    }

    // Validation
    if config.FlagName2 == "" {
        return nil, fmt.Errorf("--flagname2 is required\n\nUsage: mycommand [--flagname1] --flagname2 <value>")
    }

    return config, nil
}
```

## Best Practices

### Writing Clear Help Text

1. **Be Specific**: Instead of "Process files", write "Validate all .feature files in the specs/ directory"
2. **Explain Defaults**: "By default, output is in human-readable text format"
3. **Mention Validation**: "The command validates that the module exists before proceeding"
4. **Reference Examples**: "e.g., eac-commands, eac-core" when describing modules
5. **Explain Exit Codes**: "Exit code is 0 if all validations pass, 1 if any critical errors are found"

### Flag Design

1. **Provide Shorthands**: Common flags (`-d` for debug, `-v` for verbose, `-q` for quiet)
2. **Use Completion**: Provide static completion values when options are limited
3. **Required Flags**: Only use `required=true` when the flag is absolutely necessary
4. **Default Values**: Always specify defaults for optional flags
5. **Consistent Naming**: Use kebab-case (`--output-path`, not `--outputPath`)

### Command Organization

1. **Group Related Commands**: Use parent commands (specs, show, get, etc.)
2. **Clear Naming**: Command names should be verb-noun (create, validate) or verb (show)
3. **Consistent Structure**: All commands follow the same pattern
4. **Error Handling**: Always return meaningful error messages with context

## Modern Flag Validation (Recommended)

Instead of manual switch-based flag parsing, use the centralized flag validation infrastructure:

```go
import (
    "github.com/ready-to-release/eac/go/cli/eac/internal/flags"
)

// Define your command's flags
var myCommandFlags = []flags.FlagDefinition{
    {
        Name:      "--module",
        Shorthand: "-m",
        HasValue:  true,
        ValueType: "string",
        Aliases:   []string{"--mod"}, // Common typos to catch
    },
    {
        Name:      "--debug",
        Shorthand: "-d",
        HasValue:  false,
        ValueType: "bool",
    },
}

func MyCommand() int {
    // Validate flags BEFORE parsing
    if err := flags.ValidateFlags(os.Args[2:], myCommandFlags); err != nil {
        log.Errorf("%v", err)
        return 1
    }

    // Now do your normal flag parsing
    // Invalid flags are already caught with helpful error messages
    for i := 2; i < len(os.Args); i++ {
        // ... your parsing logic
    }
}
```

**Benefits:**

- ✅ Catches typos (e.g., `--mod` suggests `--module`)
- ✅ Shows all valid flags when unknown flag detected
- ✅ Consistent error messages across commands
- ✅ No code duplication

## Debug Content Logging

When implementing `--debug` functionality, log debug content via the logger instead of writing files:

```go
import (
    "github.com/ready-to-release/eac/go/core/logging"
)

// Initialize logger
logger, err := logging.NewWithDebug("my-command", workspaceRoot)
if err != nil {
    return 1
}
defer logger.Sync()

// Later, instead of writing debug files:
// ❌ OLD: os.WriteFile("out/logs/my-command/debug-prompt.md", content, 0644)

// ✅ NEW: Use logger
if logger.IsDebugMode() {
    logger.LogDebugContent("prompt", promptText)
    log.Info("Debug content logged to commands.log")
}
```

**Benefits:**

- ✅ All debug output in unified `commands.log`
- ✅ No scattered debug files in `out/logs/`
- ✅ Can filter by content type in log viewer
- ✅ Respects logging configuration

## Testing Your Help Text

After creating your command, test the help output:

```bash
# View command list
go run . help

# View specific command help
go run . help your-command

# Test flag parsing
go run . your-command --help

# Generate completion (verify flags appear)
go run . completion bash
```

## Examples to Reference

See these well-documented commands for examples:

- `go/cli/eac/impl/commit/message.go` - Complex command with AI integration
- `go/cli/eac/impl/create/spec/create.go` - Multiple flags with completion
- `go/cli/eac/impl/validate/specs.go` - Format flag with completion
- `go/cli/eac/impl/show/modules.go` - Simple read-only command
- `go/cli/eac/impl/help/help.go` - Command with verbose flag
