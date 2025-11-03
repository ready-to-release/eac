package extensions

import (
	"fmt"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/rs/zerolog/log"
)

// Installer handles extension installation operations
type Installer struct {
	host *docker.ContainerHost
}

// NewInstaller creates a new extension installer
func NewInstaller() (*Installer, error) {
	host, err := docker.NewContainerHost()
	if err != nil {
		return nil, fmt.Errorf("failed to create container host: %w", err)
	}

	return &Installer{
		host: host,
	}, nil
}

// Close cleans up the installer resources
func (i *Installer) Close() error {
	if i.host != nil {
		return i.host.Close()
	}
	return nil
}

// EnsureExtensionImage ensures an extension's Docker image is available locally
// It returns true if the image was actually updated, false if it was already up-to-date
func (i *Installer) EnsureExtensionImage(extensionName string) (bool, error) {
	// Find extension configuration
	extConfig, err := i.host.FindExtension(extensionName)
	if err != nil {
		return false, fmt.Errorf("extension '%s' not found: %w", extensionName, err)
	}

	// Check if image exists before pulling
	beforePull, _ := i.host.InspectImage(extConfig.Image)
	imageExistedBefore := beforePull != nil
	var beforeID string
	var beforeDigest string
	if imageExistedBefore {
		beforeID = beforePull.ID
		// Get the digest from RepoDigests if available
		if len(beforePull.RepoDigests) > 0 {
			beforeDigest = beforePull.RepoDigests[0]
		}
	}

	// Ensure image exists (will pull if needed based on policy)
	if err := i.host.EnsureImageExists(extConfig.Image, extConfig.ImagePullPolicy, extConfig.LoadLocal); err != nil {
		return false, fmt.Errorf("failed to ensure image exists: %w", err)
	}

	// Check if image was actually updated
	afterPull, _ := i.host.InspectImage(extConfig.Image)
	imageExistsAfter := afterPull != nil

	// Image was updated if:
	// 1. It didn't exist before, OR
	// 2. The image ID changed, OR
	// 3. The digest changed (for images that were pulled but had same ID)
	wasUpdated := false
	if !imageExistedBefore {
		wasUpdated = true
	} else if imageExistsAfter {
		// Check if ID changed
		if beforeID != afterPull.ID {
			wasUpdated = true
		} else if len(afterPull.RepoDigests) > 0 && beforeDigest != afterPull.RepoDigests[0] {
			// Check if digest changed (can happen with manifest list updates)
			wasUpdated = true
		}
	}

	return wasUpdated, nil
}

// InstallExtension installs a single extension by name
func (i *Installer) InstallExtension(ext conf.Extension) error {
	log.Debug().Str("extension", ext.Name).Msg("Installing extension")

	pulled, err := i.EnsureExtensionImage(ext.Name)
	if err != nil {
		return fmt.Errorf("failed to install extension '%s': %w", ext.Name, err)
	}

	if pulled {
		log.Info().Str("extension", ext.Name).Msg("Extension image pulled successfully")
	} else {
		log.Info().Str("extension", ext.Name).Msg("Extension image already up to date")
	}

	return nil
}

// InstallAllExtensions installs all configured extensions
func (i *Installer) InstallAllExtensions() error {
	// Validate extensions exist
	if err := i.host.ValidateExtensions(); err != nil {
		return fmt.Errorf("extension validation failed: %w", err)
	}

	extensions := conf.Global.Extensions
	successCount := 0
	failureCount := 0

	for _, ext := range extensions {
		if err := i.InstallExtension(ext); err != nil {
			log.Error().Err(err).Str("extension", ext.Name).Msg("Failed to install extension")
			failureCount++
		} else {
			successCount++
		}
	}

	if failureCount > 0 {
		return fmt.Errorf("failed to install %d out of %d extensions", failureCount, len(extensions))
	}

	log.Info().Int("count", successCount).Msg("All extensions installed successfully")
	return nil
}

// GetContainerHost returns the underlying container host for advanced operations
func (i *Installer) GetContainerHost() *docker.ContainerHost {
	return i.host
}
