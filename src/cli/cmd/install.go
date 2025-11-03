package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/extensions"
	"github.com/ready-to-release/eac/src/cli/internal/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	RootCmd.AddCommand(InstallCmd)

	// Add --load-local flag to control local image usage
	InstallCmd.Flags().Bool("load-local", false, "Use local development images instead of pulling from registry")
}

var InstallCmd = &cobra.Command{
	Use:   "install [extension-name]",
	Short: "Install configured extensions or add and install new ones",
	Long: `Install extensions by pulling their Docker images.

When no extension name is provided, installs all configured extensions.
When an extension name is provided, adds it to the configuration with the latest SHA tag and installs it.

Examples:
  # Install all configured extensions
  r2r install
  
  # Add and install a specific extension
  r2r install pwsh
  r2r install python
  r2r install go
  
  # Install with local development images
  r2r install pwsh --load-local`,
	Run: func(cmd *cobra.Command, args []string) {
		// If extension name provided, add it to config (creates config if needed)
		if len(args) > 0 {
			extensionName := args[0]

			// Add extension to config (will create config file if needed)
			if err := addExtensionToConfig(extensionName); err != nil {
				log.Error().Err(err).Msg("Failed to add extension to config")
				os.Exit(1)
			}
			fmt.Printf("✅ Added %s to configuration\n", extensionName)
		} else {
			// No extension name provided - need to check if config exists
			repoRoot, err := conf.FindRepositoryRoot()
			if err != nil {
				log.Error().Err(err).Msg("Failed to find repository root")
				os.Exit(1)
			}

			configPath := filepath.Join(repoRoot, "r2r-cli.yml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Println("❌ No configuration file found.")
				fmt.Printf("To install all configured extensions, you need a configuration file at: %s\n", configPath)
				fmt.Println("\nTo get started:")
				fmt.Println("  • Run 'r2r init' to create a configuration file")
				fmt.Println("  • Or install a specific extension: 'r2r install <extension-name>'")
				fmt.Println("\nExamples:")
				fmt.Println("  r2r install pwsh")
				fmt.Println("  r2r install python")
				os.Exit(1)
			}
		}

		// Load configuration
		conf.InitConfig()

		// Check for --load-local flag and temporarily override global setting
		loadLocal, _ := cmd.Flags().GetBool("load-local")
		var originalLoadLocal bool
		if loadLocal {
			originalLoadLocal = conf.Global.LoadLocal
			conf.Global.LoadLocal = true
			log.Debug().Bool("load_local", true).Msg("Temporarily overriding load_local setting from --load-local flag")
		}
		defer func() {
			if loadLocal {
				conf.Global.LoadLocal = originalLoadLocal
				log.Debug().Bool("load_local", originalLoadLocal).Msg("Restored original load_local setting")
			}
		}()

		// Create extension installer
		installer, err := extensions.NewInstaller()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create extension installer")
			os.Exit(1)
		}
		defer installer.Close()

		// Determine which extensions to install
		var extsToInstall []conf.Extension
		if len(args) > 0 {
			// Install only the specified extension
			extensionName := args[0]

			// First try to find in existing configuration
			found := false
			for _, ext := range conf.Global.Extensions {
				if ext.Name == extensionName {
					extsToInstall = append(extsToInstall, ext)
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("❌ Extension %s not found in configuration\n", extensionName)
				os.Exit(1)
			}
		} else {
			// Install all configured extensions
			extsToInstall = conf.Global.Extensions
			if len(extsToInstall) == 0 {
				fmt.Println("❌ No extensions configured. Add an extension with:")
				fmt.Println("  r2r install <extension-name>")
				fmt.Println("\nExamples:")
				fmt.Println("  r2r install pwsh")
				fmt.Println("  r2r install python")
				os.Exit(1)
			}
		}

		// Install the extensions
		fmt.Printf("📦 Installing %d extension(s)...\n", len(extsToInstall))

		successCount := 0
		for _, ext := range extsToInstall {
			fmt.Printf("\n🔧 Installing %s...\n", ext.Name)

			pulled, err := installer.EnsureExtensionImage(ext.Name)
			if err != nil {
				log.Error().Err(err).Str("extension", ext.Name).Msg("Failed to install extension")
				fmt.Printf("❌ Failed to install %s: %v\n", ext.Name, err)
			} else {
				if pulled {
					fmt.Printf("✅ %s installed (new image pulled)\n", ext.Name)
				} else {
					fmt.Printf("✅ %s already up to date\n", ext.Name)
				}
				successCount++
			}
		}

		if successCount == len(extsToInstall) {
			fmt.Println("\n✅ All extensions installed successfully")
		} else {
			fmt.Printf("\n⚠️  %d of %d extensions installed successfully\n", successCount, len(extsToInstall))
			os.Exit(1)
		}
	},
}

