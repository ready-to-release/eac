// Command: validate config
// Short: Validate effective configuration from all sources
// Long: Validate effective configuration from all sources.
// Long:
// Long: This command validates the complete configuration stack:
// Long:   1. Contract defaults (from contracts/eac-*/defaults/)
// Long:   2. User overrides (from .eac/)
// Long:   3. Personal overrides (from .eac/*.personal.yml)
// Long:
// Long: Validation phases:
// Long:   - File Checks: All config files are readable
// Long:   - Schema Validation: YAML matches JSON schemas
// Long:   - Cross-Reference: Dependencies exist, component types valid
// Long:   - Completeness: Required fields present
// Long:
// Long: Expected Output:
// Long:   Shows validation status for each config layer. Reports any errors
// Long:   or warnings found. Exit code 0 if valid, 1 if errors.
// Long:
// Long: Example:
// Long:   validate config
// Long:   validate config --strict
// Long:   validate config --format json
// Flag: --strict: Treat warnings as errors (default: false)
// Flag: --format: Output format: text, json, github (default: text)
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

func init() {
	registry.Register(ValidateConfigCmd)
}

// ConfigValidationResult holds the result of config validation.
type ConfigValidationResult struct {
	Valid       bool              `json:"valid"`
	Errors      []ConfigIssue     `json:"errors,omitempty"`
	Warnings    []ConfigIssue     `json:"warnings,omitempty"`
	FilesLoaded []ConfigFileInfo  `json:"files_loaded"`
}

