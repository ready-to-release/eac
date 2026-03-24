package squashmessage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/commands/repository/create/aiutil"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getSquashMessageCommand struct{}

var _ core.SimpleCommandPort = (*getSquashMessageCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&getSquashMessageCommand{},
	}
}

func (c *getSquashMessageCommand) Name() string { return "get squash-message" }

func (c *getSquashMessageCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-squash-message",
		Short:         "Generate squash commit message from branch commits",
		Long: "Analyzes all commits in the current branch compared to the base branch\nand generates a comprehensive, cohesive commit message suitable for\nsquash merges in pull requests or local branch merging.\n\nThis command examines the full commit history and cumulative diff\nto create a message that accurately represents the entire feature\nor change set, rather than individual commit details.\n\nThe generated message is designed to be copied into GitHub's PR squash\nmerge UI for use when squashing and merging pull requests.",
		Notes: "Expected Output:\n- Comprehensive commit message for squash merge\n- Suitable for GitHub PR squash merge UI",
		Examples: []string{
			"eac get squash-message",
			"eac get squash-message --base=develop",
			"eac get squash-message --debug",
		},
		Flags: []core.FlagSpec{
			{Name: "base", Type: "string", DefaultValue: "", Usage: "Base branch for comparison (default: trunk_branch from config, or \"main\")"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Enable debug mode to save intermediate outputs"},
		},
	}
}

func (c *getSquashMessageCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetSquashMessage()
}

var log = logging.C()

// GetSquashMessage is the main entry point for the get squash-message command.
func GetSquashMessage() int {
	return getSquashMessage(defaultDeps())
}

func getSquashMessage(deps *Deps) int {
	// Phase 1: Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("ERROR: %v", err)
		return 1
	}

	// Phase 2: Get workspace root (needed for logger)
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("ERROR: Failed to get workspace root: %v", err)
		return 1
	}

	// Phase 3: Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, config.debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Debug("Starting get squash-message command")

	// Phase 4: Open git repository
	repo, err := deps.GetGitRepo(workspaceRoot)
	if err != nil {
		log.Errorf("Failed to open git repository: %v", err)
		return 1
	}

	// Phase 5: Get current branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		log.Errorf("Failed to get current branch: %v", err)
		return 1
	}
	log.Debugf("Current branch: branch=%s", currentBranch)

	// Phase 6: Get branch commits
	// Resolve effective base branch: explicit flag → upstream → trunk_branch from config → "main"
	config.baseBranch = resolveBaseBranch(config, workspaceRoot, repo, currentBranch)
	log.Infof("🔍 Analyzing commits from %s to HEAD...", config.baseBranch)
	commits, err := repo.GetBranchCommits(config.baseBranch)
	if err != nil {
		// Log the error and ensure "no commits ahead" message is in stdout
		log.Error(err.Error())
		return 1
	}
	log.Debugf("Found commits: count=%d", len(commits))

	// Phase 7: Get branch diff and files
	diff, err := repo.GetBranchDiff(config.baseBranch)
	if err != nil {
		log.Warnf("Failed to get branch diff: %v", err)
		diff = ""
	}

	diffStats, err := repo.GetBranchDiffStats(config.baseBranch)
	if err != nil {
		log.Warnf("Failed to get diff stats: %v", err)
		diffStats = ""
	}

	fileNames, err := repo.GetBranchFiles(config.baseBranch)
	if err != nil {
		log.Errorf("Failed to get branch files: %v", err)
		return 1
	}
	log.Debugf("Found changed files: count=%d", len(fileNames))

	// Phase 8: Get module mappings
	filesWithModules, err := enrichFilesWithModuleInfo(fileNames, workspaceRoot)
	if err != nil {
		log.Errorf("Failed to get module mappings: %v", err)
		return 1
	}

	affectedModules := extractAffectedModules(filesWithModules)
	log.Debugf("Affected modules: modules=%v", affectedModules)

	// Phase 9: Build prompt context
	promptCtx := buildSquashContext(currentBranch, config.baseBranch, commits, filesWithModules, diff, diffStats, affectedModules)
	logDebugArtifact("SQUASH-CONTEXT", promptCtx)

	// Phase 10: Generate top-level message with retry loop
	const maxRetries = 5
	var topLevelMessage string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 3 {
			log.Warnf("retry attempt: attempt=%d, max=%d", attempt, maxRetries)
		}

		msg, err := attemptSquashGeneration(workspaceRoot, promptCtx)
		if err == nil {
			topLevelMessage = msg
			break
		}
		log.Warnf("Generation attempt %d failed: %v", attempt, err)

		if attempt == maxRetries {
			log.Error("Maximum retry attempts reached")
			log.Info("The AI is having difficulty generating a valid squash message.")
			log.Info("Please try one of the following:")
			log.Info("  - Simplify your branch changes")
			log.Info("  - Reduce the number of commits")
			log.Info("  - Write the squash message manually")
			return 1
		}

		log.Infof("Retrying squash message generation: attempt=%d, max=%d", attempt+1, maxRetries)
	}

	// Phase 11: Generate module sections (reuse commit-message logic)
	moduleSections, err := generateModuleSections(workspaceRoot, affectedModules, filesWithModules, diff)
	if err != nil {
		log.Errorf("Failed to generate module sections: %v", err)
		return 1
	}

	// Phase 12: Assemble final message
	finalMessage := assembleMessage(topLevelMessage, moduleSections)
	logDebugArtifact("SQUASH-FINAL", finalMessage)

	// Phase 13: Validate and show warnings (non-blocking)
	validator := NewSquashMessageValidator()
	validationErrors := validator.Validate(finalMessage, nil)
	if len(validationErrors) > 0 {
		log.Warn("Generated squash message has validation issues:")
		for _, verr := range validationErrors {
			log.Warnf("  - %s", verr.Message)
		}
	}

	// Phase 14: Output with marker
	fmt.Println(">>>>>>OUTPUT START<<<<<<")
	fmt.Println(finalMessage)

	log.Debug("Get squash-message command completed successfully")
	return 0
}

