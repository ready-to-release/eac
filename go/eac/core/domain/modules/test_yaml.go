package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/domain"
	"gopkg.in/yaml.v3"
)

func TestYAML() {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return
	}
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	yamlPath := filepath.Join(workspaceRoot, ".r2r", "eac", "repository.yml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var config struct {
		Modules []domain.BaseContract `yaml:"modules"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error unmarshaling YAML: %v\n", err)
		return
	}

	for i := range config.Modules {
		mod := &config.Modules[i]
		if mod.Moniker == "r2r-cli" {
			fmt.Printf("r2r-cli: %+v\n", mod.Versioning)
			break
		}
	}
}
