package deploy

import (
	"context"
	"io"
	"os"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func deployUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int {
	deployCfg, ok := ctx.Config.DeployCmdConfig.(*DeployConfig)
	if !ok {
		output.Writeln(logWriter, "Error: DeployCmdConfig not found or wrong type")
		return 1
	}

	dctx, ok := ctx.Config.DeployCmdContext.(*deployContext)
	if !ok {
		output.Writeln(logWriter, "Error: DeployCmdContext not found or wrong type")
		return 1
	}

	module := spec.ID.Module
	compName := spec.ID.ComponentName
	toolName := spec.ID.Tool

	// Get module contract
	moduleContract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", module)
		return 1
	}

	// Resolve deployer handler
	handler := tool.GlobalDeployBridge().GetHandler(toolName)
	if handler == nil {
		output.Writeln(logWriter, "Error: no deploy handler found: %s", toolName)
		return 1
	}

	// Validate module for deployment
	if err := handler.ValidateModule(moduleContract, ctx.WorkspaceRoot, deployCfg.Environment); err != nil {
		output.Writeln(logWriter, "Validation failed: %v", err)
		return 1
	}

	// Prepare output directory: out/deploy/<module>/<environment>/
	outputDir := paths.DeployOutputPath(ctx.WorkspaceRoot, module, deployCfg.Environment)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		output.Writeln(logWriter, "Error creating output dir: %v", err)
		return 1
	}

	// Build deploy options
	opts := tool.DeployOptions{
		Environment: deployCfg.Environment,
		DryRun:      ctx.Config.DryRun,
		Component:   compName,
	}

	// Load environment deploy config if available
	if ctx.EACConfig != nil && ctx.EACConfig.Environments != nil {
		if env, envOK := ctx.EACConfig.Environments.GetEnvironment(deployCfg.Environment); envOK {
			opts.DeployConfig = resolveDeployEnvironmentConfig(ctx.EACConfig.Environments, env)
		}
	}

	// Record UoW start
	unitID := spec.ID
	if dctx.tracker != nil {
		_ = dctx.tracker.RecordStart(unitID)
	}

	// Execute deployment
	output.Writeln(logWriter, "=== Deploying %s to %s (handler: %s) ===", compName, deployCfg.Environment, handler.Name())

	var exitCode int
	if ctx.Config.DryRun {
		exitCode = handler.DryRun(goCtx, moduleContract, ctx.WorkspaceRoot, outputDir, logWriter, opts)
	} else {
		exitCode = handler.Deploy(goCtx, moduleContract, ctx.WorkspaceRoot, outputDir, logWriter, opts)
	}

	// Record manifest
	if dctx.tracker != nil {
		manifest := &coreoutput.UoWManifest{
			ExitCode:   exitCode,
			ExecutedAt: time.Now().UTC(),
			Version:    "1.0.0",
		}
		_ = dctx.tracker.RecordComplete(unitID, manifest)
	}

	if exitCode != 0 {
		output.Writeln(logWriter, "Deployment failed for %s (exit code: %d)", compName, exitCode)
	} else {
		output.Writeln(logWriter, "Deployment succeeded for %s", compName)
	}

	return exitCode
}

// resolveDeployEnvironmentConfig builds a DeployEnvironmentConfig from the environment.
func resolveDeployEnvironmentConfig(envsCfg *config.EnvironmentsConfig, env *config.Environment) *tool.DeployEnvironmentConfig {
	cfg := &tool.DeployEnvironmentConfig{
		SubscriptionID: envsCfg.ResolveSubscriptionID(env),
		TenantID:       envsCfg.ResolveTenantID(env),
		ResourceGroup:  env.GetDefinition("resource_group"),
		Location:       env.GetDefinition("region"),
	}

	// Copy all definition entries as env vars
	if len(env.Definition) > 0 {
		cfg.Env = make(map[string]string, len(env.Definition))
		for k, v := range env.Definition {
			cfg.Env[k] = v
		}
	}

	return cfg
}
