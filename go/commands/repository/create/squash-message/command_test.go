package squashmessage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/git"
)

// TestResolveBaseBranch verifies the priority chain used to determine the base branch.
// Priority: explicit --base flag > upstream tracking branch > trunk_branch config > "main"
func TestResolveBaseBranch(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *squashConfig
		setupRepo    func() *git.MockRepository
		workspaceDir func(t *testing.T) string
		want         string
	}{
		{
			name: "explicit base flag takes precedence over upstream",
			cfg: &squashConfig{
				baseBranch:   "feature/explicit",
				baseExplicit: true,
			},
			setupRepo: func() *git.MockRepository {
				return git.NewMockRepository("/tmp").
					WithUpstreamBranch("origin/main")
			},
			workspaceDir: func(t *testing.T) string { return t.TempDir() },
			want:         "feature/explicit",
		},
		{
			name: "upstream tracking branch used when no explicit flag",
			cfg:  &squashConfig{},
			setupRepo: func() *git.MockRepository {
				return git.NewMockRepository("/tmp").
					WithUpstreamBranch("develop")
			},
			workspaceDir: func(t *testing.T) string { return t.TempDir() },
			want:         "develop",
		},
		{
			name: "trunk_branch from config used when upstream returns error",
			cfg:  &squashConfig{},
			setupRepo: func() *git.MockRepository {
				m := git.NewMockRepository("/tmp")
				m.UpstreamBranchError = fmt.Errorf("no upstream")
				return m
			},
			workspaceDir: func(t *testing.T) string {
				dir := t.TempDir()
				eacDir := filepath.Join(dir, ".eac")
				if err := os.MkdirAll(eacDir, 0o755); err != nil {
					t.Fatalf("failed to create .eac dir: %v", err)
				}
				repoYML := filepath.Join(eacDir, "repository.yml")
				content := "repository:\n  trunk_branch: develop\n"
				if err := os.WriteFile(repoYML, []byte(content), 0o644); err != nil {
					t.Fatalf("failed to write repository.yml: %v", err)
				}
				return dir
			},
			want: "develop",
		},
		{
			name: "falls back to main when upstream errors and no config",
			cfg:  &squashConfig{},
			setupRepo: func() *git.MockRepository {
				m := git.NewMockRepository("/tmp")
				m.UpstreamBranchError = fmt.Errorf("no upstream")
				return m
			},
			workspaceDir: func(t *testing.T) string { return t.TempDir() },
			want:         "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			workspaceRoot := tt.workspaceDir(t)

			got := resolveBaseBranch(tt.cfg, workspaceRoot, repo)
			if got != tt.want {
				t.Errorf("resolveBaseBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}
