// Package commandparser provides parsing functionality for clie CLI commands
// It parses command-line arguments according to the EBNF schema and identifies
// the boundary between Viper-processed arguments and container arguments.
package commandparser

import (
	"fmt"
	"regexp"
	"strings"

	clie "github.com/ready-to-release/eac/contracts/clie/0.1.0"
)

// embeddedEBNFSchema is loaded from the contracts module at init time.
var embeddedEBNFSchema string

// Grammar elements extracted from the EBNF schema at init time.
// These are populated by parseGrammarFromEBNF and used by NewParser
// so the parser stays in sync with the contract schema automatically.
var (
	ebnfGlobalFlags []string
	ebnfSubcommands []string
)

func init() {
	data, err := clie.FS.ReadFile("schemas/command.ebnf")
	if err != nil {
		panic(fmt.Sprintf("failed to load embedded EBNF schema from contracts: %v", err))
	}
	embeddedEBNFSchema = string(data)
	ebnfGlobalFlags, ebnfSubcommands = parseGrammarFromEBNF(embeddedEBNFSchema)
}

// parseGrammarFromEBNF extracts global flags and subcommand keywords from the
// EBNF schema text. It uses simple regex extraction on the production rules
// rather than a full EBNF parser, which is sufficient for the flat alternation
// lists used in the schema.
func parseGrammarFromEBNF(schema string) (globalFlags, subcommands []string) {
	// Extract quoted string literals from a named production rule.
	// Pattern: production_name = ... "literal" | "literal" ... ;
	extractLiterals := func(productionName string) []string {
		// Match the full production: name = ... ;
		prodPattern := regexp.MustCompile(`(?m)` + regexp.QuoteMeta(productionName) + `\s*=\s*([^;]+);`)
		match := prodPattern.FindStringSubmatch(schema)
		if len(match) < 2 {
			return nil
		}
		body := match[1]

		// Extract all double-quoted literals from the production body
		litPattern := regexp.MustCompile(`"([^"]+)"`)
		matches := litPattern.FindAllStringSubmatch(body, -1)
		var literals []string
		for _, m := range matches {
			literals = append(literals, m[1])
		}
		return literals
	}

	globalFlags = extractLiterals("global_flag")
	subcommands = extractLiterals("subcommand")
	return globalFlags, subcommands
}

// ParsedCommand represents a parsed command structure.
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

// Parser handles command-line parsing according to EBNF schema.
type Parser struct {
	// Grammar elements extracted from the embedded EBNF schema at init time.
	// See parseGrammarFromEBNF for the extraction logic.
	validBinaryNames  map[string]bool
	validGlobalFlags  map[string]bool
	validSubcommands  map[string]bool
	requiresExtension map[string]bool
}

// NewParser creates a new command parser.
// Global flags and subcommands are populated from the embedded EBNF schema.
// Binary names and extension requirements are supplementary rules that are
// not expressed in the EBNF grammar and therefore remain explicitly listed.
func NewParser() *Parser {
	p := &Parser{
		// Binary names include platform variants not expressed in the EBNF grammar
		validBinaryNames: map[string]bool{
			"clie":                   true,
			"clie.exe":               true,
			"clie-windows-amd64.exe": true,
			"clie-linux-amd64":       true,
			"clie-darwin-amd64":      true,
		},

		// Populated from EBNF global_flag production
		validGlobalFlags: make(map[string]bool),

		// Populated from EBNF subcommand production
		validSubcommands: make(map[string]bool),

		// Commands requiring extension name (structural rule not in EBNF alternation)
		requiresExtension: map[string]bool{
			"run":      true, // RunCommand = "run" ExtensionName
			"metadata": true, // MetadataCommand = "metadata" ExtensionName
		},
	}

	// Populate from EBNF-extracted grammar elements.
	// Exclude --help/-h from global flags: these are Cobra-managed and should not
	// affect argument boundary detection (e.g. "clie run ext --help" should pass
	// --help to the container, not treat it as a CLIE flag).
	cobraManaged := map[string]bool{"--help": true, "-h": true}
	for _, flag := range ebnfGlobalFlags {
		if !cobraManaged[flag] {
			p.validGlobalFlags[flag] = true
		}
	}
	for _, cmd := range ebnfSubcommands {
		p.validSubcommands[cmd] = true
	}

	// Supplementary subcommands used by the system but represented as
	// extension_alias in the EBNF grammar (i.e. the EBNF allows any
	// identifier via the extension_alias production).
	for _, cmd := range []string{"verify", "definitions", "metadata", "interactive", "help"} {
		p.validSubcommands[cmd] = true
	}

	return p
}

// Parse parses command-line arguments into structured components.
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
		// Accept any non-flag as subcommand (valid or not - validation happens later)
		cmd.Subcommand = args[pos]
		cmd.ViperArgs = append(cmd.ViperArgs, args[pos])
		pos++
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
		// For run command, separate clie flags from container args
		boundary := pos
		containerArgsStarted := false

		for i := pos; i < len(args); i++ {
			arg := args[i]

			// Once we've seen the first container argument, all subsequent arguments
			// (even CLIE flags) should be treated as container arguments
			if containerArgsStarted {
				cmd.ContainerArgs = append(cmd.ContainerArgs, arg)
			} else {
				// Check if this is an clie flag that should be processed by Viper
				if p.IsCLIEFlag(arg) {
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

// ParseArgumentBoundary finds where Viper args end and container args begin.
func (p *Parser) ParseArgumentBoundary(args []string) int {
	cmd := p.Parse(args)
	return cmd.ArgumentBoundary
}

// SplitArguments separates Viper and container arguments.
func (p *Parser) SplitArguments(args []string) (viperArgs, containerArgs []string) {
	cmd := p.Parse(args)
	return cmd.ViperArgs, cmd.ContainerArgs
}

// IsGlobalFlag checks if a string is a valid global flag.
func (p *Parser) IsGlobalFlag(arg string) bool {
	return p.validGlobalFlags[arg]
}

// IsValidBinaryName checks if a string is a valid binary name.
func (p *Parser) IsValidBinaryName(name string) bool {
	return p.validBinaryNames[name]
}

// IsValidSubcommand checks if a string is a valid subcommand.
func (p *Parser) IsValidSubcommand(cmd string) bool {
	return p.validSubcommands[cmd]
}

// IsCLIEFlag checks if an argument is an clie flag that should be processed by Viper.
func (p *Parser) IsCLIEFlag(arg string) bool {
	// Check global flags
	if p.IsGlobalFlag(arg) {
		return true
	}

	return false
}

// FlagTakesValue checks if a flag expects a value argument.
func (p *Parser) FlagTakesValue(_ string) bool {
	// Currently, none of clie's flags take values (they're all boolean)
	// This method is here for future extensibility
	return false
}

// RequiresExtension checks if a subcommand requires an extension name.
func (p *Parser) RequiresExtension(subcommand string) bool {
	return p.requiresExtension[subcommand]
}

// IsValidExtensionName validates extension name format against the EBNF
// Identifier production: identifier = letter, { letter | digit | "-" | "_" } ;
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

// HasConflictingGlobalFlags checks for mutually exclusive global flags.
func (p *Parser) HasConflictingGlobalFlags(flags []string) bool {
	hasDebug := false
	hasQuiet := false

	for _, flag := range flags {
		if flag == "--clie-debug" {
			hasDebug = true
		}
		if flag == "--clie-quiet" {
			hasQuiet = true
		}
	}

	return hasDebug && hasQuiet
}

// GetEmbeddedSchema returns the embedded EBNF schema.
func GetEmbeddedSchema() string {
	return embeddedEBNFSchema
}
