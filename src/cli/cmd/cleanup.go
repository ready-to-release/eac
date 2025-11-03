package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/spf13/cobra"
)

var (
	cleanupAll     bool
	cleanupDryRun  bool
	keepVersions   int
)

func init() {
	RootCmd.AddCommand(CleanupCmd)
	CleanupCmd.Flags().BoolVarP(&cleanupAll, "all", "a", false, "Remove all unused images, not just r2r-cli extensions")
	CleanupCmd.Flags().BoolVarP(&cleanupDryRun, "dry-run", "n", false, "Show what would be removed without actually removing")
	CleanupCmd.Flags().IntVarP(&keepVersions, "keep", "k", 1, "Number of versions to keep per extension (default: 1)")
}

var CleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old extension images to free disk space",
	Long: `Remove old versions of extension Docker images to reclaim disk space.

By default, keeps only the most recent version of each extension.
Use --keep to retain more versions, or --all to also clean non-extension images.`,
	Example: `  # Remove all but the latest version of each extension
  r2r cleanup

  # Keep 3 versions of each extension
  r2r cleanup --keep 3

  # Show what would be removed without actually removing
  r2r cleanup --dry-run

  # Clean all Docker images (not just extensions)
  r2r cleanup --all`,
	Run: func(cmd *cobra.Command, args []string) {
		conf.InitConfig()

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
