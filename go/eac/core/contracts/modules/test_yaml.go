package modules

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

func TestYAML() {
	workspaceRoot, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}
	
	yamlPath := filepath.Join(workspaceRoot, ".r2r/eac/repository.yml")
	data, _ := os.ReadFile(yamlPath)
	
	var config struct {
		Modules []contracts.BaseContract `yaml:"modules"`
	}
	
	yaml.Unmarshal(data, &config)
	
	for _, mod := range config.Modules {
		if mod.Moniker == "r2r-cli" {
			fmt.Printf("r2r-cli: %+v\n", mod.Versioning)
			break
		}
	}
}
