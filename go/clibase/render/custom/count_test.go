//go:build L1 && ov
// +build L1,ov

package custom

import (
	"strings"
	"testing"
)

func TestRenderCount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "single module",
			input: `- moniker: cli
  type: application
`,
			want: "Total modules: 1\n",
		},
		{
			name: "multiple modules",
			input: `- moniker: cli
  type: application
- moniker: core
  type: library
- moniker: api
  type: service
`,
			want: "Total modules: 3\n",
		},
		{
			name:  "empty list",
			input: "[]",
			want:  "Total modules: 0\n",
		},
		{
			name:    "invalid yaml",
			input:   "{{not valid yaml",
			wantErr: true,
		},
		{
			name:    "non-list yaml",
			input:   "key: value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCount([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RenderCount() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderCount_ErrorMessage(t *testing.T) {
	_, err := RenderCount([]byte("not: a list"))
	if err == nil {
		t.Fatal("expected error for non-list YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("error should mention parsing failure, got: %v", err)
	}
}
