package docker

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// ImagePullRequest represents a single image pull operation
type ImagePullRequest struct {
	ImageName   string
	PullPolicy  string
	LoadLocal   bool
	Index       int // Original index for result mapping
}

// ImagePullResult represents the result of an image pull operation
type ImagePullResult struct {
	ImageName string
	Index     int
	Error     error
	Success   bool
}

// ParallelImagePullOptions configures parallel image pull behavior
type ParallelImagePullOptions struct {
	MaxConcurrency int // Maximum number of concurrent pulls (default: 3)
}

// ParallelEnsureImages pulls multiple images in parallel
// This is useful when starting multiple extensions that may need image pulls
func (ch *ContainerHost) ParallelEnsureImages(requests []ImagePullRequest, opts *ParallelImagePullOptions) []ImagePullResult {
	if len(requests) == 0 {
		return []ImagePullResult{}
	}

	// Default concurrency
	maxConcurrency := 3
	if opts != nil && opts.MaxConcurrency > 0 {
		maxConcurrency = opts.MaxConcurrency
	}

	log.Debug().
		Int("image_count", len(requests)).
		Int("max_concurrency", maxConcurrency).
		Msg("Starting parallel image pulls")

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Pre-allocate results slice
	results := make([]ImagePullResult, len(requests))

	// Process each image pull request
	for i, req := range requests {
		wg.Add(1)
		go func(idx int, request ImagePullRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Debug().
				Str("image", request.ImageName).
				Str("policy", request.PullPolicy).
				Int("index", idx).
				Msg("Processing image pull")

			// Execute the image pull
			err := ch.EnsureImageExists(request.ImageName, request.PullPolicy, request.LoadLocal)

			// Store result
			results[idx] = ImagePullResult{
				ImageName: request.ImageName,
				Index:     request.Index,
				Error:     err,
				Success:   err == nil,
			}

			if err != nil {
				log.Warn().
					Err(err).
					Str("image", request.ImageName).
					Msg("Image pull failed")
			} else {
				log.Debug().
					Str("image", request.ImageName).
					Msg("Image pull succeeded")
			}
		}(i, req)
	}

	// Wait for all pulls to complete
	wg.Wait()

	// Count successes and failures
	successes := 0
	failures := 0
	for _, result := range results {
		if result.Success {
			successes++
		} else {
			failures++
		}
	}

	log.Info().
		Int("total", len(results)).
		Int("successes", successes).
		Int("failures", failures).
		Msg("Parallel image pulls completed")

	return results
}

// EnsureMultipleImages is a convenience wrapper for pulling multiple images with the same policy
func (ch *ContainerHost) EnsureMultipleImages(imageNames []string, pullPolicy string, loadLocal bool) []ImagePullResult {
	requests := make([]ImagePullRequest, len(imageNames))
	for i, imageName := range imageNames {
		requests[i] = ImagePullRequest{
			ImageName:  imageName,
			PullPolicy: pullPolicy,
			LoadLocal:  loadLocal,
			Index:      i,
		}
	}

	return ch.ParallelEnsureImages(requests, nil)
}
