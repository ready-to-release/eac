package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesInstallSpecs_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantSource     string
		wantSourcePath string
		wantDest       string
		wantErr        bool
	}{
		{
			name:           "uses defaults when no flags provided",
			args:           []string{},
			wantSource:     "https://github.com/ready-to-release/eac",
			wantSourcePath: "templates/specs",
			wantDest:       ".r2r/templates/specs",
			wantErr:        false,
		},
		{
			name:           "custom source repository",
			args:           []string{"--source", "https://github.com/custom/repo"},
			wantSource:     "https://github.com/custom/repo",
			wantSourcePath: "templates/specs",
			wantDest:       ".r2r/templates/specs",
			wantErr:        false,
		},
		{
			name:           "custom destination path",
			args:           []string{"--destination", "./custom/templates"},
			wantSource:     "https://github.com/ready-to-release/eac",
			wantSourcePath: "templates/specs",
			wantDest:       "./custom/templates",
			wantErr:        false,
		},
		{
			name:           "custom source and destination",
			args:           []string{"--source", "https://github.com/custom/repo", "--destination", "./output"},
			wantSource:     "https://github.com/custom/repo",
			wantSourcePath: "templates/specs",
			wantDest:       "./output",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseSpecsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSpecsFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if config.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", config.Source, tt.wantSource)
			}
			if config.SourcePath != tt.wantSourcePath {
				t.Errorf("SourcePath = %v, want %v", config.SourcePath, tt.wantSourcePath)
			}
			if config.Destination != tt.wantDest {
				t.Errorf("Destination = %v, want %v", config.Destination, tt.wantDest)
			}
		})
	}
}

func TestTemplatesInstallSpecs_PathResolution(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		wantAbsolute bool
	}{
		{
			name:         "relative path",
			destination:  ".r2r/templates/specs",
			wantAbsolute: false,
		},
		{
			name:         "absolute path",
			destination:  filepath.Join(os.TempDir(), "templates"),
			wantAbsolute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAbs := filepath.IsAbs(tt.destination)
			if isAbs != tt.wantAbsolute {
				t.Errorf("IsAbs() = %v, want %v for path %v", isAbs, tt.wantAbsolute, tt.destination)
			}
		})
	}
}
