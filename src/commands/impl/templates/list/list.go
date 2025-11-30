// Command: templates list
// Description: List all placeholder variables found in template files
// Short: List template placeholder variables
// Long: The templates list command scans template files and displays all placeholder variables.
// Long: Placeholders use {{ .VariableName }} syntax and can be replaced during template application.
// Long: Use --debug to save scan results and intermediate outputs for troubleshooting.
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save scan results to 'out/logs/templates/list' directory
// Flag.template: type=string, usage=Git repository URL or local directory to scan (default: https://github.com/ready-to-release/eac)
// Usage: templates list [--template <git-repo-url|local-path>] [--debug]
// Flags: --template (source location), --debug (save debug outputs)
// HasSideEffects: false
package list

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

func init() {
	registry.Register(TemplatesList)
}

// TemplatesList scans templates and lists all placeholder variables
func TemplatesList() int {
	// Parse configuration from command-line args
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer config.Logger.Sync()

	config.Logger.Info("Starting templates list command",
		zap.String("source", config.TemplateSource),
		zap.Bool("debug", config.Debug))

	// Determine template directory
	templateDir, cleanup, err := resolveTemplateDirectory(config)
	if err != nil {
		config.Logger.Error("Failed to resolve template directory", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer cleanup()

	// Scan templates
	placeholderInfos, err := scanTemplates(config, templateDir)
	if err != nil {
		config.Logger.Error("Failed to scan templates", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Display results
	displayPlaceholders(config.TemplateSource, placeholderInfos)

	// Save debug output if enabled
	if config.Debug {
		saveDebugOutput(config, placeholderInfos)
	}

	config.Logger.Info("Templates list completed successfully",
		zap.Int("placeholderCount", len(placeholderInfos)))

	return 0
}

// resolveTemplateDirectory determines the template directory to scan
// Returns the directory path and a cleanup function
func resolveTemplateDirectory(config *Config) (string, func(), error) {
	var templateDir string
	var cleanup func()

	// Check if template is a Git repository URL or local path
	if internal.IsGitRepository(config.TemplateSource) {
		// Clone repository to temp directory
		config.Logger.Info("Cloning templates from Git repository",
			zap.String("url", config.TemplateSource))
		fmt.Printf("Cloning templates from %s...\n", config.TemplateSource)

		cloner := internal.NewGitCloner(config.TemplateSource)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			return "", nil, fmt.Errorf("failed to clone repository: %w", err)
		}

		// Point to the templates subdirectory within the cloned repository
		templateDir = filepath.Join(clonedDir, "templates")
		cleanup = func() {
			if err := cloner.Cleanup(); err != nil {
				config.Logger.Warn("Failed to cleanup temp directory", zap.Error(err))
			}
		}

		config.Logger.Debug("Templates cloned successfully",
			zap.String("clonedDir", clonedDir),
			zap.String("templateDir", templateDir))
		fmt.Printf("✓ Templates cloned successfully\n\n")
	} else {
		// Use local directory
		config.Logger.Info("Using local templates directory",
			zap.String("path", config.TemplateSource))
		templateDir = config.TemplateSource
		cleanup = func() {} // No cleanup needed for local paths
	}

	return templateDir, cleanup, nil
}

// scanTemplates scans the template directory for placeholders
func scanTemplates(config *Config, templateDir string) ([]internal.PlaceholderInfo, error) {
	config.Logger.Debug("Scanning templates directory",
		zap.String("dir", templateDir))

	scanner := internal.NewPlaceholderScanner(templateDir)
	placeholderInfos, err := scanner.ScanWithLocations()
	if err != nil {
		return nil, fmt.Errorf("failed to scan templates: %w", err)
	}

	config.Logger.Debug("Scan completed",
		zap.Int("placeholderCount", len(placeholderInfos)))

	return placeholderInfos, nil
}

// displayPlaceholders prints the found placeholder variables with their locations
func displayPlaceholders(templateDir string, placeholderInfos []internal.PlaceholderInfo) {
	fmt.Printf("Template Placeholders in '%s':\n", templateDir)
	fmt.Println("----------------------------")

	if len(placeholderInfos) == 0 {
		fmt.Println("No placeholders found.")
		return
	}

	// Display each placeholder with its file locations
	for _, info := range placeholderInfos {
		fmt.Printf("  {{ .%s }}\n", info.Name)
		for _, file := range info.Files {
			fmt.Printf("    - %s\n", file)
		}
	}

	fmt.Printf("\nTotal: %d placeholders\n", len(placeholderInfos))
	fmt.Println("\nTo use these templates, provide a values.json file with these keys:")
	fmt.Println("{")
	for i, info := range placeholderInfos {
		if i == len(placeholderInfos)-1 {
			fmt.Printf("  \"%s\": \"value\"\n", info.Name)
		} else {
			fmt.Printf("  \"%s\": \"value\",\n", info.Name)
		}
	}
	fmt.Println("}")
}

// saveDebugOutput saves scan results to debug files
func saveDebugOutput(config *Config, placeholderInfos []internal.PlaceholderInfo) {
	debugDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "templates", "list")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		config.Logger.Warn("Failed to create debug directory", zap.Error(err))
		return
	}

	// Build debug output content
	var content string
	content += fmt.Sprintf("Template Source: %s\n", config.TemplateSource)
	content += fmt.Sprintf("Total Placeholders: %d\n\n", len(placeholderInfos))

	for _, info := range placeholderInfos {
		content += fmt.Sprintf("Placeholder: %s\n", info.Name)
		content += "  Files:\n"
		for _, file := range info.Files {
			content += fmt.Sprintf("    - %s\n", file)
		}
		content += "\n"
	}

	debugFile := filepath.Join(debugDir, "scan-results.txt")
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		config.Logger.Warn("Failed to write debug file", zap.Error(err))
	} else {
		config.Logger.Debug("Saved scan results", zap.String("file", debugFile))
	}
}

// parseConfig parses command-line arguments into configuration
func parseConfig() (*Config, error) {
	// Get flags starting after "templates list"
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	// Parse flags manually
	var templateSource string
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		case "--template":
			if i+1 < len(args) {
				templateSource = args[i+1]
				i++ // Skip next arg
			}
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		}
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Create config with nil logger initially
	config := NewConfig(workspaceRoot, debug, nil)

	// Initialize logger
	var loggerInst *logging.Logger
	if debug {
		loggerInst, err = logging.NewWithDebug("templates", workspaceRoot)
	} else {
		loggerInst, err = logging.NewDefault("templates", workspaceRoot)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	config.Logger = loggerInst

	// Set template source if provided
	if templateSource != "" {
		// Resolve relative paths against workspace root
		if !internal.IsGitRepository(templateSource) && !filepath.IsAbs(templateSource) {
			templateSource = filepath.Join(workspaceRoot, templateSource)
		}
		config.TemplateSource = templateSource
	}

	return config, nil
}

// printUsage prints command usage information
func printUsage() {
	fmt.Println("List all placeholder variables found in template files")
	fmt.Println()
	fmt.Println("Usage: r2r templates list [--template <source>] [--debug]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --template <source>  Git repository URL or local directory to scan")
	fmt.Println("                       (default: https://github.com/ready-to-release/eac)")
	fmt.Println("  --debug, -d          Enable debug mode to save scan results")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List placeholders from default repository")
	fmt.Println("  r2r templates list")
	fmt.Println()
	fmt.Println("  # List placeholders from local directory with debug output")
	fmt.Println("  r2r templates list --template ./my-templates --debug")
}
