package flags

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

// ParsedFlags holds all parsed flag values from subscribed sets.
type ParsedFlags struct {
	Execution *ExecutionFlags
	Output    *OutputFlags
	Cache     *CacheFlags
	Module    *ModuleFlags
	DryRun    *DryRunFlags

	// Positional arguments (module monikers)
	Positional []string

	// Remaining flags (for command-specific parsing)
	Remaining []string
}

// CommandFlagConfig declares which flag sets a command subscribes to.
type CommandFlagConfig struct {
	Command   string
	Execution bool
	Output    bool
	Cache     bool
	Module    bool
	DryRun    bool
}

// Parser handles multi-set flag parsing for a command.
type Parser struct {
	config CommandFlagConfig
	sets   []FlagSet
	env    *environment.Env
}

// NewParser creates a parser for the given command configuration.
func NewParser(config CommandFlagConfig) *Parser {
	return NewParserWithEnv(config, environment.Detect())
}

// NewParserWithEnv creates a parser with a specific environment (for testing).
func NewParserWithEnv(config CommandFlagConfig, env *environment.Env) *Parser {
	p := &Parser{
		config: config,
		env:    env,
	}

	if config.Execution {
		p.sets = append(p.sets, NewExecutionFlagSet())
	}
	if config.Output {
		p.sets = append(p.sets, NewOutputFlagSet())
	}
	if config.Cache {
		p.sets = append(p.sets, NewCacheFlagSet())
	}
	if config.Module {
		p.sets = append(p.sets, NewModuleFlagSet())
	}
	if config.DryRun {
		p.sets = append(p.sets, NewDryRunFlagSet())
	}

	return p
}

// Parse processes command-line arguments through all subscribed flag sets.
func (p *Parser) Parse(args []string) (*ParsedFlags, error) {
	result := &ParsedFlags{}
	remaining := args

	// Apply defaults to all sets first
	for _, set := range p.sets {
		set.ApplyDefaults(p.env)
	}

	// Parse through each subscribed set
	for _, set := range p.sets {
		var err error
		remaining, err = set.Parse(remaining, p.env)
		if err != nil {
			return nil, fmt.Errorf("parsing %s flags: %w", set.Name(), err)
		}
	}

	// Validate each set
	for _, set := range p.sets {
		if err := set.Validate(); err != nil {
			return nil, fmt.Errorf("validating %s flags: %w", set.Name(), err)
		}
	}

	// Validate TUI if output set is subscribed
	if p.config.Output {
		for _, set := range p.sets {
			if outputSet, ok := set.(*OutputFlagSet); ok {
				if err := outputSet.ValidateTUI(p.env); err != nil {
					return nil, err
				}
			}
		}
	}

	// Extract positional arguments and unknown flags
	for _, arg := range remaining {
		if strings.HasPrefix(arg, "-") {
			result.Remaining = append(result.Remaining, arg)
		} else {
			result.Positional = append(result.Positional, arg)
		}
	}

	// Copy values from sets to result
	for _, set := range p.sets {
		switch s := set.(type) {
		case *ExecutionFlagSet:
			result.Execution = s.Values()
		case *OutputFlagSet:
			result.Output = s.Values()
		case *CacheFlagSet:
			result.Cache = s.Values()
		case *ModuleFlagSet:
			result.Module = s.Values()
		case *DryRunFlagSet:
			result.DryRun = s.Values()
		}
	}

	return result, nil
}

// AllFlags returns all flag definitions for subscribed sets.
func (p *Parser) AllFlags() []FlagDef {
	var all []FlagDef
	for _, set := range p.sets {
		all = append(all, set.Flags()...)
	}
	return all
}

// GenerateUsage generates help text for subscribed flag sets.
func (p *Parser) GenerateUsage() string {
	var sb strings.Builder

	for _, set := range p.sets {
		sb.WriteString(fmt.Sprintf("\n%s:\n", set.Description()))
		for _, flag := range set.Flags() {
			flagStr := fmt.Sprintf("  --%s", flag.Name)
			if flag.Shorthand != "" {
				flagStr += fmt.Sprintf(", -%s", flag.Shorthand)
			}
			if flag.Type != "bool" {
				flagStr += fmt.Sprintf(" <%s>", flag.Type)
			}
			for len(flagStr) < 28 {
				flagStr += " "
			}
			flagStr += flag.Usage
			if flag.Default != "" && flag.Default != "false" {
				flagStr += fmt.Sprintf(" (default: %s)", flag.Default)
			}
			sb.WriteString(flagStr + "\n")
		}
	}

	return sb.String()
}
