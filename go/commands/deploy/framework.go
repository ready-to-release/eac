package deploy

import (
	"fmt"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
)

func init() {
	// Register deploy as an action type with unit-level execution
	cmdframework.RegisterUnitProvider(core.ActionDeploy, ResolveUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionDeploy, deployUnitWorker)
}

// DeployConfig holds deploy-specific configuration.
type DeployConfig struct {
	Environment string // Target environment moniker
	Component   string // Optional component filter
}

// deployContext holds deploy-specific state during execution.
type deployContext struct {
	cfg     *DeployConfig
	tracker *coreoutput.InMemoryTracker
}

// RunDeployWithFramework runs the deploy command using the shared framework.
func RunDeployWithFramework(cmdCfg *cmdframework.CommandConfig, deployCfg *DeployConfig) int {
	dctx := &deployContext{
		cfg: deployCfg,
	}
	cmdCfg.DeployCmdContext = dctx

	hooks := &cmdframework.Hooks{
		AfterInit:    deployAfterInit,
		AfterResolve: deployAfterResolve,
	}

	return cmdframework.Run(cmdCfg, nil, hooks)
}

func deployAfterInit(ctx *cmdframework.ExecutionContext) error {
	deployCfg, ok := ctx.Config.DeployCmdConfig.(*DeployConfig)
	if !ok {
		return fmt.Errorf("deploy: missing or invalid DeployCmdConfig")
	}

	summary := initsummary.New("deploy").
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetOutputDir(paths.OutDeployRelPath)

	ctx.InitSummary = summary

	dctx := ctx.Config.DeployCmdContext.(*deployContext)
	dctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionDeploy)

	// Validate environment exists in configuration
	if ctx.EACConfig != nil && ctx.EACConfig.Environments != nil {
		_, exists := ctx.EACConfig.Environments.GetEnvironment(deployCfg.Environment)
		if !exists {
			return fmt.Errorf("unknown environment: %q (check environments.yml)", deployCfg.Environment)
		}
	}

	return nil
}

func deployAfterResolve(ctx *cmdframework.ExecutionContext) error {
	// Deploy does not do incremental detection - always deploys.
	// Validate that at least one module has deployable components.
	specs := ResolveUnitSpecs(ctx)
	if len(specs) == 0 {
		return fmt.Errorf("no deployable components found in module(s): %v", ctx.GetExecutionMonikers())
	}

	return nil
}

