package release

import (
	"os"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

// releaseScaffold holds the common infrastructure used by most release commands.
// It eliminates per-file boilerplate for loading workspace root, module registry,
// git repository, and config.
type releaseScaffold struct {
	WorkspaceRoot  string
	ModuleRegistry *modules.Registry
	Repo           git.GitRepository
	Config         *config.RepositoryConfig
}

// scaffoldOption controls which optional components are loaded.
type scaffoldOption func(*scaffoldRequest)

type scaffoldRequest struct {
	needModules bool
	needGit     bool
	needConfig  bool
}

// withModules requests loading the module registry.
func withModules() scaffoldOption {
	return func(r *scaffoldRequest) {
		r.needModules = true
	}
}

// withGit requests opening the git repository.
func withGit() scaffoldOption {
	return func(r *scaffoldRequest) {
		r.needGit = true
	}
}

// withConfig requests loading the repository config.
func withConfig() scaffoldOption {
	return func(r *scaffoldRequest) {
		r.needConfig = true
	}
}

// newReleaseScaffold validates flags, loads the workspace root, and optionally
// loads the module registry, git repository, and config based on the provided options.
// Returns nil and an exit code if any step fails (error is already logged).
func newReleaseScaffold(opts ...scaffoldOption) (*releaseScaffold, int) {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return nil, 1
	}

	return buildScaffold(opts...)
}

// newReleaseScaffoldNoFlags is like newReleaseScaffold but skips flag validation.
// Use this for commands that do their own flag validation or don't use the flag registry.
func newReleaseScaffoldNoFlags(opts ...scaffoldOption) (*releaseScaffold, int) {
	return buildScaffold(opts...)
}

// buildScaffold loads the workspace root and optional components.
func buildScaffold(opts ...scaffoldOption) (*releaseScaffold, int) {
	req := &scaffoldRequest{}
	for _, opt := range opts {
		opt(req)
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to get workspace root: %v", err)
		return nil, 1
	}

	s := &releaseScaffold{
		WorkspaceRoot: workspaceRoot,
	}

	// Load module registry if requested
	if req.needModules {
		reg, err := modules.LoadFromWorkspace("")
		if err != nil {
			log.Errorf("failed to load modules: %v", err)
			return nil, 1
		}
		s.ModuleRegistry = reg
	}

	// Open git repository if requested
	if req.needGit {
		gitMgr := git.NewManager(logging.C().Zap())
		repo, err := gitMgr.Open("")
		if err != nil {
			log.Errorf("failed to open git repository: %v", err)
			return nil, 1
		}
		s.Repo = repo
	}

	// Load config if requested
	if req.needConfig {
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			log.Errorf("failed to load config: %v", err)
			return nil, 1
		}
		s.Config = cfg.Repository
	}

	return s, 0
}
