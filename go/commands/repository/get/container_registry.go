package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/tool"
)

// ContainerRegistryQuerier queries container registry for image metadata.
type ContainerRegistryQuerier interface {
	// LastBuildSHA returns the commit SHA from the latest sha-* tag
	// for the given image name. Returns ("", nil) if no sha-* tag exists.
	LastBuildSHA(ctx context.Context, imageName string) (string, error)
}

// GHCRQuerier implements ContainerRegistryQuerier using gh api.
type GHCRQuerier struct {
	toolSystem    *tool.ToolSystem
	owner         string
	workspaceRoot string
}

// NewGHCRQuerier creates a new GHCR querier.
func NewGHCRQuerier(ts *tool.ToolSystem, owner, workspaceRoot string) *GHCRQuerier {
	return &GHCRQuerier{
		toolSystem:    ts,
		owner:         owner,
		workspaceRoot: workspaceRoot,
	}
}

func (q *GHCRQuerier) LastBuildSHA(ctx context.Context, imageName string) (string, error) {
	// Query GHCR for the latest version tags of this container image
	output, err := q.toolSystem.RunTool(ctx, "gh", q.workspaceRoot,
		"api", fmt.Sprintf("/orgs/%s/packages/container/%s/versions", q.owner, imageName),
		"--jq", `[.[].metadata.container.tags[]] | map(select(startswith("sha-"))) | first // empty`,
	)
	if err != nil {
		return "", fmt.Errorf("query GHCR tags for %s: %w", imageName, err)
	}

	tag := strings.TrimSpace(string(output))
	if tag == "" {
		return "", nil // No sha-* tag found
	}

	// Extract SHA from "sha-abc1234" -> "abc1234"
	sha := strings.TrimPrefix(tag, "sha-")
	return sha, nil
}

// mockContainerRegistryQuerier is a test double for ContainerRegistryQuerier.
type mockContainerRegistryQuerier struct {
	results map[string]string // component name -> last build SHA
	errors  map[string]error  // component name -> error
}

func (m *mockContainerRegistryQuerier) LastBuildSHA(_ context.Context, imageName string) (string, error) {
	if err, ok := m.errors[imageName]; ok {
		return "", err
	}
	return m.results[imageName], nil
}
