package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
)

func TestComputeLintInputHash_NilContext(t *testing.T) {
	// When the lint context is not the expected type, return empty string.
	ctx := &cmdframework.ExecutionContext{
		Config: &cmdframework.CommandConfig{
			LintCmdContext: nil,
		},
	}

	got := computeLintInputHash(ctx, "mod-a")
	if got != "" {
		t.Errorf("computeLintInputHash() = %q, want empty string for nil context", got)
	}
}

func TestComputeLintInputHash_WrongContextType(t *testing.T) {
	// When LintCmdContext is not *lintContext, return empty string.
	ctx := &cmdframework.ExecutionContext{
		Config: &cmdframework.CommandConfig{
			LintCmdContext: "wrong-type",
		},
	}

	got := computeLintInputHash(ctx, "mod-a")
	if got != "" {
		t.Errorf("computeLintInputHash() = %q, want empty string for wrong context type", got)
	}
}

func TestComputeLintInputHash_PrecomputedHash(t *testing.T) {
	tests := []struct {
		name   string
		module string
		hashes map[string]string
		want   string
	}{
		{
			name:   "returns pre-computed hash when available",
			module: "mod-a",
			hashes: map[string]string{
				"mod-a": "abc123",
				"mod-b": "def456",
			},
			want: "abc123",
		},
		{
			name:   "returns different hash for different module",
			module: "mod-b",
			hashes: map[string]string{
				"mod-a": "abc123",
				"mod-b": "def456",
			},
			want: "def456",
		},
		{
			name:   "returns empty string when module has empty hash value",
			module: "mod-empty",
			hashes: map[string]string{
				"mod-empty": "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lctx := &lintContext{
				moduleInputHashes: tt.hashes,
				moduleFiles:       make(map[string][]string),
			}
			ctx := &cmdframework.ExecutionContext{
				Config: &cmdframework.CommandConfig{
					LintCmdContext: lctx,
				},
			}

			got := computeLintInputHash(ctx, tt.module)
			if got != tt.want {
				t.Errorf("computeLintInputHash() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeLintInputHash_EmptyPrecomputedHashes(t *testing.T) {
	// When moduleInputHashes is an empty map (not nil), the lookup fails
	// gracefully and falls through to the file-based path.
	// With empty moduleFiles and no matching module in the map,
	// it will try the registry fallback.
	lctx := &lintContext{
		moduleInputHashes: map[string]string{},
		moduleFiles:       make(map[string][]string),
	}
	ctx := &cmdframework.ExecutionContext{
		Config: &cmdframework.CommandConfig{
			LintCmdContext: lctx,
		},
	}

	// Module is not in precomputed hashes and not in moduleFiles.
	// The fallback tries ModuleRegistry.Get which would panic on nil registry.
	// But moduleInputHashes is non-nil so it checks there first and misses,
	// then checks moduleFiles and misses, then falls to registry path.
	// We cannot safely test the registry fallback without a real registry,
	// so we test that a module present in moduleInputHashes (but empty hash) returns empty.
	lctx.moduleInputHashes["mod-empty"] = ""
	got := computeLintInputHash(ctx, "mod-empty")
	if got != "" {
		t.Errorf("computeLintInputHash() = %q, want empty string for empty hash value", got)
	}
}

func TestComputeLintInputHash_FallbackWithCachedFiles(t *testing.T) {
	// When moduleInputHashes is nil but moduleFiles has the module's files,
	// it uses those files for hashing. Since the files don't exist on disk,
	// hash.Files will fail and return empty string.
	lctx := &lintContext{
		moduleInputHashes: nil,
		moduleFiles: map[string][]string{
			"mod-a": {"nonexistent-file.go"},
		},
	}
	ctx := &cmdframework.ExecutionContext{
		Config: &cmdframework.CommandConfig{
			LintCmdContext: lctx,
		},
		WorkspaceRoot: t.TempDir(),
	}

	got := computeLintInputHash(ctx, "mod-a")
	// hash.Files will fail on nonexistent files, returning empty string
	if got != "" {
		t.Errorf("computeLintInputHash() = %q, want empty string for nonexistent files", got)
	}
}

func TestComputeLintInputHash_FallbackWithEmptyFileList(t *testing.T) {
	// When moduleFiles has the module but with an empty file list,
	// hash.Files should handle it (likely returns a hash of nothing or error).
	lctx := &lintContext{
		moduleInputHashes: nil,
		moduleFiles: map[string][]string{
			"mod-empty-files": {},
		},
	}
	ctx := &cmdframework.ExecutionContext{
		Config: &cmdframework.CommandConfig{
			LintCmdContext: lctx,
		},
		WorkspaceRoot: t.TempDir(),
	}

	// With empty file list, hash.Files may return an empty hash or an error.
	// Either way, the function should not panic.
	got := computeLintInputHash(ctx, "mod-empty-files")
	// We only verify no panic; the return value depends on hash.Files behavior
	_ = got
}
