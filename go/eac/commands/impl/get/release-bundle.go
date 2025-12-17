// Command: get release-bundle
// Description: Get release bundle configuration with module details
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
// Long:
// Long: Expected Output:
// Long: Structured release bundle configuration including:
// Long:   - title_format: Template for release title
// Long:   - headline: Map of label -> module info for title modules
// Long:   - categories: Grouped modules with their details
// Long: Use this in CI to create release notes without hardcoding module names.
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetReleaseBundle)
}

// ReleaseBundleOutput is the structured output for get release-bundle
type ReleaseBundleOutput struct {
	TitleFormat string                        `yaml:"title_format" json:"title_format"`
	Headline    map[string]HeadlineModule     `yaml:"headline" json:"headline"`
	Categories  []ReleaseBundleCategoryOutput `yaml:"categories" json:"categories"`
}

// HeadlineModule contains info about a headline module
type HeadlineModule struct {
	Moniker    string `yaml:"moniker" json:"moniker"`
	Name       string `yaml:"name" json:"name"`
	Versioning string `yaml:"versioning" json:"versioning"`
}

// ReleaseBundleCategoryOutput contains category info with enriched module details
type ReleaseBundleCategoryOutput struct {
	Name        string                `yaml:"name" json:"name"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	Modules     []ReleaseBundleModule `yaml:"modules" json:"modules"`
}

// ReleaseBundleModule contains module info for release bundle
type ReleaseBundleModule struct {
	Moniker    string `yaml:"moniker" json:"moniker"`
	Name       string `yaml:"name" json:"name"`
	Versioning string `yaml:"versioning" json:"versioning"`
}

// releaseBundleFlags defines valid flags for the get release-bundle command
var releaseBundleFlags = []flags.FlagDefinition{
	{Name: "--as-yaml", HasValue: false, ValueType: "bool"},
	{Name: "--as-json", HasValue: false, ValueType: "bool"},
}

func GetReleaseBundle() int {
	// Validate flags before parsing
	if err := flags.ValidateFlags(os.Args[3:], releaseBundleFlags); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return getReleaseBundleData(workspaceRoot)
	})
}

func getReleaseBundleData(workspaceRoot string) (*ReleaseBundleOutput, error) {
	// Load module contracts via the reports API (same as get modules)
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load modules: %w", err)
	}

	// Find the release module with release_bundle config
	var releaseBundle *struct {
		TitleFormat string
		Headline    map[string]string
		Categories  []struct {
			Name        string
			Description string
			Modules     []string
		}
	}

	for _, mod := range report.Modules {
		if mod.ReleaseBundle != nil {
			releaseBundle = &struct {
				TitleFormat string
				Headline    map[string]string
				Categories  []struct {
					Name        string
					Description string
					Modules     []string
				}
			}{
				TitleFormat: mod.ReleaseBundle.TitleFormat,
				Headline:    mod.ReleaseBundle.Headline,
			}
			for _, cat := range mod.ReleaseBundle.Categories {
				releaseBundle.Categories = append(releaseBundle.Categories, struct {
					Name        string
					Description string
					Modules     []string
				}{
					Name:        cat.Name,
					Description: cat.Description,
					Modules:     cat.Modules,
				})
			}
			break
		}
	}

	if releaseBundle == nil {
		return nil, fmt.Errorf("no module with release_bundle configuration found")
	}

	// Build output
	output := &ReleaseBundleOutput{
		TitleFormat: releaseBundle.TitleFormat,
		Headline:    make(map[string]HeadlineModule),
		Categories:  []ReleaseBundleCategoryOutput{},
	}

	// Populate headline modules
	for label, moniker := range releaseBundle.Headline {
		mod, ok := report.GetModuleByMoniker(moniker)
		if !ok {
			return nil, fmt.Errorf("headline module %q not found", moniker)
		}
		versioning := ""
		if mod.Versioning != nil {
			versioning = mod.Versioning.Scheme
		}
		output.Headline[label] = HeadlineModule{
			Moniker:    moniker,
			Name:       mod.Name,
			Versioning: versioning,
		}
	}

	// Populate categories with enriched module info
	for _, cat := range releaseBundle.Categories {
		catOutput := ReleaseBundleCategoryOutput{
			Name:        cat.Name,
			Description: cat.Description,
			Modules:     []ReleaseBundleModule{},
		}

		for _, moniker := range cat.Modules {
			mod, ok := report.GetModuleByMoniker(moniker)
			if !ok {
				return nil, fmt.Errorf("module %q in category %q not found", moniker, cat.Name)
			}
			versioning := ""
			if mod.Versioning != nil {
				versioning = mod.Versioning.Scheme
			}
			catOutput.Modules = append(catOutput.Modules, ReleaseBundleModule{
				Moniker:    moniker,
				Name:       mod.Name,
				Versioning: versioning,
			})
		}

		output.Categories = append(output.Categories, catOutput)
	}

	return output, nil
}
