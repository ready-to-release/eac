// Serialize loaded module configs to compare before/after defaults changes
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ready-to-release/eac/go/core/domain/modules"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: serialize <output-file>")
		fmt.Println("Loads repository.yml and serializes resolved config to output file")
		os.Exit(1)
	}
	outputFile := os.Args[1]

	repoRoot := findRepoRoot()

	registry, err := modules.LoadFromWorkspace(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading modules: %v\n", err)
		os.Exit(1)
	}

	allModules := registry.All()
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Moniker < allModules[j].Moniker
	})

	// Output structure using components instead of files
	type ComponentOutput struct {
		Root     string   `yaml:"root,omitempty"`
		Patterns []string `yaml:"patterns,omitempty"`
	}
	type ModuleOutput struct {
		Moniker     string                     `yaml:"moniker"`
		Type        string                     `yaml:"type"`
		Description string                     `yaml:"description"`
		DependsOn   []string                   `yaml:"depends_on"`
		Changelog   string                     `yaml:"changelog,omitempty"`
		Components  map[string]ComponentOutput `yaml:"components"`
	}

	output := make([]ModuleOutput, 0, len(allModules))
	for _, m := range allModules {
		out := ModuleOutput{
			Moniker:     m.Moniker,
			Type:        m.GetComponentTypesDisplay(),
			Description: m.Description,
			DependsOn:   m.DependsOn,
			Changelog:   m.GetChangelogPath(),
			Components:  make(map[string]ComponentOutput),
		}

		// Serialize component roots
		for pkgName, root := range m.GetComponentRoots() {
			pkg := ComponentOutput{Root: root}
			// Get patterns from component if available
			if m.Components[pkgName] != nil && m.Components[pkgName].Patterns != nil {
				patterns := m.Components[pkgName].Patterns
				pkg.Patterns = append(pkg.Patterns, patterns.Source...)
				pkg.Patterns = append(pkg.Patterns, patterns.Tests...)
				pkg.Patterns = append(pkg.Patterns, patterns.Config...)
			}
			out.Components[pkgName] = pkg
		}

		output = append(output, out)
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %d modules to %s\n", len(output), outputFile)
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "." // Fallback to current directory
	}
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
