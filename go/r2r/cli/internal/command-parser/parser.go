// Package commandparser provides parsing functionality for r2r CLI commands
// It parses command-line arguments according to the EBNF schema and identifies
// the boundary between Viper-processed arguments and container arguments.
package commandparser

import (
	_ "embed"
	"strings"
)

// Embed the EBNF command schema at compile time from contracts
// Note: Path is relative to this file's location
//
//go:embed command.ebnf
var embeddedEBNFSchema string

// ParsedCommand represents a parsed command structure
type ParsedCommand struct {
	// Core components from parsing
	BinaryName    string
	GlobalFlags   []string
	Subcommand    string
	ExtensionName string

	// Argument separation - the key distinction
	ViperArgs     []string // Arguments processed by CLI framework
	ContainerArgs []string // Arguments passed to container (run command only)

	// Parsing metadata
	ArgumentBoundary int // Index where container args start (-1 if none)
}

// Parser handles command-line parsing according to EBNF schema
type Parser struct {
	// Grammar elements from EBNF schema
	// TODO: These should be parsed from EBNF dynamically
	// Currently hardcoded for simplicity and performance

	validBinaryNames  map[string]bool
	validGlobalFlags  map[string]bool
	validSubcommands  map[string]bool
	requiresExtension map[string]bool
}

// NewParser creates a new command parser
func NewParser() *Parser {
	return &Parser{
		// From BinaryName production in schema.ebnf
		validBinaryNames: map[string]bool{
			"r2r":                   true,
			"r2r.exe":               true,
			"r2r-windows-amd64.exe": true,
			"r2r-linux-amd64":       true,
			"r2r-darwin-amd64":      true,
		},

		// From GlobalFlag production in schema.ebnf
		validGlobalFlags: map[string]bool{
			"--r2r-debug": true,
			"--r2r-quiet": true,
		},

		// From Subcommand production in schema.ebnf
		validSubcommands: map[string]bool{
			"run":         true,
			"version":     true,
			"init":        true,
			"install":     true,
			"verify":      true,
			"validate":    true,
			"definitions": true,
			"metadata":    true,
			"update":      true,
			"cleanup":     true,
			"list":        true,
			"interactive": true,
			"help":        true,
		},

		// Commands requiring extension name (from EBNF productions)
		requiresExtension: map[string]bool{
			"run":      true, // RunCommand = "run" ExtensionName
			"metadata": true, // MetadataCommand = "metadata" ExtensionName
		},
	}
}

// Parse parses command-line arguments into structured components
func (p *Parser) Parse(args []string) *ParsedCommand {
	cmd := &ParsedCommand{
		GlobalFlags:      []string{},
		ViperArgs:        []string{},
		ContainerArgs:    []string{},
		ArgumentBoundary: -1,
	}

	if len(args) == 0 {
		return cmd
	}

	pos := 0

	// 1. Parse BinaryName (required)
	cmd.BinaryName = args[pos]
	cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
	pos++

	// 2. Parse GlobalFlags (optional, multiple allowed)
	for pos < len(args) && p.IsGlobalFlag(args[pos]) {
		cmd.GlobalFlags = append(cmd.GlobalFlags, args[pos])
		cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
		pos++
	}

	// 3. Parse Subcommand (optional)
	if pos >= len(args) {
		return cmd
	}

	// Check if next arg looks like a subcommand (not a flag)
	if !strings.HasPrefix(args[pos], "-") {
		// If it's not a flag, it should be a subcommand
		if p.IsValidSubcommand(args[pos]) {
			cmd.Subcommand = args[pos]
			cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
			pos++
		} else {
			// Invalid subcommand - still add to ViperArgs for validation
			cmd.Subcommand = args[pos]
			cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
			pos++
		}
	}

	// 4. Parse ExtensionName if required
	if p.requiresExtension[cmd.Subcommand] && pos < len(args) {
		// Accept any non-flag token as extension name
		if !strings.HasPrefix(args[pos], "-") {
			cmd.ExtensionName = args[pos]
			cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
			pos++
		}
	}

	// 5. Handle remaining arguments
	if cmd.Subcommand == "run" && pos < len(args) {
		// For run command, separate r2r flags from container args
		boundary := pos
		containerArgsStarted := false

		for i := pos; i < len(args); i++ {
			arg := args[i]

			// Once we've seen the first container argument, all subsequent arguments
			// (even R2R flags) should be treated as container arguments
			if containerArgsStarted {
				cmd.ContainerArgs = append(cmd.ContainerArgs, arg)
			} else {
				// Check if this is an r2r flag that should be processed by Viper
				if p.IsR2RFlag(arg) {
					cmd.ViperArgs = append(cmd.ViperArgs, arg)
					// If this is a flag that takes a value, include the next arg too
					if p.FlagTakesValue(arg) && i+1 < len(args) {
						i++ // Skip the next argument (flag value)
						cmd.ViperArgs = append(cmd.ViperArgs, args[i])
					}
				} else {
					// This is the first container argument - start container args mode
					containerArgsStarted = true
					cmd.ContainerArgs = append(cmd.ContainerArgs, arg)
					boundary = i
				}
			}
		}

		// Set boundary to first container arg, or -1 if no container args
		if len(cmd.ContainerArgs) > 0 {
			cmd.ArgumentBoundary = boundary
		} else {
			cmd.ArgumentBoundary = -1
		}
	} else if pos < len(args) {
		// For other commands, remaining args are Viper-processed
		cmd.ViperArgs = append(cmd.ViperArgs, args[pos:]...)
	}

	return cmd
}

