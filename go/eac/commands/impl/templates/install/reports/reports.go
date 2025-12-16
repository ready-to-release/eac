// Command: templates install reports
// Short: Install report templates without value replacements
// Long: Install report templates by copying files as-is (no variable substitution).
// Long: Templates preserve {{ .Variable }} placeholders for later customization.
// Long:
// Long: Template Source and Destination:
// Long:   Source: templates/reports/ (fixed)
// Long:   Destination: .r2r/templates/reports/ (fixed)
// Long:
// Long: Use Case:
// Long:   Install templates once to your project, then customize them as needed.
// Long:   This command copies files without replacing placeholders.
// Long:
// Long: Examples:
// Long:   templates install reports
// Long:   templates install reports --debug
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save detailed logs to out/logs/templates/install/
package reports

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(TemplatesInstallReports)
}

// Config holds configuration for the reports install command
type Config struct {
	Destination   string
	WorkspaceRoot string
	Debug         bool
	Logger        *logging.Logger
}

// TemplatesInstallReports installs report templates
func TemplatesInstallReports() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}
	defer config.Logger.Sync()

	config.Logger.Info("Starting templates install reports command",
		zap.String("destination", config.Destination),
		zap.Bool("debug", config.Debug))

	// Resolve template directory
	templateDir, cleanup, err := resolveTemplateDirectory(config)
	if err != nil {
		config.Logger.Error("Failed to resolve template directory", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}
	defer cleanup()

	// Install templates (copy without value replacement)
	if err := installTemplates(config, templateDir); err != nil {
		config.Logger.Error("Failed to install templates", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}

	config.Logger.Info("Report templates installed successfully",
		zap.String("destination", config.Destination))
	log.Infof("✓ Report templates installed successfully to %s", config.Destination)

	return 0
}

// resolveTemplateDirectory determines the template directory (always local)
func resolveTemplateDirectory(config *Config) (string, func(), error) {
	// Always use local templates from appropriate root
	var root string
	if containerRoot := repository.GetContainerRoot(); containerRoot != "" {
		// Running in container - use container root
		root = containerRoot
		config.Logger.Info("Running in container, using local templates",
			zap.String("containerRoot", containerRoot))
	} else {
		// Not in container - use workspace root
		root = config.WorkspaceRoot
		config.Logger.Info("Using local templates from repository",
			zap.String("workspaceRoot", root))
	}

	templateDir := paths.TemplateReportsPath(root)

	// Verify directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("template directory does not exist: %s", templateDir)
	}

	config.Logger.Debug("Local templates validated",
		zap.String("dir", templateDir))
	log.Infof("Using templates from %s", templateDir)

	return templateDir, func() {}, nil
}

// installTemplates copies templates to destination
func installTemplates(config *Config, templateDir string) error {
	config.Logger.Info("Installing templates",
		zap.String("source", templateDir),
		zap.String("destination", config.Destination))
	log.Infof("Installing templates to %s...", config.Destination)

	// Create renderer with no values (will just copy files)
	renderer := internal.NewRenderer(templateDir, config.Destination, nil)
	if err := renderer.RenderTemplates(); err != nil {
		return fmt.Errorf("failed to install templates: %w", err)
	}

	config.Logger.Debug("Templates installed successfully")

	// Save debug info if enabled
	if config.Debug {
		writeDebugFile(config, "install-summary.txt", fmt.Sprintf(
			"Template source: %s\nDestination: %s\nMode: copy (no value replacement)\nSuccess: true\n",
			templateDir, config.Destination))
	}

	return nil
}

// writeDebugFile writes debug content to file
func writeDebugFile(c *Config, filename string, content string) {
	if !c.Debug {
		return
	}

	// Use helper function for clean fallback to defaults in test environments
	debugDir := filepath.Join(config.GetLogsPath(c.WorkspaceRoot, "templates"), "install")

	if err := os.MkdirAll(debugDir, 0755); err != nil {
		c.Logger.Warn("Failed to create debug directory", zap.Error(err))
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		c.Logger.Warn("Failed to write debug file",
			zap.String("file", debugFile),
			zap.Error(err))
	} else {
		c.Logger.Debug("Saved debug file", zap.String("file", debugFile))
	}
}

// parseConfig parses command-line arguments
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

	// Initialize logger
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("templates", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("templates", workspaceRoot)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Use fixed destination path
	destination := filepath.Join(workspaceRoot, paths.R2RDir, "templates", "reports")

	config := &Config{
		Destination:   destination,
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
		Logger:        logger,
	}

	return config, nil
}
