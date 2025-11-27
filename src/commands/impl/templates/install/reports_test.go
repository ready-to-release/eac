//go:build L1 && ov
// +build L1,ov

package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesInstallReports_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantSource string
		wantDest   string
		wantLocal  bool
		wantErr    bool
	}{
		{
			name:       "uses defaults when no flags provided",
			args:       []string{},
			wantSource: "", // Empty means use default GitHub repo
			wantDest:   ".r2r/templates/reports",
			wantLocal:  false,
			wantErr:    false,
		},
		{
			name:       "custom local source path",
			args:       []string{"--source", "./custom/templates"},
			wantSource: "./custom/templates",
			wantDest:   ".r2r/templates/reports",
			wantLocal:  true,
			wantErr:    false,
		},
		{
			name:       "custom destination path",
			args:       []string{"--destination", "./custom/templates"},
			wantSource: "", // Empty means use default GitHub repo
			wantDest:   "./custom/templates",
			wantLocal:  false,
			wantErr:    false,
		},
		{
			name:       "custom local source and destination",
			args:       []string{"--source", "./local/templates", "--destination", "./output"},
			wantSource: "./local/templates",
			wantDest:   "./output",
			wantLocal:  true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseReportsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseReportsFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// For default (no custom source), Source should be empty
			// For local source, we expect the path to be resolved against working directory
			if config.Source != tt.wantSource && !tt.wantLocal {
				t.Errorf("Source = %v, want %v", config.Source, tt.wantSource)
			}

			if config.Destination != tt.wantDest {
				t.Errorf("Destination = %v, want %v", config.Destination, tt.wantDest)
			}

			if config.isLocalSrc != tt.wantLocal {
				t.Errorf("isLocalSrc = %v, want %v", config.isLocalSrc, tt.wantLocal)
			}
		})
	}
}

func TestTemplatesInstallReports_PathResolution(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		wantAbsolute bool
	}{
		{
			name:         "relative path",
			destination:  ".r2r/templates/reports",
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