// ParseArgumentBoundary finds where Viper args end and container args begin
func (p *Parser) ParseArgumentBoundary(args []string) int {
	cmd := p.Parse(args)
	return cmd.ArgumentBoundary
}

// SplitArguments separates Viper and container arguments
func (p *Parser) SplitArguments(args []string) (viperArgs []string, containerArgs []string) {
	cmd := p.Parse(args)
	return cmd.ViperArgs, cmd.ContainerArgs
}

// IsGlobalFlag checks if a string is a valid global flag
func (p *Parser) IsGlobalFlag(arg string) bool {
	return p.validGlobalFlags[arg]
}

// IsValidBinaryName checks if a string is a valid binary name
func (p *Parser) IsValidBinaryName(name string) bool {
	return p.validBinaryNames[name]
}

// IsValidSubcommand checks if a string is a valid subcommand
func (p *Parser) IsValidSubcommand(cmd string) bool {
	return p.validSubcommands[cmd]
}

// IsR2RFlag checks if an argument is an r2r flag that should be processed by Viper
func (p *Parser) IsR2RFlag(arg string) bool {
	// Check global flags
	if p.IsGlobalFlag(arg) {
		return true
	}

	return false
}

// FlagTakesValue checks if a flag expects a value argument
func (p *Parser) FlagTakesValue(flag string) bool {
	// Currently, none of r2r's flags take values (they're all boolean)
	// This method is here for future extensibility
	return false
}

// RequiresExtension checks if a subcommand requires an extension name
func (p *Parser) RequiresExtension(subcommand string) bool {
	return p.requiresExtension[subcommand]
}

// IsValidExtensionName validates extension name format
// TODO: Implement according to Identifier production from EBNF
func (p *Parser) IsValidExtensionName(name string) bool {
	if name == "" {
		return false
	}

	// Must start with a letter
	firstChar := name[0]
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')) {
		return false
	}

	// Rest can be letters, digits, hyphens, or underscores
	for _, char := range name[1:] {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// HasConflictingGlobalFlags checks for mutually exclusive global flags
func (p *Parser) HasConflictingGlobalFlags(flags []string) bool {
	hasDebug := false
	hasQuiet := false

	for _, flag := range flags {
		if flag == "--r2r-debug" {
			hasDebug = true
		}
		if flag == "--r2r-quiet" {
			hasQuiet = true
		}
	}

	return hasDebug && hasQuiet
}

// GetEmbeddedSchema returns the embedded EBNF schema
func GetEmbeddedSchema() string {
	return embeddedEBNFSchema
}

// TODO: Future enhancements
// 1. Use golang.org/x/exp/ebnf to parse schema.ebnf dynamically
// 2. Build parsing rules from parsed EBNF grammar
// 3. Support subcommand-specific flag parsing
// 4. Validate token formats according to EBNF productions
// 5. Cache parsed EBNF grammar for performance
