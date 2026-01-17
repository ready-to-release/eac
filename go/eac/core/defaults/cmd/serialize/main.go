// Serialize loaded module configs to compare before/after default removal
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
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

	// Full serialization of all fields
	type RepoFiles struct {
		Specs    []string `yaml:"specs"`
		TestImpl string   `yaml:"test_impl,omitempty"`
		Design   string   `yaml:"design,omitempty"`
		Other    []string `yaml:"other"`
		Exclude  []string `yaml:"exclude"`
	}
	type Files struct {
		Root      string    `yaml:"root"`
		Source    []string  `yaml:"source"`
		Config    []string  `yaml:"config"`
		Assets    []string  `yaml:"assets"`
		Tests     []string  `yaml:"tests"`
		Exclude   []string  `yaml:"exclude"`
		Changelog string    `yaml:"changelog"`
		Repo      RepoFiles `yaml:"repo"`
	}
	type ModuleOutput struct {
		Moniker     string   `yaml:"moniker"`
		Name        string   `yaml:"name"`
		Type        string   `yaml:"type"`
		Description string   `yaml:"description"`
		DependsOn   []string `yaml:"depends_on"`
		Files       Files    `yaml:"files"`
	}

	output := make([]ModuleOutput, 0, len(allModules))
	for _, m := range allModules {
		out := ModuleOutput{
			Moniker:     m.Moniker,
			Name:        m.Name,
			Type:        m.Type,
			Description: m.Description,
			DependsOn:   m.DependsOn,
			Files: Files{
				Root:      m.Files.Root,
				Source:    m.Files.Source,
				Config:    m.Files.Config,
				Assets:    m.Files.Assets,
				Tests:     m.Files.Tests,
				Exclude:   m.Files.Exclude,
				Changelog: m.Files.Changelog,
				Repo: RepoFiles{
					Specs:    m.Files.Repo.Specs,
					TestImpl: m.Files.Repo.TestImpl,
					Design:   m.Files.Repo.Design,
					Other:    m.Files.Repo.Other,
					Exclude:  m.Files.Repo.Exclude,
				},
			},
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
