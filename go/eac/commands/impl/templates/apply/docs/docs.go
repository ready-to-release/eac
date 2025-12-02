// Command: templates apply docs
// Short: Apply documentation templates with value replacements
// Long: Apply documentation templates by processing files and replacing {{ .Variable }} placeholders with values.
// Long:
// Long: Template Source (--source):
// Long:   If provided: uses local directory containing template files
// Long:   If omitted: automatically clones from https://github.com/ready-to-release/eac
// Long:
// Long: Value Replacement (--input-json):
// Long:   Provide a JSON file with key-value pairs: {"ProjectName": "MyApp", "Author": "John"}
// Long:   If omitted: placeholders remain unchanged ({{ .ProjectName }} stays as-is)
// Long:
// Long: Examples:
// Long:   templates apply docs
// Long:   templates apply docs --source ./my-templates
// Long:   templates apply docs --input-json values.json
// Long:   templates apply docs --destination ./output --debug
// Flag.source: type=string, usage=Local template directory (default: clones from GitHub)
// Flag.destination: type=string, usage=Output directory (default: .docs/reference)
// Flag.input-json: type=string, usage=JSON file with replacement values {"Key": "Value"}
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save detailed logs to out/logs/templates/apply/
package docs

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

const (
	defaultDocsRepoURL    = "https://github.com/ready-to-release/eac"
	defaultDocsSourcePath = "templates/docs"
	defaultDocsDest       = ".docs/reference"
)

var log = logging.C()

func init() {
	registry.Register(TemplatesApplyDocs)
}

// Config holds configuration for the docs apply command
type Config struct {
	Source        string
	Destination   string
	ValuesFile    string
	WorkspaceRoot string
	Debug         bool
	Logger        *logging.Logger
	IsLocalSource bool
}

// TemplatesApplyDocs applies documentation templates
func TemplatesApplyDocs() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}
	defer config.Logger.Sync()

	config.Logger.Info("Starting templates apply docs command",
		zap.String("destination", config.Destination),
		zap.String("source", config.Source),
		zap.Bool("hasValues", config.ValuesFile != ""),
		zap.Bool("debug", config.Debug))

	// Resolve template directory
	templateDir, cleanup, err := resolveTemplateDirectory(config)
	if err != nil {
		config.Logger.Error("Failed to resolve template directory", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}
	defer cleanup()

	// Load values if provided
	values, err := loadValues(config)
	if err != nil {
		config.Logger.Error("Failed to load values", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}

	// Apply templates
	if err := applyTemplates(config, templateDir, values); err != nil {
		config.Logger.Error("Failed to apply templates", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}

	config.Logger.Info("Documentation templates applied successfully",
		zap.String("destination", config.Destination))
	log.Infof("✓ Documentation templates applied successfully to %s", config.Destination)

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
		log.Infof("Using local templates from %s", templateDir)
		cleanup = func() {}
	} else {
		// Clone from Git repository
		config.Logger.Info("Cloning templates from Git",
			zap.String("url", defaultDocsRepoURL))
		log.Infof("Cloning templates from %s...", defaultDocsRepoURL)

		cloner := internal.NewGitCloner(defaultDocsRepoURL)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			return "", nil, fmt.Errorf("failed to clone repository: %w", err)
		}

		cleanup = func() {
			if err := cloner.Cleanup(); err != nil {
				config.Logger.Warn("Failed to cleanup temp directory", zap.Error(err))
			}
		}

		// Point to docs subdirectory
		templateDir = filepath.Join(clonedDir, defaultDocsSourcePath)

		// Verify subdirectory exists
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			cleanup()
			return "", nil, fmt.Errorf("template directory does not exist: %s in repository: %s",
				defaultDocsSourcePath, defaultDocsRepoURL)
		}

		config.Logger.Debug("Templates cloned successfully",
			zap.String("clonedDir", clonedDir),
			zap.String("templateDir", templateDir))
		log.Info("✓ Templates cloned successfully")

		// Save debug info if enabled
		if config.Debug {
			writeDebugFile(config, "clone-info.txt", fmt.Sprintf(
				"Cloned from: %s\nClone directory: %s\nTemplate directory: %s\n",
				defaultDocsRepoURL, clonedDir, templateDir))
		}
	}

	return templateDir, cleanup, nil
}

// loadValues loads template values from JSON file if provided
func loadValues(config *Config) (internal.TemplateValues, error) {
	if config.ValuesFile == "" {
		config.Logger.Debug("No values file provided, using empty values")
		return nil, nil
	}

	config.Logger.Info("Loading template values",
		zap.String("file", config.ValuesFile))

	values, err := internal.LoadValuesFromJSON(config.ValuesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load values: %w", err)
	}

	config.Logger.Debug("Values loaded successfully",
		zap.Int("count", len(values)))
	log.Infof("Loaded %d replacement values", len(values))

	// Save debug info if enabled
	if config.Debug {
		var content string
		content += fmt.Sprintf("Values file: %s\n", config.ValuesFile)
		content += fmt.Sprintf("Value count: %d\n\n", len(values))
		for k, v := range values {
			content += fmt.Sprintf("%s = %v\n", k, v)
		}
		writeDebugFile(config, "values.txt", content)
	}

	return values, nil
}

// applyTemplates renders and applies templates
func applyTemplates(config *Config, templateDir string, values internal.TemplateValues) error {
	config.Logger.Info("Applying templates",
		zap.String("source", templateDir),
		zap.String("destination", config.Destination),
		zap.Int("valueCount", len(values)))
	log.Infof("Applying templates to %s...", config.Destination)

	renderer := internal.NewRenderer(templateDir, config.Destination, values)
	if err := renderer.RenderTemplates(); err != nil {
		return fmt.Errorf("failed to apply templates: %w", err)
	}

	config.Logger.Debug("Templates applied successfully")

	// Save debug info if enabled
	if config.Debug {
		writeDebugFile(config, "apply-summary.txt", fmt.Sprintf(
			"Template source: %s\nDestination: %s\nValues count: %d\nSuccess: true\n",
			templateDir, config.Destination, len(values)))
	}

	return nil
}

// writeDebugFile writes debug content to file
func writeDebugFile(config *Config, filename string, content string) {
	if !config.Debug {
		return
	}

	debugDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "templates", "apply")
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
		args = os.Args[4:] // Skip "binary templates apply docs"
	}

	// Parse flags manually
	var source string
	destination := defaultDocsDest
	var valuesFile string
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
		case "--input-json":
			if i+1 < len(args) {
				valuesFile = args[i+1]
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

	// Resolve source and values paths
	isLocalSource := source != ""
	if isLocalSource {
		if !filepath.IsAbs(source) {
			source = filepath.Join(workspaceRoot, source)
		}
	}

	if valuesFile != "" {
		if !filepath.IsAbs(valuesFile) {
			valuesFile = filepath.Join(workspaceRoot, valuesFile)
		}

		// Validate values file exists
		if _, err := os.Stat(valuesFile); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("values file does not exist: %s", valuesFile)
			}
			return nil, fmt.Errorf("cannot access values file: %w", err)
		}
	}

	config := &Config{
		Source:        source,
		Destination:   destination,
		ValuesFile:    valuesFile,
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
		Logger:        logger,
		IsLocalSource: isLocalSource,
	}

	return config, nil
}
