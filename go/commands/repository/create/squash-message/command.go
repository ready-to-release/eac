package squashmessage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type createSquashMessageCommand struct{}

var _ core.SimpleCommandPort = (*createSquashMessageCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&createSquashMessageCommand{},
	}
}

func (c *createSquashMessageCommand) Name() string { return "create squash-message" }

func (c *createSquashMessageCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "create-squash-message",
		Short:         "Generate squash commit message from branch commits",
		Long:          "Analyzes all commits in the current branch compared to the base branch\nand generates a comprehensive, cohesive commit message suitable for\nsquash merges in pull requests or local branch merging.\n\nThis command examines the full commit history and cumulative diff\nto create a message that accurately represents the entire feature\nor change set, rather than individual commit details.\n\nThe generated message is designed to be copied into GitHub's PR squash\nmerge UI for use when squashing and merging pull requests.\n\nExpected Output:\n- Comprehensive commit message for squash merge\n- Suitable for GitHub PR squash merge UI\n\nExample:\n  create squash-message\n  create squash-message --base=develop\n  create squash-message --debug",
		Flags: []core.FlagSpec{
			{Name: "base", Type: "string", DefaultValue: "main", Usage: "Base branch for comparison"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Enable debug mode to save intermediate outputs"},
		},
	}
}

func (c *createSquashMessageCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return CreateSquashMessage()
}

var log = logging.C()

// CreateSquashMessage is the main entry point for the create squash-message command.
func CreateSquashMessage() int {
	return createSquashMessage(defaultDeps())
}

func createSquashMessage(deps *Deps) int {
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

	log.Debug("Starting create squash-message command")

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
	log.Infof("Analyzing commits from base branch to HEAD: baseBranch=%s", config.baseBranch)
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

	// Phase 10: Generate top-level message using AI
	log.Info("Generating squash commit message using AI...")
	topLevelMessage, err := generateTopLevelMessage(workspaceRoot, promptCtx)
	if err != nil {
		log.Errorf("Failed to generate top-level message: %v", err)
		return 1
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

	// Phase 13: Output message
	fmt.Println(finalMessage)

	log.Debug("Create squash-message command completed successfully")
	return 0
}

// squashConfig holds the parsed command configuration.
type squashConfig struct {
	baseBranch string
	debug      bool
}

// parseConfig parses command line arguments.
func parseConfig() (*squashConfig, error) {
	args := os.Args[3:] // Skip "clie", "create", "squash-message"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	cfg := &squashConfig{
		baseBranch: "main",
		debug:      flags.ParseDebugFlag(args),
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--base="):
			cfg.baseBranch = strings.TrimPrefix(arg, "--base=")
		case arg == "--base":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--base requires a value")
			}
			i++
			cfg.baseBranch = args[i]
		case arg == "--debug" || arg == "-d":
			// Already handled by shared flags package
		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return cfg, nil
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
