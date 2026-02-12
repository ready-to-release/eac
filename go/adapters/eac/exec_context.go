package eac

import "github.com/ready-to-release/eac/go/core/tool"

// buildExecContext constructs a tool.ExecutionContext from shared adapter fields.
// It resolves the workspace root from cfg (falling back to the adapter default)
// and returns both the context and the resolved workspace path.
func buildExecContext(defaultWorkspace string, args []string, cfg *ExecConfig) (*tool.ExecutionContext, string) {
	if cfg == nil {
		cfg = &ExecConfig{}
	}
	ws := cfg.WorkspaceRoot
	if ws == "" {
		ws = defaultWorkspace
	}
	return &tool.ExecutionContext{
		WorkspaceRoot: ws,
		ModuleRoot:    cfg.ModuleRoot,
		OutputDir:     cfg.OutputDir,
		StdoutWriter:  cfg.StdoutWriter,
		StderrWriter:  cfg.StderrWriter,
		StdinReader:   cfg.StdinReader,
		EnvOverrides:  cfg.Env,
		FullEnv:       cfg.FullEnv,
		ArgsOverrides: args,
	}, ws
}
