package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/ready-to-release/eac/go/cli/clie/internal/extensions"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"github.com/spf13/cobra"
)

// Note: fmt is still imported for fmt.Sprintf and fmt.Errorf

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
  clie install
  
  # Add and install a specific extension
  clie install eac
  clie install python
  clie install go
  
  # Install with local development images
  clie install eac --load-local`,
	Run: func(cmd *cobra.Command, args []string) {
		// If extension name provided, add it to config (creates config if needed)
		if len(args) > 0 {
			extensionName := args[0]

			// Add extension to config (will create config file if needed)
			if err := addExtensionToConfig(extensionName); err != nil {
				logging.Errorf("Failed to add extension to config: %v", err)
				os.Exit(1)
			}
			logging.Infof("✅ Added %s to configuration", extensionName)
		} else {
			// No extension name provided - need to check if config exists
			repoRoot, err := conf.FindRepositoryRoot()
			if err != nil {
				logging.Errorf("Failed to find repository root: %v", err)
				os.Exit(1)
			}

			// Check for config file at .clie/clie.yml
			configPaths := []string{
				filepath.Join(repoRoot, ".clie", "clie.yml"),
				filepath.Join(repoRoot, ".clie", "clie.yaml"),
			}
			configFound := false
			for _, cp := range configPaths {
				if _, err := os.Stat(cp); err == nil {
					configFound = true
					break
				}
			}
			if !configFound {
				configPath := filepath.Join(repoRoot, ".clie", "clie.yml")
				logging.Error("❌ No configuration file found.")
				logging.Infof("To install all configured extensions, you need a configuration file at: %s", configPath)
				logging.Info("\nTo get started:")
				logging.Info("  • Run 'clie init' to create a configuration file")
				logging.Info("  • Or install a specific extension: 'clie install <extension-name>'")
				logging.Info("\nExamples:")
				logging.Info("  clie install eac")
				logging.Info("  clie install python")
				os.Exit(1)
			}
		}

		// Load configuration
		conf.InitConfig()

		// Check for --load-local flag and temporarily override global setting
		loadLocal, _ := cmd.Flags().GetBool("load-local") //nolint:errcheck // flag registered in init
		var originalLoadLocal bool
		if loadLocal {
			originalLoadLocal = conf.Global.LoadLocal
			conf.Global.LoadLocal = true
			logging.Debugf("Temporarily overriding load_local setting from --load-local flag: load_local=%v", true)
		}
		defer func() {
			if loadLocal {
				conf.Global.LoadLocal = originalLoadLocal
				logging.Debugf("Restored original load_local setting: load_local=%v", originalLoadLocal)
			}
		}()

		// Create extension installer
		installer, err := extensions.NewInstaller()
		if err != nil {
			logging.Errorf("Failed to create extension installer: %v", err)
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
				logging.Errorf("❌ Extension %s not found in configuration", extensionName)
				os.Exit(1)
			}
		} else {
			// Install all configured extensions
			extsToInstall = conf.Global.Extensions
			if len(extsToInstall) == 0 {
				logging.Error("❌ No extensions configured. Add an extension with:")
				logging.Info("  clie install <extension-name>")
				logging.Info("\nExamples:")
				logging.Info("  clie install eac")
				logging.Info("  clie install python")
				os.Exit(1)
			}
		}

		// Install the extensions
		logging.Infof("📦 Installing %d extension(s)...", len(extsToInstall))

		successCount := 0
		for _, ext := range extsToInstall {
			logging.Infof("\n🔧 Installing %s...", ext.Name)

			pulled, err := installer.EnsureExtensionImage(ext.Name)
			if err != nil {
				logging.Errorf("Failed to install extension: extension=%s error=%v", ext.Name, err)
				logging.Errorf("❌ Failed to install %s: %v", ext.Name, err)
			} else {
				if pulled {
					logging.Infof("✅ %s installed (new image pulled)", ext.Name)
				} else {
					logging.Infof("✅ %s already up to date", ext.Name)
				}
				successCount++
			}
		}

		if successCount == len(extsToInstall) {
			logging.Info("\n✅ All extensions installed successfully")
		} else {
			logging.Warnf("\n⚠️  %d of %d extensions installed successfully", successCount, len(extsToInstall))
			os.Exit(1)
		}
	},
}

// addExtensionToConfig adds an extension to the config file with the latest SHA version.
// This orchestrates config file operations using focused helper functions in install_config.go.
func addExtensionToConfig(extensionName string) error {
	if extensionName == "" {
		return fmt.Errorf("extension name is required")
	}

	repoRoot, err := conf.FindRepositoryRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find or create config file
	configPath, configMap, err := findOrCreateConfigFile(repoRoot)
	if err != nil {
		return err
	}

	// Get extensions list and check for duplicates
	extensions := getExtensionsList(configMap)

	if !extensionExistsInConfig(extensions, extensionName) {
		extensions, err = verifyAndAddExtension(extensions, extensionName)
		if err != nil {
			return err
		}
	}

	// Resolve SHA tags
	shaMap, err := resolveLatestSHA(extensions, extensionName)
	if err != nil {
		return err
	}

	// Update the extension with SHA
	extensions, err = updateExtensionWithSHA(extensions, extensionName, shaMap)
	if err != nil {
		return err
	}

	// Write updated config
	return writeConfigFile(configPath, configMap, extensions)
}
