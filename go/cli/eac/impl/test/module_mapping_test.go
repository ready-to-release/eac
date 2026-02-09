package test

import (
	"testing"
)

func TestExtractPathSuffix(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		moniker  string
		want     string
	}{
		{
			name:    "standard module path",
			pkgPath: "go/eac/core/config/test",
			moniker: "eac-core",
			want:    "config/test",
		},
		{
			name:    "exact module root match",
			pkgPath: "go/eac/core",
			moniker: "eac-core",
			want:    "",
		},
		{
			name:    "specs implementation path",
			pkgPath: "go/specs/impl/eac/repository",
			moniker: "eac",
			want:    "repository",
		},
		{
			name:    "no match returns empty",
			pkgPath: "some/other/path",
			moniker: "eac-core",
			want:    "",
		},
		{
			name:    "cli module path",
			pkgPath: "go/clie/commands/release",
			moniker: "clie-commands",
			want:    "release",
		},
		{
			name:    "hyphenated moniker",
			pkgPath: "go/eac/module-deps/verify",
			moniker: "eac-module-deps",
			want:    "verify",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPathSuffix(tt.pkgPath, tt.moniker)
			if got != tt.want {
				t.Errorf("extractPathSuffix(%q, %q) = %q, want %q",
					tt.pkgPath, tt.moniker, got, tt.want)
			}
		})
	}
}

func TestBuildModuleOutputPath(t *testing.T) {
	mapper := &ModuleMapper{} // Fields not needed for this function

	tests := []struct {
		name     string
		pkgPath  string
		moniker  string
		want     string
	}{
		{
			name:    "unknown module",
			pkgPath: "some/path",
			moniker: "",
			want:    "unknown/packages/some/path",
		},
		{
			name:    "godog three-part path",
			pkgPath: "login:go/specs/impl/eac:specs/eac/login",
			moniker: "eac",
			want:    "eac/packages/login",
		},
		{
			name:    "module with no suffix",
			pkgPath: "go/eac/core",
			moniker: "eac-core",
			want:    "eac-core/packages/root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.BuildModuleOutputPath(tt.pkgPath, tt.moniker)
			if got != tt.want {
				t.Errorf("BuildModuleOutputPath(%q, %q) = %q, want %q",
					tt.pkgPath, tt.moniker, got, tt.want)
			}
		})
	}
}
