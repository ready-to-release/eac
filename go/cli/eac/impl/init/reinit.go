package init

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/ready-to-release/eac/go/core/config"
)

// ExistingConfig represents detected existing configuration
type ExistingConfig struct {
	HasRepositoryYML         bool
	HasBooksYML              bool
	HasEnvironmentsYML       bool
	HasAIProviderYML         bool
	HasAIProviderPersonalYML bool

	// Parsed AI provider settings (if present)
	AIProvider string
	AIModel    string

	// Module count from repository.yml
	ModuleCount int
}

// detectExistingConfig checks for existing .eac configuration
func detectExistingConfig(workspaceRoot string) *ExistingConfig {
	eacDir := filepath.Join(workspaceRoot, ".eac")

	// Check if .eac directory exists
	if _, err := os.Stat(eacDir); os.IsNotExist(err) {
		return nil // No existing config
	}

	config := &ExistingConfig{
		HasRepositoryYML:         fileExists(filepath.Join(eacDir, "repository.yml")),
		HasBooksYML:              fileExists(filepath.Join(eacDir, "books.yml")),
		HasEnvironmentsYML:       fileExists(filepath.Join(eacDir, "environments.yml")),
		HasAIProviderYML:         fileExists(filepath.Join(eacDir, "ai-provider.yml")),
		HasAIProviderPersonalYML: fileExists(filepath.Join(eacDir, "ai-provider.personal.yml")),
	}

	// Load AI provider settings (personal takes precedence)
	if config.HasAIProviderPersonalYML {
		personalPath := filepath.Join(eacDir, "ai-provider.personal.yml")
		if aiConfig, err := loadAIConfig(personalPath); err == nil {
			config.AIProvider = aiConfig.AI.Provider
			config.AIModel = aiConfig.AI.Model
		}
	} else if config.HasAIProviderYML {
		teamPath := filepath.Join(eacDir, "ai-provider.yml")
		if aiConfig, err := loadAIConfig(teamPath); err == nil {
			config.AIProvider = aiConfig.AI.Provider
			config.AIModel = aiConfig.AI.Model
		}
	}

	// Load module count from repository.yml
	if config.HasRepositoryYML {
		repoPath := filepath.Join(eacDir, "repository.yml")
		if repoConfig, err := loadRepositoryYML(repoPath); err == nil {
			config.ModuleCount = len(repoConfig.Modules)
		}
	}

	return config
}

// aiConfigFile represents the structure of ai-provider.yml
type aiConfigFile struct {
	AI struct {
		Provider string `yaml:"provider"`
		Model    string `yaml:"model"`
	} `yaml:"ai"`
}

