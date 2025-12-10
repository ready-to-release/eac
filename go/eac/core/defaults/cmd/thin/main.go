// Thin repository.yml by removing fields that match type defaults
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

func main() {
	repoRoot := findRepoRoot()
	modulesPath := filepath.Join(paths.EACConfigPath(repoRoot), "repository.yml")
	typesPath := filepath.Join(paths.EACConfigPath(repoRoot), "module-types.yml")

	// Load module types
	typesData, err := os.ReadFile(typesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading module-types.yml: %v\n", err)
		os.Exit(1)
	}

	var typesConfig struct {
		Types []struct {
			Name     string `yaml:"name"`
			Defaults *struct {
				Files *struct {
					Source    []string `yaml:"source,omitempty"`
					Config    []string `yaml:"config,omitempty"`
					Assets    []string `yaml:"assets,omitempty"`
					Tests     []string `yaml:"tests,omitempty"`
					Changelog string   `yaml:"changelog,omitempty"`
				} `yaml:"files,omitempty"`
				Repo *struct {
					Specs    []string `yaml:"specs,omitempty"`
					TestImpl string   `yaml:"test_impl,omitempty"`
					Design   string   `yaml:"design,omitempty"`
				} `yaml:"repo,omitempty"`
			} `yaml:"defaults,omitempty"`
		} `yaml:"types"`
	}
	if err := yaml.Unmarshal(typesData, &typesConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing module-types.yml: %v\n", err)
		os.Exit(1)
	}

	// Build type defaults map
	typeDefaults := make(map[string]struct {
		specsEmpty bool
		config     []string
		source     []string
		assets     []string
	})
	for _, t := range typesConfig.Types {
		td := struct {
			specsEmpty bool
			config     []string
			source     []string
			assets     []string
		}{}
		if t.Defaults != nil && t.Defaults.Repo != nil && t.Defaults.Repo.Specs != nil && len(t.Defaults.Repo.Specs) == 0 {
			td.specsEmpty = true
		}
		if t.Defaults != nil && t.Defaults.Files != nil {
			td.config = t.Defaults.Files.Config
			td.source = t.Defaults.Files.Source
			td.assets = t.Defaults.Files.Assets
		}
		typeDefaults[t.Name] = td
	}

	// Load modules as raw YAML to preserve structure
	modulesData, err := os.ReadFile(modulesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading repository.yml: %v\n", err)
		os.Exit(1)
	}

	var modulesConfig struct {
		Modules []map[string]interface{} `yaml:"modules"`
	}
	if err := yaml.Unmarshal(modulesData, &modulesConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing repository.yml: %v\n", err)
		os.Exit(1)
	}

	removed := 0
	for i, m := range modulesConfig.Modules {
		typeName, _ := m["type"].(string)
		td, hasType := typeDefaults[typeName]

		filesRaw, hasFiles := m["files"].(map[string]interface{})
		if !hasFiles {
			continue
		}

		repoRaw, hasRepo := filesRaw["repo"].(map[string]interface{})
		if hasRepo && hasType && td.specsEmpty {
			specsRaw, hasSpecs := repoRaw["specs"]
			if hasSpecs {
				if specsList, ok := specsRaw.([]interface{}); ok && len(specsList) == 0 {
					delete(repoRaw, "specs")
					removed++
					fmt.Printf("Removed specs: [] from %s (type %s has empty specs default)\n", m["moniker"], typeName)
				}
			}
			// If repo is now empty, remove it
			if len(repoRaw) == 0 {
				delete(filesRaw, "repo")
				fmt.Printf("  Also removed empty repo block from %s\n", m["moniker"])
			}
		}
		modulesConfig.Modules[i] = m
	}

	// Write updated modules
	output, err := yaml.Marshal(modulesConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling: %v\n", err)
		os.Exit(1)
	}

	// Add header comment
	header := "# Repository Module Contracts\n# Version: 0.2.0\n\n"
	if err := os.WriteFile(modulesPath, append([]byte(header), output...), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nRemoved %d specs: [] entries\n", removed)
}

func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
