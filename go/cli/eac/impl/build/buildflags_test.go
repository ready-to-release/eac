package build

import (
	"testing"
)

func TestParseBuildSpecificFlags_TidyFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantTidyFirst    bool
		wantNoTidy       bool
		wantTidyExplicit bool
		wantRemaining    []string
	}{
		{
			name:             "with-tidy explicitly enables tidy",
			args:             []string{"--with-tidy", "module1"},
			wantTidyFirst:    true,
			wantNoTidy:       false,
			wantTidyExplicit: true,
			wantRemaining:    []string{"module1"},
		},
		{
			name:             "no-tidy disables tidy",
			args:             []string{"--no-tidy", "module1"},
			wantTidyFirst:    false,
			wantNoTidy:       true,
			wantTidyExplicit: true,
			wantRemaining:    []string{"module1"},
		},
		{
			name:             "default behavior preserved (no flags)",
			args:             []string{"module1"},
			wantTidyFirst:    false,
			wantNoTidy:       false,
			wantTidyExplicit: false,
			wantRemaining:    []string{"module1"},
		},
		{
			name:             "other flags pass through",
			args:             []string{"--with-tidy", "--version", "1.0.0", "module1"},
			wantTidyFirst:    true,
			wantNoTidy:       false,
			wantTidyExplicit: true,
			wantRemaining:    []string{"module1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseBuildSpecificFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseBuildSpecificFlags() error: %v", err)
			}

			if flags.TidyFirst != tt.wantTidyFirst {
				t.Errorf("TidyFirst = %v, want %v", flags.TidyFirst, tt.wantTidyFirst)
			}
			if flags.NoTidy != tt.wantNoTidy {
				t.Errorf("NoTidy = %v, want %v", flags.NoTidy, tt.wantNoTidy)
			}
			if flags.TidyExplicit != tt.wantTidyExplicit {
				t.Errorf("TidyExplicit = %v, want %v", flags.TidyExplicit, tt.wantTidyExplicit)
			}

			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
				return
			}
			for i, r := range remaining {
				if r != tt.wantRemaining[i] {
					t.Errorf("remaining[%d] = %v, want %v", i, r, tt.wantRemaining[i])
				}
			}
		})
	}
}

