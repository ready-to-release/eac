package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/cli/eac/impl/templates/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

// Config holds configuration for a template install command.
type Config struct {
	Destination   string
	WorkspaceRoot string
	Debug         bool
}

// TemplateDirResolver resolves the template source directory from a root path.
type TemplateDirResolver func(root string) string

// ConfigParser builds a Config given the workspace root and parsed args.
type ConfigParser func(workspaceRoot string, args []string) (*Config, error)

// Params defines the parameters for a generic template install command.
type Params struct {
	// TemplateName is the human-readable name used in log messages (e.g. "AI", "Documentation").
	TemplateName string

	// TemplateDir resolves the template source directory from a root path.
	TemplateDir TemplateDirResolver

	// ParseConfig builds a Config from the workspace root and remaining CLI args.
	ParseConfig ConfigParser
}

// Run executes the generic template install workflow.
// It validates flags, parses configuration, resolves the template directory,
// installs templates, and logs the result.
func Run(p Params) int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse configuration
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:]
	}

	cfg, err := p.ParseConfig(workspaceRoot, args)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Configure logging system (logs to out/commands.log)
	if err := logging.ConfigureLoggingSimple(cfg.WorkspaceRoot, "commands", nil, cfg.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Debugf("Starting templates install %s command: destination=%s, debug=%v",
		p.TemplateName, cfg.Destination, cfg.Debug)

	// Resolve template directory
	templateDir, err := resolveTemplateDirectory(cfg, p.TemplateDir)
	if err != nil {
		log.Debugf("Failed to resolve template directory: error=%v", err)
		log.Errorf("%v", err)
		return 1
	}

	// Install templates (copy without value replacement)
	if err := installTemplates(cfg, templateDir); err != nil {
		log.Debugf("Failed to install templates: error=%v", err)
		log.Errorf("%v", err)
		return 1
	}

	log.Debugf("%s templates installed successfully: destination=%s", p.TemplateName, cfg.Destination)
	log.Infof("\u2713 %s templates installed successfully to %s", p.TemplateName, cfg.Destination)

	return 0
}

// resolveTemplateDirectory determines the template directory (always local).
func resolveTemplateDirectory(cfg *Config, resolve TemplateDirResolver) (string, error) {
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

	templateDir := resolve(root)

	// Verify directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return "", fmt.Errorf("template directory does not exist: %s", templateDir)
	}

	log.Debugf("Local templates validated: dir=%s", templateDir)
	log.Infof("Using templates from %s", templateDir)

	return templateDir, nil
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

// DebugOnlyConfig builds a ConfigParser that only supports --debug/-d flag
// and uses a fixed destination path computed from the workspace root.
func DebugOnlyConfig(destinationFn func(workspaceRoot string) string) ConfigParser {
	return func(workspaceRoot string, args []string) (*Config, error) {
		debug := false
		for _, arg := range args {
			if arg == "--debug" || arg == "-d" {
				debug = true
			}
		}

		return &Config{
			Destination:   destinationFn(workspaceRoot),
			WorkspaceRoot: workspaceRoot,
			Debug:         debug,
		}, nil
	}
}