// loadAIConfig loads and parses ai-provider.yml or ai-provider.personal.yml
func loadAIConfig(path string) (*aiConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config aiConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// repositoryYMLFile represents the minimal structure we need from repository.yml
type repositoryYMLFile struct {
	Modules []map[string]interface{} `yaml:"modules"`
}

// loadRepositoryYML loads and parses repository.yml
func loadRepositoryYML(path string) (*repositoryYMLFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config repositoryYMLFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// reinitialize handles re-initialization of an existing EAC project.
// It loads the existing config, re-scans if needed, merges with preservation of user edits,
// and writes the updated configuration.
func reinitialize(deps *Deps, workspaceRoot, eacDir string, scan bool, aiProvider string) int {
	log.Info("🔍 Re-scanning repository structure...")

	// Load existing repository config
	existingRepoPath := filepath.Join(eacDir, "repository.yml")
	existingConfig, err := config.Load(config.LoadOptions{
		RepoRoot:        workspaceRoot,
		ValidateSchemas: false,
	})
	if err != nil {
		log.Warn(fmt.Sprintf("   ⚠️  Could not load existing config: %v", err))
		log.Info("   Falling back to fresh initialization")
		// Continue with fresh init logic below
	}

	var newRepoConfig *config.RepositoryConfig

	if scan {
		// Re-scan repository
		scanResult, err := ScanRepository(workspaceRoot)
		if err != nil {
			log.Error(fmt.Sprintf("Error scanning repository: %v", err))
			return 1
		}

		if len(scanResult.Modules) == 0 {
			log.Warn("   ⚠️  No modules detected")
		} else {
			log.Info(fmt.Sprintf("   ✓ Detected: %d module(s)", len(scanResult.Modules)))
		}

		// Generate new config from scan using strategy pattern
		var repoYAML string
		if aiProvider != "" {
			log.Info(fmt.Sprintf("🤖 Regenerating configuration with AI (%s)...", aiProvider))
			// Use AI generator with rule-based fallback
			aiGen := newAIGenerator(aiProvider, deps)
			ruleGen := NewRuleBasedGenerator()

			// Try AI generation first
			repoYAML, err = aiGen.Generate(workspaceRoot, scanResult)
			if err != nil {
				log.Warn(fmt.Sprintf("   ⚠️  AI generation failed: %v", err))
				log.Info("   Falling back to rule-based generation")
				repoYAML, err = ruleGen.Generate(workspaceRoot, scanResult)
				if err != nil {
					log.Error(fmt.Sprintf("Error generating config: %v", err))
					return 1
				}
			}
		} else {
			log.Info("🔧 Regenerating configuration (rule-based)...")
			generator := NewRuleBasedGenerator()
			repoYAML, err = generator.Generate(workspaceRoot, scanResult)
			if err != nil {
				log.Error(fmt.Sprintf("Error generating config: %v", err))
				return 1
			}
		}

		// Parse the YAML into config struct
		if err := yaml.Unmarshal([]byte(repoYAML), &newRepoConfig); err != nil {
			log.Error(fmt.Sprintf("Error parsing generated config: %v", err))
			return 1
		}
	} else {
		// Load current defaults (no scan)
		log.Info("🔧 Updating configuration with current defaults...")
		newConfig, err := config.Load(config.LoadOptions{
			RepoRoot:        workspaceRoot,
			ValidateSchemas: false,
		})
		if err != nil {
			log.Error(fmt.Sprintf("Error loading defaults: %v", err))
			return 1
		}
		newRepoConfig = newConfig.Repository
	}

	// Merge configs if we have existing config
	var result *MergeResult
	if existingConfig != nil && existingConfig.Repository != nil {
		log.Info("")
		log.Info("🔀 Merging with existing configuration...")
		newRepoConfig, result = mergeConfigs(workspaceRoot, existingConfig.Repository, newRepoConfig)
		log.Info("   ✓ Configuration merged")
	} else {
		// No existing config, treat as new
		result = &MergeResult{
			ModulesAdded: make([]string, len(newRepoConfig.Modules)),
		}
		for i, mod := range newRepoConfig.Modules {
			result.ModulesAdded[i] = mod.Moniker
		}
	}

	// Write updated repository.yml
	repoYAMLBytes, err := yaml.Marshal(newRepoConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Error serializing config: %v", err))
		return 1
	}

	header := `# EAC Repository Configuration
# Updated by: clie eac init
#
# This file defines your repository settings and module domain.
# Edit this file to customize modules, dependencies, and build settings.
#
# Documentation: https://eac.readthedocs.io/configuration/repository

`

	if err := os.WriteFile(existingRepoPath, []byte(header+string(repoYAMLBytes)), 0o644); err != nil {
		log.Error(fmt.Sprintf("Error writing repository.yml: %v", err))
		return 1
	}
	log.Info(fmt.Sprintf("   ✓ Updated %s", existingRepoPath))

	// Show merge result summary
	showMergeResult(result)

	// Handle AI provider configuration if specified AND valid
	if aiProvider != "" {
		// Validate provider before trying to configure
		validProviders := []string{"claude-api", "openai", "gemini"}
		isValid := false
		for _, vp := range validProviders {
			if aiProvider == vp {
				isValid = true
				break
			}
		}

		if isValid {
			log.Info("")
			log.Info("🤖 Updating AI Provider Configuration")
			log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			// Configure the AI provider
			agentCfg, err := configureAgent(aiProvider)
			if err != nil {
				log.Error(fmt.Sprintf("Error configuring AI provider: %v", err))
				return 1
			}

			// Write AI provider config (always team config during re-init)
			configPath, err := writeConfig(workspaceRoot, agentCfg, &tokenConfig{})
			if err != nil {
				log.Error(fmt.Sprintf("Error writing AI provider config: %v", err))
				return 1
			}

			log.Info(fmt.Sprintf("   ✓ Updated AI provider configuration: %s", configPath))
		}
		// If invalid provider, silently skip (might be a test artifact)
	}

	// Success
	log.Info("")
	log.Info("✅ EAC project re-initialized")
	log.Info("")
	log.Info("📋 Next steps:")
	log.Info("   1. Review changes: git diff .eac/repository.yml")
	log.Info("   2. Verify modules: clie eac show modules")
	log.Info("   3. Commit updates: git add .eac/ && git commit -m \"Update EAC config\"")
	log.Info("")

	return 0
}