func TestParseBuildSpecificFlags_ArtifactsFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantArtifacts string
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "artifacts flag with all",
			args:          []string{"--artifacts", "all", "module1"},
			wantArtifacts: "all",
			wantRemaining: []string{"module1"},
		},
		{
			name:          "artifacts flag with reduced",
			args:          []string{"--artifacts", "reduced"},
			wantArtifacts: "reduced",
			wantRemaining: nil,
		},
		{
			name:          "artifacts flag with equals syntax",
			args:          []string{"--artifacts=reduced"},
			wantArtifacts: "reduced",
			wantRemaining: nil,
		},
		{
			name:          "--all is alias for --artifacts all",
			args:          []string{"--all"},
			wantArtifacts: "all",
			wantRemaining: nil,
		},
		{
			name:          "no artifacts flag keeps default empty",
			args:          []string{"module1"},
			wantArtifacts: "",
			wantRemaining: []string{"module1"},
		},
		{
			name:    "artifacts flag missing value",
			args:    []string{"--artifacts"},
			wantErr: true,
		},
		{
			name:    "artifacts flag invalid value",
			args:    []string{"--artifacts", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseBuildSpecificFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseBuildSpecificFlags() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseBuildSpecificFlags() error: %v", err)
			}

			if flags.Artifacts != tt.wantArtifacts {
				t.Errorf("Artifacts = %q, want %q", flags.Artifacts, tt.wantArtifacts)
			}

			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestParseBuildSpecificFlags_ComponentFlag(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantComponents []string
		wantRemaining  []string
		wantErr        bool
	}{
		{
			name:           "single component with space",
			args:           []string{"--component", "site", "module1"},
			wantComponents: []string{"site"},
			wantRemaining:  []string{"module1"},
		},
		{
			name:           "single component with equals",
			args:           []string{"--component=site", "module1"},
			wantComponents: []string{"site"},
			wantRemaining:  []string{"module1"},
		},
		{
			name:           "multiple components",
			args:           []string{"--component", "site", "--component", "pdf", "module1"},
			wantComponents: []string{"site", "pdf"},
			wantRemaining:  []string{"module1"},
		},
		{
			name:           "multiple components with equals",
			args:           []string{"--component=site", "--component=pdf"},
			wantComponents: []string{"site", "pdf"},
			wantRemaining:  nil,
		},
		{
			name:    "component missing value",
			args:    []string{"--component"},
			wantErr: true,
		},
		{
			name:           "no component flag",
			args:           []string{"module1"},
			wantComponents: nil,
			wantRemaining:  []string{"module1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseBuildSpecificFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseBuildSpecificFlags() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseBuildSpecificFlags() error: %v", err)
			}

			if len(flags.Components) != len(tt.wantComponents) {
				t.Errorf("Components = %v, want %v", flags.Components, tt.wantComponents)
			} else {
				for i, c := range flags.Components {
					if c != tt.wantComponents[i] {
						t.Errorf("Components[%d] = %q, want %q", i, c, tt.wantComponents[i])
					}
				}
			}

			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestRebuildUnconsumedArgs(t *testing.T) {
	tests := []struct {
		name       string
		original   []string
		remaining  []string
		positional []string
		want       []string
	}{
		{
			name:       "preserves order with component flag",
			original:   []string{"docs", "--no-tui", "--component", "site"},
			remaining:  []string{"--component"},
			positional: []string{"docs", "site"},
			want:       []string{"docs", "--component", "site"},
		},
		{
			name:       "preserves order with version flag",
			original:   []string{"core", "--version", "1.0.0", "--no-tui"},
			remaining:  []string{"--version"},
			positional: []string{"core", "1.0.0"},
			want:       []string{"core", "--version", "1.0.0"},
		},
		{
			name:       "no remaining flags",
			original:   []string{"docs", "--no-tui"},
			remaining:  nil,
			positional: []string{"docs"},
			want:       []string{"docs"},
		},
		{
			name:       "equals syntax preserved as single token",
			original:   []string{"docs", "--component=site", "--no-tui"},
			remaining:  []string{"--component=site"},
			positional: []string{"docs"},
			want:       []string{"docs", "--component=site"},
		},
		{
			name:       "multiple components",
			original:   []string{"docs", "--component", "site", "--component", "pdf"},
			remaining:  []string{"--component", "--component"},
			positional: []string{"docs", "site", "pdf"},
			want:       []string{"docs", "--component", "site", "--component", "pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rebuildUnconsumedArgs(tt.original, tt.remaining, tt.positional)
			if len(got) != len(tt.want) {
				t.Errorf("rebuildUnconsumedArgs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("rebuildUnconsumedArgs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseBuildSpecificFlags_OtherFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantVersion    string
		wantRepro      string
		wantAcceptWarn bool
		wantListArt    bool
		wantRemaining  []string
		wantErr        bool
	}{
		{
			name:          "version flag with space",
			args:          []string{"--version", "1.2.3"},
			wantVersion:   "1.2.3",
			wantRepro:     "auto",
			wantRemaining: nil,
		},
		{
			name:          "version flag with equals",
			args:          []string{"--version=2.0.0"},
			wantVersion:   "2.0.0",
			wantRepro:     "auto",
			wantRemaining: nil,
		},
		{
			name:          "reproducible flag",
			args:          []string{"--reproducible", "true"},
			wantRepro:     "true",
			wantRemaining: nil,
		},
		{
			name:           "accept-warnings flag",
			args:           []string{"--accept-warnings"},
			wantAcceptWarn: true,
			wantRepro:      "auto",
			wantRemaining:  nil,
		},
		{
			name:          "list-artifacts flag",
			args:          []string{"--list-artifacts"},
			wantListArt:   true,
			wantRepro:     "auto",
			wantRemaining: nil,
		},
		{
			name:    "version missing value",
			args:    []string{"--version"},
			wantErr: true,
		},
		{
			name:    "reproducible invalid value",
			args:    []string{"--reproducible", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseBuildSpecificFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseBuildSpecificFlags() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseBuildSpecificFlags() error: %v", err)
			}

			if flags.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", flags.Version, tt.wantVersion)
			}
			if flags.Reproducible != tt.wantRepro {
				t.Errorf("Reproducible = %v, want %v", flags.Reproducible, tt.wantRepro)
			}
			if flags.AcceptWarnings != tt.wantAcceptWarn {
				t.Errorf("AcceptWarnings = %v, want %v", flags.AcceptWarnings, tt.wantAcceptWarn)
			}
			if flags.ListArtifacts != tt.wantListArt {
				t.Errorf("ListArtifacts = %v, want %v", flags.ListArtifacts, tt.wantListArt)
			}

			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}
