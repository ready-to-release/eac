package scanner

import "testing"

func TestFS_DefaultFiles(t *testing.T) {
	files := []string{"scanners.yml", "policies.yml", "risk-config.yml"}
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			data, err := FS.ReadFile(DefaultPath(f))
			if err != nil {
				t.Fatalf("FS.ReadFile(%q) error: %v", DefaultPath(f), err)
			}
			if len(data) == 0 {
				t.Fatalf("FS.ReadFile(%q) returned empty data", DefaultPath(f))
			}
		})
	}
}
