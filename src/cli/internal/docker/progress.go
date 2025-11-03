package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

// DockerProgress represents a Docker progress update
type DockerProgress struct {
	Status         string `json:"status"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
	ID string `json:"id"`
}

// DockerError represents a Docker error response
type DockerError struct {
	Error string `json:"error"`
}

// DisplayDockerProgress reads Docker JSON progress and displays it to the user
func DisplayDockerProgress(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	layerProgress := make(map[string]string)
	lastStatus := ""
	hasActualDownload := false // Track if we're actually downloading anything
	showedPullingMessage := false // Track if we've shown the initial pulling message
	dotCount := 0 // Track number of progress dots shown

	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse as Docker error first
		var dockerErr DockerError
		if err := json.Unmarshal([]byte(line), &dockerErr); err == nil && dockerErr.Error != "" {
			return fmt.Errorf("docker error: %s", dockerErr.Error)
		}

		// Parse as progress update
		var progress DockerProgress
		if err := json.Unmarshal([]byte(line), &progress); err != nil {
			// If we can't parse it, log it as debug
			log.Debug().Str("line", line).Msg("Unparseable Docker output")
			continue
		}

		// Build status message
		statusMsg := progress.Status

		// Handle different status types
		switch {
		case strings.HasPrefix(progress.Status, "Pulling from"):
			// Don't show immediately - wait to see if we actually download
			lastStatus = progress.Status

		case progress.Status == "Pulling fs layer" || progress.Status == "Waiting":
			// Track layer but don't display yet
			if progress.ID != "" {
				layerProgress[progress.ID] = "⏳ Waiting"
			}

		case progress.Status == "Downloading":
			// We have actual downloads - show the pulling message if not shown yet
			if !showedPullingMessage && lastStatus != "" {
				fmt.Printf("📦 %s\n", lastStatus)
				fmt.Printf("Downloading image")
				showedPullingMessage = true
			}
			hasActualDownload = true
			// Show progress as dots instead of detailed layer info
			if progress.ID != "" {
				layerProgress[progress.ID] = "downloading"
				showProgressDot(&dotCount)
			}

		case progress.Status == "Extracting":
			// Extracting means we downloaded something
			if !showedPullingMessage && lastStatus != "" {
				fmt.Printf("📦 %s\n", lastStatus)
				fmt.Printf("Downloading image")
				showedPullingMessage = true
			}
			hasActualDownload = true
			// Show progress as dots
			if progress.ID != "" {
				layerProgress[progress.ID] = "extracting"
				showProgressDot(&dotCount)
			}

		case progress.Status == "Pull complete":
			// Mark layer as complete and show dot
			if progress.ID != "" {
				layerProgress[progress.ID] = "complete"
				showProgressDot(&dotCount)
			}

		case progress.Status == "Already exists":
			// Layer already exists locally - don't show progress for this
			if progress.ID != "" {
				log.Debug().Str("layer", progress.ID).Msg("Layer already exists")
			}

		case strings.Contains(progress.Status, "Downloaded newer image"):
			// Show final status for actual downloads
			if !showedPullingMessage && lastStatus != "" {
				// Show the pulling message before the completion
				fmt.Printf("📦 %s\n", lastStatus)
			}
			if dotCount > 0 {
				fmt.Printf("\n") // End the dots line
			}
			fmt.Printf("✅ %s\n", progress.Status)

		case strings.Contains(progress.Status, "Image is up to date"):
			// Only show "up to date" message if we showed pulling info
			if hasActualDownload || showedPullingMessage {
				if dotCount > 0 {
					fmt.Printf("\n") // End the dots line
				}
				fmt.Printf("✅ %s\n", progress.Status)
			}
			// Otherwise stay silent - image was already present

		case progress.Status == "Download complete":
			// Layer download complete
			if progress.ID != "" {
				layerProgress[progress.ID] = fmt.Sprintf("💾 %s: Downloaded", progress.ID[:12])
			}

		default:
			// Other status messages
			if statusMsg != "" && statusMsg != lastStatus {
				log.Debug().Str("status", statusMsg).Msg("Docker pull status")
				lastStatus = statusMsg
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading Docker progress: %w", err)
	}

	return nil
}

// showProgressDot displays a single dot to show progress, limiting to reasonable length
func showProgressDot(dotCount *int) {
	// Limit dots to avoid overly long lines (max ~50 dots)
	if *dotCount < 50 {
		fmt.Print(".")
		*dotCount++
	}
}
