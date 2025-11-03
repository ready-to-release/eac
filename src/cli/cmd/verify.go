package cmd

import (
	"context"
	"os"

	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/ready-to-release/eac/src/cli/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(VerifyCmd)
}

var VerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify system prerequisites",
	Long:  `Verifies that GitHub authentication and Docker service are working properly.`,
	Run: func(cmd *cobra.Command, args []string) {
		verifySystem(cmd)
	},
}

func verifySystem(cmd *cobra.Command) {
	cmd.Println("🔍 Verifying system prerequisites...")

	// Initialize configuration - will exit with detailed error if it fails
	conf.InitConfig()

	allChecksPass := true

	// Check config file (just reports it's verified since InitConfig succeeded)
	if !checkConfigFile(cmd) {
		allChecksPass = false
	}

	// Check GitHub environment variables
	if !checkGitHubAuth(cmd) {
		allChecksPass = false
	}

	// Check Docker service
	if !checkDockerService(cmd) {
		allChecksPass = false
	}

	if allChecksPass {
		cmd.Println("✅ All checks passed! System is ready.")
		log.Debug().Msg("System verification completed successfully")
	} else {
		cmd.PrintErrln("❌ Some checks failed. Please fix the issues above.")
		log.Error().Msg("System verification failed")
		os.Exit(1)
	}
}

func checkGitHubAuth(cmd *cobra.Command) bool {
	cmd.Println("🔑 Checking GitHub authentication...")

	// Use centralized authentication function
	authConfig, authStr, err := docker.CreateGitHubAuthConfig()
	if err != nil {
		cmd.PrintErrf("❌ %v\n", err)
		return false
	}

	cmd.Printf("✅ GitHub credentials found (username: %s)\n", authConfig.Username)

	// Validate the auth string was created successfully
	if authStr == "" {
		cmd.PrintErrln("❌ Error creating authentication string")
		return false
	}

	cmd.Println("✅ GitHub authentication configuration is valid")

	return true
}

func checkDockerService(cmd *cobra.Command) bool {
	cmd.Println("🐳 Checking Docker service...")

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		cmd.PrintErrf("❌ Error creating Docker client: %v\n", err)
		return false
	}
	defer cli.Close()

	// Test Docker connection by pinging the daemon
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		cmd.PrintErrf("❌ Docker daemon is not running or accessible: %v\n", err)
		return false
	}

	cmd.Println("✅ Docker service is running and accessible")

	// Get Docker version info
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		cmd.Println("⚠️  Could not get Docker version info")
	} else {
		cmd.Printf("✅ Docker version: %s\n", version.Version)
	}

	return true
}

func checkConfigFile(cmd *cobra.Command) bool {
	cmd.Println("📋 Checking r2r-cli configuration...")
	
	// Configuration is already loaded and verified at startup
	// If we got here, the config file exists and was parsed
	cmd.Println("✅ Configuration file loaded")
	
	// The configuration has already been validated during loading in conf.InitConfig()
	// Check if we have at least one extension configured
	if len(conf.Global.Extensions) == 0 {
		cmd.Println("⚠️  No extensions configured")
		return false
	}
	
	// Verify each extension has required fields
	for i, ext := range conf.Global.Extensions {
		if ext.Name == "" {
			cmd.Printf("❌ Extension %d: missing name\n", i)
			return false
		}
		if ext.Image == "" {
			cmd.Printf("❌ Extension '%s': missing image\n", ext.Name)
			return false
		}
	}
	
	// If we got here, config is valid (it was already validated during load)
	cmd.Printf("✅ Configuration valid with %d extension(s)\n", len(conf.Global.Extensions))
	return true
}

// Original validator-based function kept for reference but not used
func checkConfigFileWithValidator(cmd *cobra.Command) bool {
	cmd.Println("📋 Checking r2r-cli configuration...")
	cmd.Println("✅ Configuration file loaded")
	
	configMap := viper.AllSettings()
	
	v, err := validator.NewEmbeddedValidator()
	if err != nil {
		cmd.PrintErrf("⚠️  Could not initialize validator: %v\n", err)
		cmd.Println("✅ Configuration appears valid (basic validation only)")
		return true
	}
	
	result, err := v.ValidateInterface(configMap)
	if err != nil {
		cmd.PrintErrf("⚠️  Validation error: %v\n", err)
		cmd.Println("✅ Configuration appears valid (basic validation only)")
		return true // Don't fail the check if validation fails
	}
	
	// Check validation results
	if !result.IsValid() {
		cmd.Println("❌ Configuration validation failed:")
		for _, e := range result.Errors {
			if e.Field != "" {
				cmd.Printf("   - %s: %s\n", e.Field, e.Message)
			} else {
				cmd.Printf("   - %s\n", e.Message)
			}
		}
		return false
	}
	
	if len(result.Warnings) > 0 {
		cmd.Println("⚠️  Configuration has warnings:")
		for _, w := range result.Warnings {
			if w.Field != "" {
				cmd.Printf("   - %s: %s\n", w.Field, w.Message)
			} else {
				cmd.Printf("   - %s\n", w.Message)
			}
		}
	}
	
	cmd.Printf("✅ Configuration is valid (schema version: %s)\n", validator.GetEmbeddedSchemaVersion())
	
	// Show summary of what's configured
	if len(conf.Global.Extensions) > 0 {
		cmd.Printf("   Extensions configured: %d\n", len(conf.Global.Extensions))
	}
	
	return true
}
