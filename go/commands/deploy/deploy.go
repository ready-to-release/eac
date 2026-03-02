package deploy

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

var log = logging.C()

type deployCommand struct{}

var _ core.SimpleCommandPort = (*deployCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&deployCommand{},
	}
}

func (c *deployCommand) Name() string { return "deploy" }

func (c *deployCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "deploy",
		Short:         "Deploy a module to a target environment",
		Long: `Deploy a module to a target environment.

Deploys infrastructure and applications using the configured deployer
for each deployable component in the module.

Arguments:
  module       Module moniker to deploy
  environment  Target environment moniker (must match environments.yml)

Example:
  deploy infra development            Deploy infra to development
  deploy infra production --dry-run   Preview production changes`,
		Args: "module environment",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", Usage: "Preview changes without applying (maps to az what-if)"},
			{Name: "skip-deps", Type: "bool", Usage: "Skip system dependency verification"},
			{Name: "debug", Type: "bool", Usage: "Enable debug logs to console"},
			{Name: "tui", Type: "bool", Usage: "Enable TUI console"},
			{Name: "no-tui", Type: "bool", Usage: "Disable TUI console"},
			{Name: "component", Type: "string", Usage: "Only deploy a specific component within the module"},
		},
	}
}

func (c *deployCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Deploy()
}

// Deploy is the entry point for the deploy command.
func Deploy() int {
	args := os.Args[2:] // Skip program name and "deploy"

	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printDeployUsage()
		return 0
	}

	env := environment.Detect()
	shared, deployFlags, module, envMoniker, err := parseDeployArgs(args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printDeployUsage()
		return 1
	}

	cmdCfg := &cmdframework.CommandConfig{
		Type:        core.ActionDeploy,
		CommandPath: "deploy",
		OutputDir:   paths.OutDeployRelPath,
		Monikers:    []string{module},
		SkipDeps:    shared.SkipDeps,
		DryRun:      shared.DryRun,
		UseTUI:      shared.UseTUI,
		TUIHeight:   shared.TUIHeight,
		DebugMode:   shared.Debug,
	}

	deployCfg := &DeployConfig{
		Environment: envMoniker,
		Component:   deployFlags.Component,
	}
	cmdCfg.DeployCmdConfig = deployCfg

	return RunDeployWithFramework(cmdCfg, deployCfg)
}

// parseDeployArgs parses shared flags and deploy-specific positional args.
// Returns shared flags, deploy flags, module moniker, environment moniker, and error.
func parseDeployArgs(args []string, env *environment.Env) (*flags.SharedFlags, *DeploySpecificFlags, string, string, error) {
	shared, err := flags.ParseSharedFlagsWithEnv(flags.DeployConfig(), args, env)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Rebuild unconsumed args in original order so value-taking flags like
	// --component stay paired with their values.
	deployArgs := flags.RebuildUnconsumedArgs(args, shared.Remaining, shared.Monikers)

	// Parse deploy-specific flags from the reconstructed args
	deployFlags, positional, err := parseDeploySpecificFlags(deployArgs)
	if err != nil {
		return nil, nil, "", "", err
	}

	if len(positional) != 2 {
		return nil, nil, "", "", fmt.Errorf("deploy requires exactly 2 arguments: <module> <environment>, got %d", len(positional))
	}

	module := positional[0]
	envMoniker := positional[1]

	// Apply --dry-run from deploy-specific flags
	if deployFlags.DryRun {
		shared.DryRun = true
	}

	return shared, deployFlags, module, envMoniker, nil
}

// DeploySpecificFlags holds deploy-specific command flags.
type DeploySpecificFlags struct {
	Component string // --component flag
	DryRun    bool   // --dry-run flag (also parsed by shared)
}

// parseDeploySpecificFlags parses deploy-specific flags from remaining args.
func parseDeploySpecificFlags(args []string) (*DeploySpecificFlags, []string, error) {
	f := &DeploySpecificFlags{}
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--component" || arg == "-c":
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("--component requires a value")
			}
			i++
			f.Component = args[i]
		case strings.HasPrefix(arg, "--component="):
			f.Component = strings.TrimPrefix(arg, "--component=")
		case arg == "--dry-run":
			f.DryRun = true
		case strings.HasPrefix(arg, "--"):
			return nil, nil, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}

	return f, positional, nil
}

func printDeployUsage() {
	fmt.Println(`Usage: eac deploy <module> <environment> [flags]

Deploy a module to a target environment.

Arguments:
  module       Module moniker to deploy
  environment  Target environment moniker (must match environments.yml)

Flags:
  --dry-run            Preview changes without applying (maps to az what-if)
  --component <name>   Only deploy a specific component within the module
  --skip-deps          Skip system dependency verification
  --debug              Enable debug logs to console
  --tui                Enable TUI console
  --no-tui             Disable TUI console

Examples:
  eac deploy infra development            Deploy infra to development
  eac deploy infra production --dry-run   Preview production changes
  eac deploy infra development --component networking`)
}
