# template command

<!-- book:cmd template command -->

## Purpose

This is a **template file** for creating new CLI commands. It provides a complete, working example that developers can copy and customize.

## Usage

1. Copy `go/eac/commands/internal/template/template.go` to your target location
2. Update the package name and file name
3. Replace all `REPLACE:` placeholders
4. Customize the flag definitions and implementation
5. Delete the instruction comments

## Template Structure

The template demonstrates:

- **Header comments** for automatic metadata extraction
- **Flag definitions** with all supported attributes
- **Config struct** pattern for organizing flags
- **Three-phase execution**: parse → validate → execute
- **Error handling** with user-friendly messages

## See Also

- [Creating New Commands](../../../how-to-guides/eac/creating-commands.md)
- [Command Conventions](../overview/naming-conventions.md)
