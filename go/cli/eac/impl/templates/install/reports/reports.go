package reports

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/templates/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type templatesInstallReportsCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallReportsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallReportsCommand{},
	}
}

func (c *templatesInstallReportsCommand) Name() string { return "templates install reports" }

func (c *templatesInstallReportsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-reports",
		Short:         "Install report templates without value replacements",
		Long:          "Install report templates by copying files as-is (no variable substitution).\nTemplates preserve {{ .Variable }} placeholders for later customization.\n\nTemplate Source and Destination:\n  Source: templates/reports/ (fixed)\n  Destination: .clie/templates/reports/ (fixed)\n\nUse Case:\n  Install templates once to your project, then customize them as needed.\n  This command copies files without replacing placeholders.\n\nExamples:\n  templates install reports\n  templates install reports --debug",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallReportsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallReports()
}

var log = logging.C()
// Config holds configuration for the reports install command.
type Config struct {
	Destination   string
	WorkspaceRoot string
	Debug         bool
}

// TemplatesInstallReports installs report templates.
func TemplatesInstallReports() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse configuration
	cfg, err := parseConfig()
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Configure logging system (logs to out/commands.log)
	if err := logging.ConfigureLoggingSimple(cfg.WorkspaceRoot, "commands", nil, cfg.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Debugf("Starting templates install reports command: destination=%s, debug=%v", cfg.Destination, cfg.Debug)

	// Resolve template directory
	templateDir, cleanup, err := resolveTemplateDirectory(cfg)
	if err != nil {
		log.Debugf("Failed to resolve template directory: error=%v", err)
		log.Errorf("%v", err)
		return 1
	}
	defer cleanup()

	// Install templates (copy without value replacement)
	if err := installTemplates(cfg, templateDir); err != nil {
		log.Debugf("Failed to install templates: error=%v", err)
		log.Errorf("%v", err)
		return 1
	}

	log.Debugf("Report templates installed successfully: destination=%s", cfg.Destination)
	log.Infof("✓ Report templates installed successfully to %s", cfg.Destination)

	return 0
}

// resolveTemplateDirectory determines the template directory (always local).
func resolveTemplateDirectory(cfg *Config) (string, func(), error) {
	// Always use local templates from appropriate root
	var root string
	if containerRoot := repository.GetContainerRoot(); containerRoot != "" {
		// Running in container - use container root
		root = containerRoot
		log.Debugf("Running in container, using local templates: containerRoot=%s", containerRoot)
	} else {
		// Not in container - use workspace root
		root = cfg.WorkspaceRoot
		log.Debugf("Using local templates from repository: workspaceRoot=%s", root)
	}

	templateDir := paths.TemplateReportsPath(root)

	// Verify directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("template directory does not exist: %s", templateDir)
	}

	log.Debugf("Local templates validated: dir=%s", templateDir)
	log.Infof("Using templates from %s", templateDir)

	return templateDir, func() {}, nil
}

// installTemplates copies templates to destination.
func installTemplates(cfg *Config, templateDir string) error {
	log.Debugf("Installing templates: source=%s, destination=%s", templateDir, cfg.Destination)
	log.Infof("Installing templates to %s...", cfg.Destination)

	// Create renderer with no values (will just copy files)
	renderer := internal.NewRenderer(templateDir, cfg.Destination, nil)
	if err := renderer.RenderTemplates(); err != nil {
		return fmt.Errorf("failed to install templates: %w", err)
	}

	log.Debugf("Templates installed successfully")

	// Save debug info if enabled
	if cfg.Debug {
		writeDebugFile(cfg, "install-summary.txt", fmt.Sprintf(
			"Template source: %s\nDestination: %s\nMode: copy (no value replacement)\nSuccess: true\n",
			templateDir, cfg.Destination))
	}

	return nil
}

// writeDebugFile writes debug content to file.
func writeDebugFile(c *Config, filename, content string) {
	if !c.Debug {
		return
	}

	// Use helper function for clean fallback to defaults in test environments
	debugDir := filepath.Join(config.GetLogsPath(c.WorkspaceRoot, "templates"), "install")

	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		log.Debugf("Failed to create debug directory: error=%v", err)
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0o644); err != nil {
		log.Debugf("Failed to write debug file: file=%s, error=%v", debugFile, err)
	} else {
		log.Debugf("Saved debug file: file=%s", debugFile)
	}
}

// parseConfig parses command-line arguments.
func parseConfig() (*Config, error) {
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip "binary templates install reports"
	}

	// Parse flags manually (only --debug supported)
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		}
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Use fixed destination path
	destination := filepath.Join(workspaceRoot, paths.CLIEDir, "templates", "reports")

	cfg := &Config{
		Destination:   destination,
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
	}

	return cfg, nil
}
