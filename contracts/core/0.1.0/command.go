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

// CommandRegistryPort provides read access to registered commands.
type CommandRegistryPort interface {
	Get(name string) (CommandPort, bool)
	GetByCanonical(canonicalName string) (CommandPort, bool)
	All() []CommandPort
	Names() []string
	Subcommands(parentName string) []CommandPort
}
