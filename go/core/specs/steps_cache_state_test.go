// Package specs contains godog step implementations for core features.
//
// This file contains build and lint state management functions for cache
// invalidation tests, including creating manifests and simulating
// successful or failed build/lint outcomes.
package specs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/output"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

// withContainerRoot temporarily sets CLIE_CONTAINER_ROOT to the original repo root
// and returns a cleanup function that restores the previous value.
func withContainerRoot(ctx *eacgodog.TestContext) func() {
	orig := os.Getenv(environments.EnvCLIEContainerRoot)
	os.Setenv(environments.EnvCLIEContainerRoot, ctx.OriginalRepoRoot)
	return func() {
		if orig == "" {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		} else {
			os.Setenv(environments.EnvCLIEContainerRoot, orig)
		}
	}
}

func buildAllModules(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()
	defer withContainerRoot(ctx)()

	reg, err := getOrLoadRegistry(ctx.IsolatedDir)
	if err != nil {
		return err
	}

	// Create UoW manifests for each module (simulating a successful build)
	for _, contract := range reg.All() {
		patterns := contract.GetGlobPatterns()
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, patterns)
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", contract.Moniker, err)
		}
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", contract.Moniker, err)
		}
		// Create UoW manifest in the expected location
		// Format: out/build/<module>/<component>-<tool>/uow.manifest.json
		manifest := &output.UoWManifest{
			Action:     core.ActionBuild,
			Module:     contract.Moniker,
			Component:  "go",  // Default component for Go modules
			Tool:       "go",  // Default tool
			ExitCode:   0,
			InputHash:  sourceHash,
			ExecutedAt: time.Now().UTC().Truncate(time.Second),
			Duration:   time.Second,
			Artifacts:  []output.Artifact{},
			OutputHash: "sha256:output-" + contract.Moniker,
			Version:    "1.0.0",
		}

		manifestDir := filepath.Join(ctx.IsolatedDir, "out", "build", contract.Moniker, "go_go")
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifest dir for %s: %w", contract.Moniker, err)
		}

		manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal manifest for %s: %w", contract.Moniker, err)
		}
		if err := os.WriteFile(manifestPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write manifest for %s: %w", contract.Moniker, err)
		}
	}

	return nil
}

func lintModuleSuccessfully(ctx *eacgodog.TestContext, module string) error {
	ctx.MustBeIsolated()
	defer withContainerRoot(ctx)()

	reg, err := getOrLoadRegistry(ctx.IsolatedDir)
	if err != nil {
		return err
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create UoW manifest for successful lint
	manifest := &output.UoWManifest{
		Action:     core.ActionLint,
		Module:     module,
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   0,
		InputHash:  sourceHash,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   time.Second,
		Artifacts:  []output.Artifact{},
		OutputHash: "sha256:lint-output-" + module,
		Version:    "1.0.0",
	}

	manifestDir := filepath.Join(ctx.IsolatedDir, "out", "lint", module, "go_golangci-lint")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest dir for %s: %w", module, err)
	}

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for %s: %w", module, err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest for %s: %w", module, err)
	}

	return nil
}

func setLintStateFailed(ctx *eacgodog.TestContext, module string) error {
	ctx.MustBeIsolated()
	defer withContainerRoot(ctx)()

	reg, err := getOrLoadRegistry(ctx.IsolatedDir)
	if err != nil {
		return err
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create UoW manifest for failed lint (exit_code: 1)
	manifest := &output.UoWManifest{
		Action:     core.ActionLint,
		Module:     module,
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   1, // Failed
		InputHash:  sourceHash,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   time.Second,
		Artifacts:  []output.Artifact{},
		OutputHash: "sha256:lint-output-" + module,
		Version:    "1.0.0",
	}

	manifestDir := filepath.Join(ctx.IsolatedDir, "out", "lint", module, "go_golangci-lint")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest dir for %s: %w", module, err)
	}

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for %s: %w", module, err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest for %s: %w", module, err)
	}

	return nil
}

func ensureNoModificationsInDir(ctx *eacgodog.TestContext, dir string) error {
	ctx.MustBeIsolated()
	_ = coretesting.GitCheckoutPath(ctx.IsolatedDir, dir)
	return nil
}

func buildSpecificModules(ctx *eacgodog.TestContext, mod1, mod2 string) error {
	ctx.MustBeIsolated()
	defer withContainerRoot(ctx)()

	reg, err := getOrLoadRegistry(ctx.IsolatedDir)
	if err != nil {
		return err
	}

	moduleNames := []string{mod1, mod2}
	for _, moniker := range moduleNames {
		contract, ok := reg.Get(moniker)
		if !ok {
			return fmt.Errorf("module not found: %s", moniker)
		}
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", moniker, err)
		}
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", moniker, err)
		}

		// Create UoW manifest for successful build
		manifest := &output.UoWManifest{
			Action:     core.ActionBuild,
			Module:     moniker,
			Component:  "go",
			Tool:       "go",
			ExitCode:   0,
			InputHash:  sourceHash,
			ExecutedAt: time.Now().UTC().Truncate(time.Second),
			Duration:   time.Second,
			Artifacts:  []output.Artifact{},
			OutputHash: "sha256:output-" + moniker,
			Version:    "1.0.0",
		}

		manifestDir := filepath.Join(ctx.IsolatedDir, "out", "build", moniker, "go_go")
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifest dir for %s: %w", moniker, err)
		}

		manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal manifest for %s: %w", moniker, err)
		}
		if err := os.WriteFile(manifestPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write manifest for %s: %w", moniker, err)
		}
	}

	return nil
}