// ConfigIssue represents a validation issue.
type ConfigIssue struct {
	File    string `json:"file"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// ConfigFileInfo represents a loaded config file.
type ConfigFileInfo struct {
	Path   string `json:"path"`
	Layer  string `json:"layer"` // "contract", "user", "personal"
	Exists bool   `json:"exists"`
	Valid  bool   `json:"valid"`
	Error  string `json:"error,omitempty"`
}

// ValidateConfigCmd validates effective configuration from all sources.
func ValidateConfigCmd() int {
	// Parse flags
	strict := false
	format := "text"

	args := os.Args[2:] // Skip program name and "validate"
	if len(args) > 0 && args[0] == "config" {
		args = args[1:] // Skip "config"
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--strict":
			strict = true
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--help", "-h":
			reg := registry.GetCommand("validate config")
			if reg != nil {
				fmt.Println(reg.Short)
				fmt.Println()
				fmt.Println(reg.Long)
			}
			return 0
		}
	}

	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Get paths from config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return 1
	}

	repoRoot := cfg.RepoRoot
	configRoot := cfg.ConfigRoot

	result := runConfigValidation(repoRoot, configRoot)

	// Apply strict mode
	if strict && len(result.Warnings) > 0 {
		result.Valid = false
		for _, w := range result.Warnings {
			result.Errors = append(result.Errors, w)
		}
		result.Warnings = nil
	}

	// Output results
	switch format {
	case "json":
		configOutputJSON(result)
	case "github":
		configOutputGitHub(result)
	default:
		configOutputText(result)
	}

	if result.Valid {
		return 0
	}
	return 1
}

// runConfigValidation performs all validation checks.
func runConfigValidation(repoRoot, configRoot string) *ConfigValidationResult {
	result := &ConfigValidationResult{
		Valid: true,
	}

	// Phase 1: File Checks
	runFileChecks(repoRoot, configRoot, result)

	// Phase 2: Load and validate configuration
	runLoadValidation(repoRoot, configRoot, result)

	// Phase 3: Cross-reference validation
	runCrossReferenceValidation(repoRoot, configRoot, result)

	return result
}

// runFileChecks checks that all config files exist and are readable.
func runFileChecks(repoRoot, configRoot string, result *ConfigValidationResult) {
	// Check contract defaults
	contractFiles := []struct {
		contract string
		file     string
	}{
		{"core", "repository.yml"},
		{"core", "component-types.yml"},
		{"core", "books.yml"},
		{"security", "scanners.yml"},
		{"security", "policies.yml"},
		{"testing", "suites.yml"},
		{"testing", "tags.yml"},
		{"tools", "tools.yml"},
	}

	for _, cf := range contractFiles {
		path := filepath.Join(repoRoot, "contracts", cf.contract, paths.DefaultsVersion, "defaults", cf.file)
		info := ConfigFileInfo{
			Path:  path,
			Layer: "contract",
		}

		if _, err := os.Stat(path); err == nil {
			info.Exists = true
			info.Valid = true
		} else if os.IsNotExist(err) {
			info.Exists = false
			info.Valid = true // Optional contract files are OK to be missing
		} else {
			info.Exists = false
			info.Valid = false
			info.Error = err.Error()
			result.Errors = append(result.Errors, ConfigIssue{
				File:    path,
				Message: fmt.Sprintf("cannot read file: %v", err),
			})
			result.Valid = false
		}

		result.FilesLoaded = append(result.FilesLoaded, info)
	}

	// Check user config files
	userFiles := []string{
		"repository.yml",
		"environments.yml",
	}

	for _, file := range userFiles {
		path := filepath.Join(configRoot, file)
		info := ConfigFileInfo{
			Path:  path,
			Layer: "user",
		}

		if _, err := os.Stat(path); err == nil {
			info.Exists = true
			info.Valid = true
		} else if os.IsNotExist(err) {
			info.Exists = false
			info.Valid = true // User files are optional
		} else {
			info.Exists = false
			info.Valid = false
			info.Error = err.Error()
		}

		result.FilesLoaded = append(result.FilesLoaded, info)
	}

	// Check personal config files
	personalConfigs := config.AllPersonalConfigInfo(configRoot)
	for _, pc := range personalConfigs {
		if pc.HasPersonal {
			info := ConfigFileInfo{
				Path:   pc.PersonalPath,
				Layer:  "personal",
				Exists: true,
				Valid:  true,
			}
			result.FilesLoaded = append(result.FilesLoaded, info)
		}
	}
}

// runLoadValidation attempts to load all configurations.
func runLoadValidation(repoRoot, configRoot string, result *ConfigValidationResult) {
	opts := config.LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        false,
	}

	cfg, err := config.Load(opts)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ConfigIssue{
			Message: fmt.Sprintf("failed to load configuration: %v", err),
		})
		return
	}

	// Try to load each config type
	configTypes := []struct {
		name   string
		loader func(bool) error
	}{
		{"repository", cfg.LoadRepository},
		{"environments", cfg.LoadEnvironments},
		{"component-types", cfg.LoadComponentTypes},
	}

	for _, ct := range configTypes {
		if err := ct.loader(true); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ConfigIssue{
				File:    ct.name + ".yml",
				Message: err.Error(),
			})
		}
	}

	// Try loading the new contract-based configs
	if err := cfg.LoadSecurity(); err != nil {
		result.Warnings = append(result.Warnings, ConfigIssue{
			File:    "security",
			Message: fmt.Sprintf("security config: %v", err),
		})
	}

	if err := cfg.LoadTesting(); err != nil {
		result.Warnings = append(result.Warnings, ConfigIssue{
			File:    "testing",
			Message: fmt.Sprintf("testing config: %v", err),
		})
	}
}

// runCrossReferenceValidation checks that referenced items exist.
func runCrossReferenceValidation(repoRoot, configRoot string, result *ConfigValidationResult) {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return // Already reported in load phase
	}

	// Check that module component types are valid
	if cfg.Repository != nil && cfg.ComponentTypes != nil {
		for _, mod := range cfg.Repository.Modules {
			for compName, comp := range mod.Components {
				compType := comp.Type
				if compType == "" {
					compType = compName
				}
				if cfg.ComponentTypes.Get(compType) == nil {
					result.Warnings = append(result.Warnings, ConfigIssue{
						File:    "repository.yml",
						Message: fmt.Sprintf("module %q component %q has unknown type %q", mod.Moniker, compName, compType),
					})
				}
			}
		}
	}

	// Check that module dependencies reference existing modules
	if cfg.Repository != nil {
		moduleIndex := make(map[string]bool)
		for _, mod := range cfg.Repository.Modules {
			moduleIndex[mod.Moniker] = true
		}

		for _, mod := range cfg.Repository.Modules {
			for _, dep := range mod.DependsOn {
				if !moduleIndex[dep] {
					result.Warnings = append(result.Warnings, ConfigIssue{
						File:    "repository.yml",
						Message: fmt.Sprintf("module %q has unknown dependency %q", mod.Moniker, dep),
					})
				}
			}
		}
	}
}

// configOutputText outputs results in human-readable text format.
func configOutputText(result *ConfigValidationResult) {
	log.Info("Configuration Validation")
	log.Info("===================================================")
	log.Info("")

	// Files loaded
	log.Info("Files Loaded:")
	for _, f := range result.FilesLoaded {
		status := "OK"
		if !f.Valid {
			status = "ERROR"
		} else if !f.Exists {
			status = "not found"
		}
		log.Infof("  [%s] %s (%s)", f.Layer, filepath.Base(f.Path), status)
	}
	log.Info("")

	// Errors
	if len(result.Errors) > 0 {
		log.Info("Errors:")
		for _, e := range result.Errors {
			if e.File != "" {
				log.Errorf("  %s: %s", e.File, e.Message)
			} else {
				log.Errorf("  %s", e.Message)
			}
		}
		log.Info("")
	}

	// Warnings
	if len(result.Warnings) > 0 {
		log.Info("Warnings:")
		for _, w := range result.Warnings {
			if w.File != "" {
				log.Warnf("  %s: %s", w.File, w.Message)
			} else {
				log.Warnf("  %s", w.Message)
			}
		}
		log.Info("")
	}

	// Summary
	if result.Valid {
		log.Info("Configuration is valid")
	} else {
		log.Errorf("Configuration validation failed with %d error(s)", len(result.Errors))
	}
}

// configOutputJSON outputs results in JSON format.
func configOutputJSON(result *ConfigValidationResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Errorf("Failed to serialize result: %v", err)
		return
	}
	fmt.Println(string(data))
}

// configOutputGitHub outputs results in GitHub Actions format.
func configOutputGitHub(result *ConfigValidationResult) {
	for _, e := range result.Errors {
		if e.File != "" && e.Line > 0 {
			fmt.Printf("::error file=%s,line=%d::%s\n", e.File, e.Line, e.Message)
		} else if e.File != "" {
			fmt.Printf("::error file=%s::%s\n", e.File, e.Message)
		} else {
			fmt.Printf("::error::%s\n", e.Message)
		}
	}
	for _, w := range result.Warnings {
		if w.File != "" && w.Line > 0 {
			fmt.Printf("::warning file=%s,line=%d::%s\n", w.File, w.Line, w.Message)
		} else if w.File != "" {
			fmt.Printf("::warning file=%s::%s\n", w.File, w.Message)
		} else {
			fmt.Printf("::warning::%s\n", w.Message)
		}
	}
}
