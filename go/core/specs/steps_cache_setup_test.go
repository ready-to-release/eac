// Package specs contains godog step implementations for core features.
//
// This file contains repository setup, module structure scaffolding,
// and file manipulation helpers for cache invalidation tests.
package specs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// cacheModuleConfig describes a module in the test repository.
type cacheModuleConfig struct {
	Moniker   string
	DependsOn string
	GoRoot    string
}

// captureHeadSHA captures the current HEAD SHA after isolation is set up.
func captureHeadSHA(ctx *eacgodog.TestContext) {
	if ctx.IsolatedDir == "" {
		return
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}
}

func setupMultiModuleStructure(ctx *eacgodog.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()

	clieDir := filepath.Join(ctx.IsolatedDir, ".eac")
	if err := os.MkdirAll(clieDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .clie directory: %w", err)
	}

	var mods []cacheModuleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) < 3 {
			continue
		}
		mods = append(mods, cacheModuleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    row.Cells[2].Value,
		})
	}

	repoYAML := generateCacheRepositoryYAML(mods)
	repoPath := filepath.Join(clieDir, "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	for _, mod := range mods {
		goDir := filepath.Join(ctx.IsolatedDir, mod.GoRoot)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			return fmt.Errorf("failed to create go directory: %w", err)
		}

		mainPath := filepath.Join(goDir, "main.go")
		mainContent := fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", mod.Moniker)
		if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}

		helperPath := filepath.Join(goDir, "helper.go")
		helperContent := fmt.Sprintf("package main\n\n// Helper functions for %s\n", mod.Moniker)
		if err := os.WriteFile(helperPath, []byte(helperContent), 0o644); err != nil {
			return fmt.Errorf("failed to write helper.go: %w", err)
		}

		internalDir := filepath.Join(goDir, "internal")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			return fmt.Errorf("failed to create internal directory: %w", err)
		}

		handlerPath := filepath.Join(internalDir, "handler.go")
		handlerContent := fmt.Sprintf("package internal\n\n// Handler for %s\n", mod.Moniker)
		if err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644); err != nil {
			return fmt.Errorf("failed to write handler.go: %w", err)
		}

		changelogPath := filepath.Join(goDir, "CHANGELOG.md")
		changelogContent := fmt.Sprintf("# Changelog for %s\n\n## v1.0.0\n- Initial release\n", mod.Moniker)
		if err := os.WriteFile(changelogPath, []byte(changelogContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
		}
	}

	workflowsDir := filepath.Join(ctx.IsolatedDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}
	for _, mod := range mods {
		workflowPath := filepath.Join(workflowsDir, fmt.Sprintf("ci-%s.yaml", mod.Moniker))
		workflowContent := fmt.Sprintf("name: CI %s\non: [push]\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n", mod.Moniker)
		if err := os.WriteFile(workflowPath, []byte(workflowContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CI workflow for %s: %w", mod.Moniker, err)
		}
	}

	gitignorePath := filepath.Join(ctx.IsolatedDir, ".gitignore")
	gitignoreContent := "out/\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add multi-module structure for cache invalidation tests")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}

	return nil
}

func generateCacheRepositoryYAML(mods []cacheModuleConfig) string {
	var sb strings.Builder
	sb.WriteString("# Auto-generated repository.yml for cache invalidation tests\n\n")
	sb.WriteString("repository:\n")
	sb.WriteString("  type: mono\n")
	sb.WriteString("  remote:\n")
	sb.WriteString("    owner: test-owner\n")
	sb.WriteString("    repo: test-repo\n\n")
	sb.WriteString("modules:\n")

	for _, mod := range mods {
		sb.WriteString(fmt.Sprintf("  - moniker: %s\n", mod.Moniker))
		sb.WriteString(fmt.Sprintf("    name: %s Module\n", mod.Moniker))
		sb.WriteString(fmt.Sprintf("    description: Test module for %s\n", mod.Moniker))
		if mod.DependsOn != "" {
			sb.WriteString("    depends_on:\n")
			for _, dep := range strings.Split(mod.DependsOn, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					sb.WriteString(fmt.Sprintf("      - %s\n", dep))
				}
			}
		}
		sb.WriteString("    components:\n")
		sb.WriteString("      - type: go\n")
		sb.WriteString("        name: go\n")
		sb.WriteString(fmt.Sprintf("        root: %s\n", mod.GoRoot))
		sb.WriteString("        patterns:\n")
		sb.WriteString("          source: [\"**/*.go\"]\n")
	}

	return sb.String()
}

func ensureNoBuildState(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()
	buildDir := filepath.Join(ctx.IsolatedDir, "out", "build")
	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to remove build state directory: %w", err)
	}
	return nil
}

func ensureNoModifications(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()
	cmd := exec.Command("git", "checkout", "--", ".")
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run()
	return nil
}

func appendToFile(ctx *eacgodog.TestContext, content, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + content + "\n"); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}

	return nil
}

func deleteFile(ctx *eacgodog.TestContext, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func addNewModule(ctx *eacgodog.TestContext, moniker, goRoot string) error {
	ctx.MustBeIsolated()
	repoPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

	newConfig := fmt.Sprintf(`  - moniker: %s
    name: %s Module
    description: Test module for %s
    components:
      - type: go
        root: %s
        patterns:
          source: ["**/*.go"]
`, moniker, moniker, moniker, goRoot)

	newContent := string(content) + newConfig
	if err := os.WriteFile(repoPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	goDir := filepath.Join(ctx.IsolatedDir, goRoot)
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		return fmt.Errorf("failed to create go directory: %w", err)
	}

	mainPath := filepath.Join(goDir, "main.go")
	mainContent := fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", moniker)
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

func setupCircularDependency(ctx *eacgodog.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()
	var mods []cacheModuleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) < 2 {
			continue
		}

		mods = append(mods, cacheModuleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    fmt.Sprintf("go/%s", row.Cells[0].Value),
		})
	}

	repoYAML := generateCacheRepositoryYAML(mods)
	repoPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	for _, mod := range mods {
		goDir := filepath.Join(ctx.IsolatedDir, mod.GoRoot)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			return fmt.Errorf("failed to create go directory: %w", err)
		}

		mainPath := filepath.Join(goDir, "main.go")
		mainContent := "package main\n\nfunc main() {}\n"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}
	}

	return nil
}

func corruptBuildState(ctx *eacgodog.TestContext, content string) error {
	ctx.MustBeIsolated()
	buildDir := filepath.Join(ctx.IsolatedDir, "out", "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	statePath := filepath.Join(buildDir, ".build-state.json")
	if err := os.WriteFile(statePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write corrupted build state: %w", err)
	}

	return nil
}
