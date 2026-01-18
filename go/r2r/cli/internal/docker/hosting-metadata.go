package docker

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/cache"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/conf"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"gopkg.in/yaml.v3"
)

// ExecuteMetadataCommand executes the "extension-meta" command in an extension container
// and returns the raw YAML output string or an error.
func (ch *ContainerHost) ExecuteMetadataCommand(ext *ExtensionConfig) (string, error) {
	// Ensure image exists (pull if needed based on ImagePullPolicy)
	if err := ch.EnsureImageExists(ext.Image, ext.ImagePullPolicy, false); err != nil {
		return "", fmt.Errorf("error ensuring image exists: %w", err)
	}

	// Inspect image to get configuration
	imageInspect, err := ch.InspectImage(ext.Image)
	if err != nil {
		return "", fmt.Errorf("error inspecting image: %w", err)
	}

	// Create container configuration for metadata command
	containerConfig := ch.CreateContainerConfig(ext, ModeRun, []string{"extension-meta"}, imageInspect)

	// Override TTY and stdin settings for metadata retrieval
	containerConfig.Tty = false
	containerConfig.OpenStdin = false

	// No cache volumes for metadata retrieval (avoid chicken-and-egg problem)
	hostConfig := ch.CreateHostConfig(ext, nil)

	// Create container
	containerID, err := ch.CreateContainer(containerConfig, hostConfig)
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}

	// Attach to container to capture output
	attachResp, err := ch.client.ContainerAttach(ch.ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("error attaching to container: %w", err)
	}
	defer attachResp.Close()

	// Start container
	if err := ch.StartContainer(containerID); err != nil {
		return "", fmt.Errorf("error starting container: %w", err)
	}

	// Create timeout context (60 seconds)
	timeoutCtx, cancel := context.WithTimeout(ch.ctx, 60*time.Second)
	defer cancel()

	// Wait for container to finish with timeout
	statusCh, errCh := ch.client.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

	// Capture output - use stdcopy to demultiplex since TTY=false
	outputChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Since TTY is disabled, Docker multiplexes stdout/stderr with headers
		// We need to use stdcopy.StdCopy to demultiplex
		var stdout, stderr bytes.Buffer
		_, err := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
		if err != nil {
			errorChan <- fmt.Errorf("error reading output: %w", err)
			return
		}
		// Return stdout content (stderr is typically empty for metadata command)
		outputChan <- stdout.String()
	}()

	// Wait for container completion or timeout
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		// Check exit code
		if status.StatusCode != 0 {
			// Try to get error output
			select {
			case output := <-outputChan:
				return "", fmt.Errorf("extension-meta command failed with exit code %d: %s", status.StatusCode, output)
			case <-time.After(1 * time.Second):
				return "", fmt.Errorf("extension-meta command failed with exit code %d", status.StatusCode)
			}
		}
	case <-timeoutCtx.Done():
		// Timeout occurred, try to stop the container
		_ = ch.StopContainer(containerID)
		return "", fmt.Errorf("extension-meta command timed out after 60 seconds")
	}

	// Get the output
	select {
	case output := <-outputChan:
		return output, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout reading command output")
	}
}

// GetExtensionMetadata fetches and caches extension metadata
// Returns cached metadata if valid, otherwise fetches fresh metadata.
func (ch *ContainerHost) GetExtensionMetadata(ext *ExtensionConfig) (*cache.ExtensionMeta, error) {
	// Get current image digest for cache validation
	imageDigest, err := ch.GetImageDigest(ext.Image)
	if err != nil {
		logging.Debugf("Failed to get image digest, will fetch fresh metadata: image=%s err=%v", ext.Image, err)
		imageDigest = "" // Continue without digest, metadata will be fetched
	}

	// Try to load cached metadata
	cachedMeta, err := cache.LoadMetadataCache(ch.rootDir, ext.Name)
	if err != nil {
		logging.Warnf("Error loading metadata cache: extension=%s err=%v", ext.Name, err)
	}

	// Check if cache is valid
	if cache.IsMetadataCacheValid(cachedMeta, imageDigest) {
		logging.Debugf("Using cached extension metadata: extension=%s digest=%s", ext.Name, imageDigest)
		return cachedMeta.Metadata, nil
	}

	// Fetch fresh metadata
	logging.Debugf("Fetching fresh extension metadata: extension=%s", ext.Name)
	yamlOutput, err := ch.ExecuteMetadataCommand(ext)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch extension metadata: %w", err)
	}

	// Parse YAML metadata
	meta, err := parseExtensionMetadata(yamlOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension metadata: %w", err)
	}

	// Save to cache
	newCache := cache.NewMetadataCache(ext.Name, imageDigest, meta)
	if err := cache.SaveMetadataCache(ch.rootDir, newCache); err != nil {
		logging.Warnf("Failed to save metadata cache: extension=%s err=%v", ext.Name, err)
		// Continue even if cache save fails
	}

	return meta, nil
}

// parseExtensionMetadata parses YAML extension metadata into ExtensionMeta struct.
func parseExtensionMetadata(yamlData string) (*cache.ExtensionMeta, error) {
	// Import yaml package is needed - using json as intermediate for simplicity
	// since the YAML structure matches our JSON struct tags
	var meta cache.ExtensionMeta

	// Use gopkg.in/yaml.v3 to parse
	if err := yaml.Unmarshal([]byte(yamlData), &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata YAML: %w", err)
	}

	return &meta, nil
}

// MergeMetadataEnv merges environment variables from extension metadata into config
// Metadata env vars are added first, then config env vars override them.
func MergeMetadataEnv(ext *ExtensionConfig, meta *cache.ExtensionMeta) {
	if meta == nil || len(meta.Env) == 0 {
		return
	}

	// Build a map of existing config env vars for deduplication
	existingEnv := make(map[string]bool)
	for _, env := range ext.Env {
		existingEnv[env.Name] = true
	}

	// Add metadata env vars that aren't already in config
	for _, metaEnv := range meta.Env {
		if !existingEnv[metaEnv.Name] {
			ext.Env = append(ext.Env, conf.EnvVar{
				Name:     metaEnv.Name,
				Value:    metaEnv.Value,
				Required: metaEnv.Required,
			})
			logging.Debugf("Added environment variable from extension metadata: env=%s required=%v", metaEnv.Name, metaEnv.Required)
		}
	}
}
