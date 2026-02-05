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
			name:             "tidy-first still works (backward compat)",
			args:             []string{"--tidy-first", "module1"},
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
