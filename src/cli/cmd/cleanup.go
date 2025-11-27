package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/spf13/cobra"
)

var (
	cleanupAll        bool
	cleanupDryRun     bool
	keepVersions      int
	cleanupContainers bool
	cleanupExtensions bool
	cleanupOlderThan  string
)

func init() {
	RootCmd.AddCommand(CleanupCmd)
	CleanupCmd.Flags().BoolVarP(&cleanupAll, "all", "a", false, "Remove all unused images, not just r2r-cli extensions")
	CleanupCmd.Flags().BoolVarP(&cleanupDryRun, "dry-run", "n", false, "Show what would be removed without actually removing")
	CleanupCmd.Flags().IntVarP(&keepVersions, "keep", "k", 1, "Number of versions to keep per extension (default: 1)")
	CleanupCmd.Flags().BoolVarP(&cleanupContainers, "containers", "c", false, "Also clean up stopped containers")
	CleanupCmd.Flags().BoolVarP(&cleanupExtensions, "extensions-only", "e", false, "Only clean up extension containers (use with --containers)")
	CleanupCmd.Flags().StringVarP(&cleanupOlderThan, "older-than", "o", "", "Only clean containers older than duration (e.g., '24h', '7d')")
}

var CleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old extension images and containers to free disk space",
	Long: `Remove old versions of extension Docker images and stopped containers to reclaim disk space.

By default, keeps only the most recent version of each extension image.
Use --keep to retain more versions, or --all to also clean non-extension images.
Use --containers to also clean up stopped containers.`,
	Example: `  # Remove all but the latest version of each extension
  r2r cleanup

  # Keep 3 versions of each extension
  r2r cleanup --keep 3

  # Also clean up stopped containers
  r2r cleanup --containers

  # Clean only extension containers
  r2r cleanup --containers --extensions-only

  # Clean containers older than 24 hours
  r2r cleanup --containers --older-than 24h

  # Show what would be removed without actually removing
  r2r cleanup --dry-run

  # Clean all Docker images (not just extensions)
  r2r cleanup --all`,
	Run: func(cmd *cobra.Command, args []string) {
		conf.InitConfig()

		// Handle container cleanup if requested
		if cleanupContainers {
			cleanupDockerContainers()
			fmt.Println() // Add spacing between sections
		}

		// Handle image cleanup
		if cleanupAll {
			cleanAllDockerImages()
		} else {
			cleanExtensionImages()
		}
	},
}

func cleanExtensionImages() {
	fmt.Println("🧹 Cleaning up old extension images...")

	// Get list of configured extensions
	extensions := conf.Global.Extensions
	if len(extensions) == 0 {
		fmt.Println("No extensions configured")
		return
	}

	for _, ext := range extensions {
		// Extract base image name without tag
		baseImage := ext.Image
		if idx := strings.LastIndex(baseImage, ":"); idx > 0 {
			baseImage = baseImage[:idx]
		}

		// Get all versions of this image
		cmd := exec.Command("docker", "images", baseImage, "--format", "{{.Tag}}|{{.ID}}|{{.Size}}")
		output, err := cmd.Output()
		if err != nil {
			continue // Skip if can't list images
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) <= keepVersions {
			fmt.Printf("  %s: %d version(s) found, keeping all\n", ext.Name, len(lines))
			continue
		}

		// Parse and sort by tag (newest first)
		type imageInfo struct {
			tag  string
			id   string
			size string
		}

		var images []imageInfo
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) == 3 {
				images = append(images, imageInfo{
					tag:  parts[0],
					id:   parts[1],
					size: parts[2],
				})
			}
		}

		// Skip if we have the configured number or fewer
		if len(images) <= keepVersions {
			continue
		}

		// Remove old versions (keep the first N)
		toRemove := images[keepVersions:]
		fmt.Printf("  %s: Removing %d old version(s)\n", ext.Name, len(toRemove))

		for i, img := range toRemove {
			imageRef := fmt.Sprintf("%s:%s", baseImage, img.tag)

			if cleanupDryRun {
				fmt.Printf("    [DRY RUN] Would remove: %s (%s)\n", imageRef, img.size)
			} else {
				fmt.Printf("    [%d/%d] Removing: %s (%s)...\n", i+1, len(toRemove), imageRef, img.size)
				cmd := exec.Command("docker", "rmi", imageRef)
				if err := cmd.Run(); err != nil {
					// Try removing by ID if tag removal fails
					cmd = exec.Command("docker", "rmi", img.id)
					cmd.Run() // Best effort
				}
			}
		}
	}

	// Run docker system prune to clean up dangling images
	if !cleanupDryRun {
		fmt.Println("\n🔧 Cleaning up dangling images and build cache...")
		fmt.Println("   ⏳ Removing dangling images (this may take a moment)...")
		cmd := exec.Command("docker", "image", "prune", "-f")
		if output, err := cmd.Output(); err == nil {
			fmt.Print(string(output))
		}

		// Also prune build cache
		fmt.Println("   ⏳ Pruning build cache (this may take several seconds)...")
		cmd = exec.Command("docker", "builder", "prune", "-f")
		if output, err := cmd.Output(); err == nil {
			fmt.Print(string(output))
		}
	}

	// Show disk usage after cleanup
	fmt.Println("\n📊 Calculating Docker disk usage (please wait)...")
	cmd := exec.Command("docker", "system", "df")
	if output, err := cmd.Output(); err == nil {
		fmt.Print(string(output))
	}
}

