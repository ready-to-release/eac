package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/render"
	"gopkg.in/yaml.v3"
)

// getCallerCommandName extracts the canonical command name from the calling file
// Returns kebab-case format (e.g., "get-files")
func getCallerCommandName() string {
	// Try to derive from os.Args (the executed command)
	// Format: "go run . get files" -> join args[2:] and kebab-case
	if len(os.Args) >= 2 {
		// Join all command parts and convert to kebab-case
		commandParts := []string{}
		for _, arg := range os.Args[1:] {
			if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
				commandParts = append(commandParts, arg)
			} else {
				break // Stop at first flag
			}
		}
		if len(commandParts) > 0 {
			return strings.Join(commandParts, "-")
		}
	}
	return ""
}

// OutputFormat represents the desired output format
type OutputFormat struct {
	AsYAML           bool
	AsJSON           bool
	AsTOML           bool
	CustomRenderer   string
}

// ParseOutputFlags parses output format flags from command arguments
// commandName should be in kebab-case format (e.g., "get-files")
// Returns the OutputFormat and any error encountered
func ParseOutputFlags(args []string, commandName string) (*OutputFormat, error) {
	format := &OutputFormat{}

	// Get available custom renderers for this specific command
	availableRenderers := render.ListCustomRenderers(commandName)

	for _, arg := range args {
		if arg == "--as-yaml" {
			format.AsYAML = true
		} else if arg == "--as-json" {
			format.AsJSON = true
		} else if arg == "--as-toml" {
			format.AsTOML = true
		} else if strings.HasPrefix(arg, "--as-") {
			// Extract renderer name from --as-<name>
			rendererName := strings.TrimPrefix(arg, "--as-")

			// Check if it's a known custom renderer for this command
			isCustom := false
			for _, name := range availableRenderers {
				if name == rendererName {
					format.CustomRenderer = rendererName
					isCustom = true
					break
				}
			}

			// If not a known flag, show error
			if !isCustom && rendererName != "json" && rendererName != "toml" && rendererName != "yaml" {
				return nil, fmt.Errorf("unknown flag --as-%s for command %s\nAvailable custom renderers: %s",
					rendererName, commandName, strings.Join(availableRenderers, ", "))
			}
		}
	}

	return format, nil
}

// RenderAndOutput takes data, converts to YAML, then renders in the requested format
// commandName should be in kebab-case format (e.g., "get-files")
// This ensures YAML is always the single source of truth
func RenderAndOutput(data interface{}, format *OutputFormat, commandName string) error {
	// First, always marshal to YAML (single source of truth)
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Then render in the requested format
	if format.CustomRenderer != "" {
		// Use custom renderer (pass command name for filtering)
		output, err := render.RenderAsCustom(data, format.CustomRenderer, commandName)
		if err != nil {
			return fmt.Errorf("custom renderer failed: %w", err)
		}
		fmt.Print(output)
	} else if format.AsJSON {
		// Render as JSON
		output, err := render.RenderAsJSON(data)
		if err != nil {
			return fmt.Errorf("failed to render JSON: %w", err)
		}
		fmt.Println(output)
	} else if format.AsTOML {
		// Render as TOML
		output, err := render.RenderAsTOML(data)
		if err != nil {
			return fmt.Errorf("failed to render TOML: %w", err)
		}
		fmt.Print(output)
	} else {
		// Default: output as YAML
		fmt.Print(string(yamlBytes))
	}

	return nil
}

// ExecuteGetCommand is a helper that wraps the common pattern for get commands:
// 1. Parse output format flags
// 2. Execute data fetching function
// 3. Render and output the result
func ExecuteGetCommand(dataFetcher func() (interface{}, error)) int {
	// Get the canonical command name from executed command
	commandName := getCallerCommandName()

	// Parse output format flags (with command-specific filtering)
	format, err := ParseOutputFlags(os.Args[1:], commandName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Fetch the data
	data, err := dataFetcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Render and output (with command name for custom renderer filtering)
	if err := RenderAndOutput(data, format, commandName); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering output: %v\n", err)
		return 1
	}

	return 0
}
