package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/cache"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/github"
	"github.com/spf13/cobra"
)

var (
	listShowTags   bool
	listRefresh    bool
	listClearCache bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available extensions from the registry",
	Long: `List all available extensions from the GitHub Container Registry.
	
Dynamically discovers extensions by querying the registry.
Shows extension names, latest versions, and installation status.
Can optionally show all available tags for each extension.`,
	Example: `  # List all available extensions
  r2r list
  
  # List with all available tags for each extension
  r2r list --tags
  
  # Refresh cache and list
  r2r list --refresh

  # Clear cache without listing
  r2r list --clear-cache`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle clear cache flag - just clear cache and exit
		if listClearCache {
			registryCache, _ := cache.Load()
			if registryCache != nil {
				registryCache.Clear()
				err := registryCache.Save()
				if err != nil {
					fmt.Printf("❌ Failed to clear cache: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("✅ Cache cleared successfully")
			} else {
				fmt.Println("ℹ️  No cache found to clear")
			}
			return
		}

		// Discover available extensions from registry
		log.Debug().Msg("Discovering available extensions from registry...")

		// Try to get extensions from cache first
		registryCache, _ := cache.Load()
		var knownExtensions map[string]string

		// Check if we need to discover extensions
		if registryCache == nil || len(registryCache.Extensions) == 0 || listRefresh {
			// Try to query registry if we have credentials
			if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_USERNAME") != "" {
				// We have credentials, try to discover extensions
				client, err := github.NewRegistryClient()
				if err == nil {
					extensions, err := client.ListExtensions()
					if err == nil && len(extensions) > 0 {
						// Build the knownExtensions map from discovered extensions
						knownExtensions = make(map[string]string)
						for _, ext := range extensions {
							knownExtensions[ext.Name] = ext.ImagePath
							log.Debug().Str("extension", ext.Name).Str("path", ext.ImagePath).Msg("Discovered extension")
						}
					}
				}
			}

			// If we don't have extensions yet (no auth or API failed), error
			if len(knownExtensions) == 0 {
				fmt.Println("❌ No extensions discovered. Set GITHUB_TOKEN and GITHUB_USERNAME environment variables.")
				fmt.Println("\nTo set credentials:")
				fmt.Println("  export GITHUB_TOKEN=your_github_token")
				fmt.Println("  export GITHUB_USERNAME=your_username")
				os.Exit(1)
			}
		} else {
			// Use cached extension list
			knownExtensions = make(map[string]string)
			for name := range registryCache.Extensions {
				// Reconstruct image path from cache
				knownExtensions[name] = fmt.Sprintf("ghcr.io/ready-to-release/r2r-cli/extensions/%s", name)
			}
		}

		// Optionally refresh cache
		if listRefresh {
			// Clear cache to force refresh
			if registryCache != nil {
				registryCache.Clear()
				registryCache.Save()
			}
			fmt.Println("ℹ️  Cache cleared, fetching fresh data from registry...")
		}

		// Load configuration to get currently installed extensions
		conf.InitConfig()

		// Build a map of configured extensions for status checking
		configuredExtensions := make(map[string]string)
		for _, ext := range conf.Global.Extensions {
			configuredExtensions[ext.Name] = ext.Image
		}

		// Get latest versions only if we have credentials
		latestVersions := make(map[string]string)
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_USERNAME") != "" {
			// Create temporary config with all known extensions to fetch their data
			tempConfig := &conf.Config{
				Extensions: []conf.Extension{},
			}
			for name, baseImage := range knownExtensions {
				tempConfig.Extensions = append(tempConfig.Extensions, conf.Extension{
					Name:  name,
					Image: baseImage + ":latest", // Will be used to fetch actual versions
				})
			}

			// Use ValidatePinnedExtensions to populate cache with latest versions
			// This will fetch from registry if needed
			unpinnedMessages, _ := conf.ValidatePinnedExtensions(tempConfig, false)

			// Parse the messages to extract the SHA tags
			for _, msg := range unpinnedMessages {
				// Message format: "'name' must be pinned, latest is: sha-xxxxx"
				if parts := strings.Split(msg, "'"); len(parts) >= 2 {
					name := parts[1]
					if idx := strings.Index(msg, "latest is: "); idx > 0 {
						sha := strings.TrimSpace(msg[idx+11:])
						latestVersions[name] = sha
					}
				}
			}
		} else {
			// Without credentials, we can't get latest versions
			for name := range knownExtensions {
				latestVersions[name] = "unknown"
			}
		}

		// Reload cache to get all tags
		registryCache, _ = cache.Load()

		// Create a tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Print header
		fmt.Fprintln(w, "EXTENSION\tLATEST VERSION\tSTATUS\tCONFIGURED VERSION")
		fmt.Fprintln(w, "─────────\t──────────────\t──────\t──────────────────")

		// List each available extension
		for name := range knownExtensions {
			// Get latest version
			latestVersion := latestVersions[name]
			if latestVersion == "" || latestVersion == "sha-<unavailable>" {
				latestVersion = "sha-<pending>"
			}

			// Check if configured
			status := "Not installed"
			configuredVersion := "-"
			if configuredImage, ok := configuredExtensions[name]; ok {
				// Extract tag from configured image
				if idx := strings.LastIndex(configuredImage, ":"); idx > 0 {
					configuredVersion = configuredImage[idx+1:]
				} else {
					configuredVersion = "latest"
				}

				// Determine status
				if configuredVersion == latestVersion {
					status = "✅ Up to date"
				} else if strings.HasPrefix(configuredVersion, "sha-") {
					status = "📌 Pinned"
				} else if configuredVersion == "latest" || configuredVersion == "main" || configuredVersion == "master" {
					status = "⚠️  Unpinned"
				} else {
					status = "📦 Custom"
				}
			}

			// Print basic info
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				name,
				latestVersion,
				status,
				configuredVersion,
			)

			// Optionally show all available tags
			if listShowTags && registryCache != nil {
				if extCache, ok := registryCache.GetExtension(name); ok && len(extCache.Tags) > 0 {
					// Group tags by type
					var shaTags, runTags, otherTags []string
					for _, tag := range extCache.Tags {
						if strings.HasPrefix(tag, "sha-") {
							shaTags = append(shaTags, tag)
						} else if strings.HasPrefix(tag, "run-") {
							runTags = append(runTags, tag)
						} else {
							otherTags = append(otherTags, tag)
						}
					}

					// Show tags grouped by type
					if len(shaTags) > 0 {
						fmt.Fprintf(w, "\t  SHA tags:\t%s\n", strings.Join(limitTags(shaTags, 3), ", "))
					}
					if len(runTags) > 0 {
						fmt.Fprintf(w, "\t  Run tags:\t%s\n", strings.Join(limitTags(runTags, 3), ", "))
					}
					if len(otherTags) > 0 {
						fmt.Fprintf(w, "\t  Other tags:\t%s\n", strings.Join(limitTags(otherTags, 5), ", "))
					}
				}
			}
		}

		// Flush the table
		w.Flush()

		// Show installation help if any extensions are not installed
		hasUninstalled := false
		for name := range knownExtensions {
			if _, ok := configuredExtensions[name]; !ok {
				hasUninstalled = true
				break
			}
		}

		if hasUninstalled {
			fmt.Println("\n📦 To add and install extensions:")
			fmt.Println("  r2r install <name>")
		}

		// Show tips
		fmt.Println("\n💡 Tips:")
		fmt.Println("  • Use 'r2r list --tags' to see all available tags")
		fmt.Println("  • Use 'r2r list --refresh' to update cache from registry")
		fmt.Println("  • Use 'r2r list --clear-cache' to clear cache without listing")
		fmt.Println("  • Use 'r2r install <name>' to add extensions with latest SHA")
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&listShowTags, "tags", "t", false, "Show all available tags for each extension")
	listCmd.Flags().BoolVarP(&listRefresh, "refresh", "r", false, "Refresh cache from registry")
	listCmd.Flags().BoolVar(&listClearCache, "clear-cache", false, "Clear the registry cache and exit")
}

// limitTags returns at most n tags, with ellipsis if there are more
func limitTags(tags []string, n int) []string {
	if len(tags) <= n {
		return tags
	}
	result := tags[:n]
	result = append(result, fmt.Sprintf("... (%d more)", len(tags)-n))
	return result
}

// formatDuration formats a time as a human-readable duration
func formatDuration(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}

// formatRemainingTime formats a duration as human-readable remaining time
func formatRemainingTime(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
