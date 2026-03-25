package core

import (
	"context"
	"io"
)

// CommandPort defines metadata for a CLI command.
// Parent/group commands implement only this.
type CommandPort interface {
	Name() string
	Metadata() CommandMetadata
}

// SimpleCommandPort is an executable command with context and request.
type SimpleCommandPort interface {
	CommandPort
	Execute(ctx context.Context, req *CommandRequest) int
}

// CommandRequest provides execution context for a command invocation.
type CommandRequest struct {
	Args          []string
	WorkspaceRoot string
	Stdout        io.Writer
	Stderr        io.Writer
	Services      SimpleServicesPort
}

// CommandMetadata holds structured metadata for command registration and help.
type CommandMetadata struct {
	CanonicalName    string
	Short            string
	Long             string // Description text only. No embedded sub-sections.
	Notes            string // Expected output, caveats, notes (optional). Plain markdown.
	Flags            []FlagSpec
	Args             string
	IsParent         bool
	SubcommandGroups []SubcommandGroup
	Examples         []string
	Aliases          []string // Additional lookup names, e.g. ["work list"]
}

// FlagSpec defines a command flag.
type FlagSpec struct {
	Name         string
	Shorthand    string
	Type         string // "bool", "string", "int"
	DefaultValue string
	Usage        string
	Required     bool
}

// SubcommandGroup represents a logical grouping of subcommands for help display.
type SubcommandGroup struct {
	Name        string
	Subcommands []string
}

// SubcommandEntry pairs a registry key with the command it resolves to.
// The Key is the exact string used to look up the command (may be an alias),
// which is the correct label for the subcommand within its parent context.
type SubcommandEntry struct {
	Key string      // e.g. "work list" — the alias key under this parent
	Cmd CommandPort // the command (Cmd.Name() may be "show workspaces")
}

// CommandRegistryPort provides read access to registered commands.
type CommandRegistryPort interface {
	Get(name string) (CommandPort, bool)
	GetByCanonical(canonicalName string) (CommandPort, bool)
	All() []CommandPort
	Names() []string
	Subcommands(parentName string) []CommandPort
	// SubcommandEntries returns (key, command) pairs for direct children of parentName.
	// Key is the map key used to look up the command, which may be an alias.
	// Use Key (not Cmd.Name()) for display labels when listing children of a parent.
	SubcommandEntries(parentName string) []SubcommandEntry
}