// attemptSquashGeneration wraps a single generation attempt with progress display.
func attemptSquashGeneration(workspaceRoot, promptCtx string) (string, error) {
	var result string
	err := aiutil.WithProgress("🤖 Generating squash commit message...", func() error {
		msg, genErr := generateTopLevelMessage(workspaceRoot, promptCtx)
		if genErr != nil {
			return genErr
		}
		result = msg
		return nil
	})
	return result, err
}

// squashConfig holds the parsed command configuration.
type squashConfig struct {
	baseBranch    string
	baseExplicit  bool // true when --base was provided explicitly
	debug         bool
}

// parseConfig parses command line arguments.
func parseConfig() (*squashConfig, error) {
	args := os.Args[3:] // Skip "clie", "get", "squash-message"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	cfg := &squashConfig{
		debug: flags.ParseDebugFlag(args),
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--base="):
			cfg.baseBranch = strings.TrimPrefix(arg, "--base=")
			cfg.baseExplicit = true
		case arg == "--base":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--base requires a value")
			}
			i++
			cfg.baseBranch = args[i]
			cfg.baseExplicit = true
		case arg == "--debug" || arg == "-d":
			// Already handled by shared flags package
		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return cfg, nil
}

// resolveBaseBranch returns the effective base branch using a priority chain:
// 1. --base flag (explicit)
// 2. Current branch's upstream tracking branch (only if different from current branch)
// 3. trunk_branch from .eac/repository.yml
// 4. "main" (hard fallback)
//
// The upstream is skipped when it matches currentBranch because a typical push
// sets branch.<name>.merge to the same branch on origin — that is a remote tracking
// ref, not a base/parent branch, and comparing a branch against itself yields zero
// commits ahead.
func resolveBaseBranch(cfg *squashConfig, workspaceRoot string, repo git.GitBranchComparer, currentBranch string) string {
	if cfg.baseExplicit {
		return cfg.baseBranch
	}
	if upstream, err := repo.UpstreamBranch(); err == nil && upstream != "" && upstream != currentBranch {
		return upstream
	}
	if eacCfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot}); err == nil {
		if tb := eacCfg.Repository.Repository.TrunkBranch; tb != "" {
			return tb
		}
	}
	return "main"
}

// enrichFilesWithModuleInfo converts file names to FileInfo and enriches with module information.
func enrichFilesWithModuleInfo(fileNames []string, workspaceRoot string) ([]repository.RepositoryFileWithModule, error) {
	var fileInfos []repository.FileInfo
	for _, fileName := range fileNames {
		fileInfos = append(fileInfos, repository.FileInfo{
			Path:         fileName,
			AbsolutePath: filepath.Join(workspaceRoot, fileName),
			IsTracked:    true,
			IsIgnored:    false,
		})
	}
	return repository.EnrichFilesWithModules(fileInfos, workspaceRoot)
}