func cleanAllDockerImages() {
	fmt.Println("🧹 Cleaning all Docker resources...")

	if cleanupDryRun {
		fmt.Println("[DRY RUN] Would run: docker system prune -a --volumes")

		// Show what would be cleaned
		cmd := exec.Command("docker", "system", "df")
		if output, err := cmd.Output(); err == nil {
			fmt.Println("\nCurrent usage:")
			fmt.Print(string(output))
		}
	} else {
		// Run aggressive cleanup
		fmt.Println("Running: docker system prune -a --volumes -f")
		fmt.Println("This will remove:")
		fmt.Println("  - All stopped containers")
		fmt.Println("  - All networks not used by containers")
		fmt.Println("  - All images without containers")
		fmt.Println("  - All build cache")
		fmt.Println("  - All anonymous volumes")
		fmt.Println("\n⏳ Starting cleanup (this may take 30-60 seconds)...")

		cmd := exec.Command("docker", "system", "prune", "-a", "--volumes", "-f")
		if output, err := cmd.Output(); err == nil {
			fmt.Print(string(output))
		} else {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Show disk usage after cleanup
		fmt.Println("\n📊 Calculating final disk usage (please wait)...")
		cmd = exec.Command("docker", "system", "df")
		if output, err := cmd.Output(); err == nil {
			fmt.Print(string(output))
		}
	}
}

func cleanupDockerContainers() {
	fmt.Println("🧹 Cleaning up Docker containers...")

	// Create container host
	host, err := docker.NewContainerHost()
	if err != nil {
		fmt.Printf("Error: Failed to connect to Docker: %v\n", err)
		os.Exit(1)
	}

	// Parse older-than duration if specified
	var olderThan time.Duration
	if cleanupOlderThan != "" {
		// Support simplified format: "7d" -> 7 days
		durStr := cleanupOlderThan
		if strings.HasSuffix(durStr, "d") && !strings.Contains(durStr, "h") {
			days := strings.TrimSuffix(durStr, "d")
			durStr = days + "h" // Convert to hours for parsing
			if parsed, parseErr := time.ParseDuration(durStr); parseErr == nil {
				olderThan = parsed * 24 // Convert hours to days
			} else {
				fmt.Printf("Error: Invalid duration format '%s'. Use format like '24h' or '7d'\n", cleanupOlderThan)
				os.Exit(1)
			}
		} else {
			olderThan, err = time.ParseDuration(durStr)
			if err != nil {
				fmt.Printf("Error: Invalid duration format '%s'. Use format like '24h' or '7d'\n", cleanupOlderThan)
				os.Exit(1)
			}
		}
	}

	// Build cleanup options
	opts := docker.ContainerCleanupOptions{
		OnlyExtensions: cleanupExtensions,
		IncludeRunning: false,
		OlderThan:      olderThan,
		DryRun:         cleanupDryRun,
	}

	// Execute cleanup
	result, err := host.CleanupContainers(opts)
	if err != nil {
		fmt.Printf("Error during cleanup: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if cleanupDryRun {
		fmt.Printf("  [DRY RUN] Would remove %d container(s)\n", result.ContainersRemoved)
	} else {
		fmt.Printf("  ✅ Removed %d container(s)\n", result.ContainersRemoved)
		if result.SpaceReclaimed > 0 {
			fmt.Printf("  💾 Space reclaimed: %s\n", formatBytes(result.SpaceReclaimed))
		}
	}

	// Display any errors
	if len(result.Errors) > 0 {
		fmt.Printf("  ⚠️  %d error(s) occurred:\n", len(result.Errors))
		for _, cleanupErr := range result.Errors {
			fmt.Printf("     - %v\n", cleanupErr)
		}
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