// addExtensionToConfig adds an extension to the config file with the latest SHA version
func addExtensionToConfig(extensionName string) error {
	if extensionName == "" {
		return fmt.Errorf("extension name is required")
	}

	// Find the config file
	repoRoot, err := conf.FindRepositoryRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	configPath := filepath.Join(repoRoot, "r2r-cli.yml")

	// Check if config file exists
	var configMap map[string]interface{}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, create a minimal one
			fmt.Printf("📝 Creating r2r-cli.yml in repository root\n")
			configMap = map[string]interface{}{
				"version":    "1.0",
				"extensions": []interface{}{},
			}
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Parse existing config
		if err := yaml.Unmarshal(configData, &configMap); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Ensure version is set
	if configMap["version"] == nil {
		configMap["version"] = "1.0"
	}

	// Get or create extensions list
	var extensions []interface{}
	if exts, ok := configMap["extensions"].([]interface{}); ok {
		extensions = exts
	} else {
		extensions = []interface{}{}
	}

	// Check if extension already exists in config
	found := false
	for _, ext := range extensions {
		if extMap, ok := ext.(map[string]interface{}); ok {
			if name, ok := extMap["name"].(string); ok && name == extensionName {
				found = true
				break
			}
		}
	}

	if !found {
		// Need to discover from registry
		if os.Getenv("GITHUB_TOKEN") == "" || os.Getenv("GITHUB_USERNAME") == "" {
			return fmt.Errorf("GITHUB_TOKEN and GITHUB_USERNAME required to discover extension: %s", extensionName)
		}

		// Query registry to verify extension exists
		client, err := github.NewRegistryClient()
		if err != nil {
			return fmt.Errorf("failed to create registry client: %w", err)
		}

		// Try to list tags for this specific extension
		imagePath := fmt.Sprintf("ready-to-release/r2r-cli/extensions/%s", extensionName)
		tags, err := client.ListTags(imagePath)
		if err != nil || len(tags) == 0 {
			return fmt.Errorf("extension not found in registry: %s", extensionName)
		}

		// Extension exists, add it
		extensions = append(extensions, map[string]interface{}{
			"name":        extensionName,
			"description": fmt.Sprintf("%s development environment", strings.Title(extensionName)),
			"image":       fmt.Sprintf("ghcr.io/%s:latest", imagePath), // Will be replaced with SHA
		})
	}

	// Update the extension with the latest SHA tag
	fmt.Printf("📌 Getting latest SHA tag for %s...\n", extensionName)

	// Create a minimal config to use the existing logic from conf package
	tempConfig := &conf.Config{
		Extensions: []conf.Extension{},
	}

	// Convert our extensions to conf.Extension type for processing
	for _, ext := range extensions {
		extMap, ok := ext.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := extMap["name"].(string)
		image, _ := extMap["image"].(string)

		if name != "" && image != "" {
			tempConfig.Extensions = append(tempConfig.Extensions, conf.Extension{
				Name:  name,
				Image: image,
			})
		}
	}

	// Use validatePinnedExtensions to get the latest SHA tags
	// This function handles cache loading and registry fetching internally
	unpinnedMessages, _ := conf.ValidatePinnedExtensions(tempConfig, false)

	// Parse the messages to extract the SHA tags
	shaMap := make(map[string]string)
	for _, msg := range unpinnedMessages {
		// Message format: "'name' must be pinned, latest is: sha-xxxxx"
		if parts := strings.Split(msg, "'"); len(parts) >= 2 {
			name := parts[1]
			if idx := strings.Index(msg, "latest is: "); idx > 0 {
				sha := strings.TrimSpace(msg[idx+11:])
				shaMap[name] = sha
			}
		}
	}

	// Find and update only the specific extension we're adding
	updated := false
	for i, ext := range extensions {
		extMap, ok := ext.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := extMap["name"].(string)
		image, _ := extMap["image"].(string)

		// Only update the extension we're adding
		if name != extensionName {
			continue
		}

		if image == "" {
			continue
		}

		// Skip if already pinned (has sha- tag)
		if strings.Contains(image, ":sha-") {
			fmt.Printf("✅ %s already pinned\n", name)
			return nil
		}

		// Extract base image
		baseImage := image
		if idx := strings.LastIndex(image, ":"); idx > 0 {
			baseImage = image[:idx]
		}

		// Get the SHA from our map
		latestSHA, ok := shaMap[name]
		if !ok || latestSHA == "" || latestSHA == "sha-<unavailable>" {
			return fmt.Errorf("failed to get latest SHA for %s", name)
		}

		// Update the extension
		extMap["image"] = baseImage + ":" + latestSHA
		extMap["image_pull_policy"] = "IfNotPresent"
		extensions[i] = extMap
		updated = true

		fmt.Printf("📌 %s configured with %s\n", name, latestSHA)
		break
	}

	if !updated {
		return fmt.Errorf("failed to update extension %s", extensionName)
	}

	// Create YAML content with proper field ordering
	// Using a custom structure to ensure "version" comes first
	type OrderedConfig struct {
		Version    string        `yaml:"version"`
		Extensions []interface{} `yaml:"extensions"`
	}

	orderedConfig := OrderedConfig{
		Version:    configMap["version"].(string),
		Extensions: extensions,
	}

	// Marshal back to YAML
	updatedConfig, err := yaml.Marshal(&orderedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(configPath, updatedConfig, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Configuration updated in %s\n", configPath)

	return nil
}
