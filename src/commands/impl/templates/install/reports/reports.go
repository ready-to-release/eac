// Command: templates install reports
// Short: Install report templates without value replacements
// Long: Install report templates by copying files as-is (no variable substitution).
// Long: Templates preserve {{ .Variable }} placeholders for later customization.
// Long:
// Long: Template Source (--source):
// Long:   If provided: uses local directory containing template files
// Long:   If omitted: automatically clones from https://github.com/ready-to-release/eac
// Long:
// Long: Use Case:
// Long:   Install templates once to your project, then customize them as needed.
// Long:   Unlike 'apply', this command does NOT replace placeholders.
// Long:
// Long: Examples:
// Long:   templates install reports
// Long:   templates install reports --source ./my-templates
// Long:   templates install reports --destination ./.r2r/templates --debug
// Flag.source: type=string, usage=Local template directory (default: clones from GitHub)
// Flag.destination: type=string, usage=Output directory (default: .r2r/templates/reports)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save detailed logs to out/logs/templates/install/
package reports

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/src/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
)

const (
	defaultReportsRepoURL    = "https://github.com/ready-to-release/eac"
	defaultReportsSourcePath = "templates/reports"
	defaultReportsDest       = ".r2r/templates/reports"
)

func init() {
	registry.Register(TemplatesInstallReports)
}

// Config holds configuration for the reports install command
type Config struct {
	Source        string
	Destination   string
	WorkspaceRoot string
	Debug         bool
	Logger        *logging.Logger
	IsLocalSource bool
}

// TemplatesInstallReports installs report templates
func TemplatesInstallReports() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer config.Logger.Sync()

	config.Logger.Info("Starting templates install reports command",
		zap.String("destination", config.Destination),
		zap.String("source", config.Source),
		zap.Bool("debug", config.Debug))

	// Resolve template directory
	templateDir, cleanup, err := resolveTemplateDirectory(config)
	if err != nil {
		config.Logger.Error("Failed to resolve template directory", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer cleanup()

	// Install templates (copy without value replacement)
	if err := installTemplates(config, templateDir); err != nil {
		config.Logger.Error("Failed to install templates", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	config.Logger.Info("Report templates installed successfully",
		zap.String("destination", config.Destination))
	fmt.Printf("✓ Report templates installed successfully to %s\n", config.Destination)

	return 0
}

// resolveTemplateDirectory determines the template directory
func resolveTemplateDirectory(config *Config) (string, func(), error) {
	var templateDir string
	var cleanup func()

	if config.IsLocalSource {
		// Use local directory
		config.Logger.Info("Using local templates",
			zap.String("path", config.Source))
		templateDir = config.Source

		// Verify directory exists
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			return "", nil, fmt.Errorf("local template directory does not exist: %s", config.Source)
		}

		config.Logger.Debug("Local templates validated",
			zap.String("dir", templateDir))
		fmt.Printf("Using local templates from %s\n", templateDir)
		cleanup = func() {}
	} else {
		// Clone from Git repository
		config.Logger.Info("Cloning templates from Git",
			zap.String("url", defaultReportsRepoURL))
		fmt.Printf("Cloning templates from %s...\n", defaultReportsRepoURL)

		cloner := internal.NewGitCloner(defaultReportsRepoURL)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			return "", nil, fmt.Errorf("failed to clone repository: %w", err)
		}

		cleanup = func() {
			if err := cloner.Cleanup(); err != nil {
				config.Logger.Warn("Failed to cleanup temp directory", zap.Error(err))
			}
		}

		// Point to reports subdirectory
		templateDir = filepath.Join(clonedDir, defaultReportsSourcePath)

		// Verify subdirectory exists
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			cleanup()
			return "", nil, fmt.Errorf("template directory does not exist: %s in repository: %s",
				defaultReportsSourcePath, defaultReportsRepoURL)
		}

		config.Logger.Debug("Templates cloned successfully",
			zap.String("clonedDir", clonedDir),
			zap.String("templateDir", templateDir))
		fmt.Printf("✓ Templates cloned successfully\n")

		// Save debug info if enabled
		if config.Debug {
			writeDebugFile(config, "clone-info.txt", fmt.Sprintf(
				"Cloned from: %s\nClone directory: %s\nTemplate directory: %s\n",
				defaultReportsRepoURL, clonedDir, templateDir))
		}
	}

	return templateDir, cleanup, nil
}

// installTemplates copies templates to destination
func installTemplates(config *Config, templateDir string) error {
	config.Logger.Info("Installing templates",
		zap.String("source", templateDir),
		zap.String("destination", config.Destination))
	fmt.Printf("Installing templates to %s...\n", config.Destination)

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
func writeDebugFile(config *Config, filename string, content string) {
	if !config.Debug {
		return
	}

	debugDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "templates", "install")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		config.Logger.Warn("Failed to create debug directory", zap.Error(err))
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		config.Logger.Warn("Failed to write debug file",
			zap.String("file", debugFile),
			zap.Error(err))
	} else {
		config.Logger.Debug("Saved debug file", zap.String("file", debugFile))
	}
}

// parseConfig parses command-line arguments
func parseConfig() (*Config, error) {
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip "binary templates install reports"
	}

	// Parse flags manually
	var source string
	destination := defaultReportsDest
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		case "--source":
			if i+1 < len(args) {
				source = args[i+1]
				i++
			}
		case "--destination":
			if i+1 < len(args) {
				destination = args[i+1]
				i++
			}
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

	// Resolve destination path
	if !filepath.IsAbs(destination) {
		destination = filepath.Join(workspaceRoot, destination)
	}

	// Resolve source path
	isLocalSource := source != ""
	if isLocalSource {
		if !filepath.IsAbs(source) {
			source = filepath.Join(workspaceRoot, source)
		}
	}

	config := &Config{
		Source:        source,
		Destination:   destination,
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
		Logger:        logger,
		IsLocalSource: isLocalSource,
	}

	return config, nil
}
