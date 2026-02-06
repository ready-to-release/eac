package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// PackageVersion represents a container image version in GHCR.
type PackageVersion struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"` // Version name (usually digest)
	Tags      []string  `json:"tags"` // Associated tags
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Digest    string    `json:"digest"` // SHA256 digest
}

// Package represents a container package in GHCR.
type Package struct {
	Name        string `json:"name"`
	PackageType string `json:"package_type"`
	Visibility  string `json:"visibility"`
}

// PackagesAPI defines the interface for GitHub Packages operations.
type PackagesAPI interface {
	// ListOrgPackages lists all container packages for an organization.
	ListOrgPackages(ctx context.Context, org string) ([]Package, error)

	// ListPackageVersions lists all versions of a package.
	ListPackageVersions(ctx context.Context, org, packageName string) ([]PackageVersion, error)

	// DeletePackageVersion deletes a specific package version.
	DeletePackageVersion(ctx context.Context, org, packageName string, versionID int) error
}

// GHPackagesClient implements PackagesAPI using an injected CLIExecutor.
type GHPackagesClient struct {
	executor CLIExecutor
	workDir  string
}

// NewGHPackagesClient creates a new packages client.
// executor: a CLIExecutor that handles actual gh binary invocation.
// workDir: working directory for git context (empty = current).
func NewGHPackagesClient(executor CLIExecutor, workDir string) *GHPackagesClient {
	return &GHPackagesClient{executor: executor, workDir: workDir}
}

// ListOrgPackages lists all container packages for an organization.
func (c *GHPackagesClient) ListOrgPackages(ctx context.Context, org string) ([]Package, error) {
	output, err := c.executor.ExecContext(ctx, "api",
		fmt.Sprintf("/orgs/%s/packages?package_type=container&per_page=100", url.PathEscape(org)),
	)
	if err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}

	var packages []Package
	if err := json.Unmarshal(output, &packages); err != nil {
		return nil, fmt.Errorf("parsing packages: %w", err)
	}

	return packages, nil
}

// ListPackageVersions lists all versions of a package.
func (c *GHPackagesClient) ListPackageVersions(ctx context.Context, org, packageName string) ([]PackageVersion, error) {
	output, err := c.executor.ExecContext(ctx, "api",
		fmt.Sprintf("/orgs/%s/packages/container/%s/versions?per_page=100",
			url.PathEscape(org), url.PathEscape(packageName)),
	)
	if err != nil {
		return nil, fmt.Errorf("listing versions: %w", err)
	}

	// Parse the response - GitHub returns a different structure
	var rawVersions []struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Metadata  struct {
			Container struct {
				Tags []string `json:"tags"`
			} `json:"container"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(output, &rawVersions); err != nil {
		return nil, fmt.Errorf("parsing versions: %w", err)
	}

	versions := make([]PackageVersion, len(rawVersions))
	for i, v := range rawVersions {
		versions[i] = PackageVersion{
			ID:        v.ID,
			Name:      v.Name,
			Tags:      v.Metadata.Container.Tags,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Digest:    v.Name, // Name is typically the digest
		}
	}

	return versions, nil
}

// DeletePackageVersion deletes a specific package version.
func (c *GHPackagesClient) DeletePackageVersion(ctx context.Context, org, packageName string, versionID int) error {
	_, err := c.executor.ExecContext(ctx, "api",
		"--method", "DELETE",
		fmt.Sprintf("/orgs/%s/packages/container/%s/versions/%d",
			url.PathEscape(org), url.PathEscape(packageName), versionID),
	)
	if err != nil {
		return fmt.Errorf("deleting version %d: %w", versionID, err)
	}

	return nil
}
