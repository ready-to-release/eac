package test

import (
	"testing"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestMapGOOSToDepTag(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"linux", "linux"},
		{"darwin", "macos"},
		{"windows", "windows"},
		{"freebsd", "freebsd"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			got := mapGOOSToDepTag(tt.goos)
			if got != tt.want {
				t.Errorf("mapGOOSToDepTag(%q) = %q, want %q", tt.goos, got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no ANSI codes",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "single color code",
			input: "\x1b[33mwarning\x1b[0m",
			want:  "warning",
		},
		{
			name:  "multiple color codes",
			input: "\x1b[1m\x1b[31merror\x1b[0m text",
			want:  "error text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only ANSI codes",
			input: "\x1b[0m\x1b[33m\x1b[0m",
			want:  "",
		},
		{
			name:  "reset code",
			input: "before\x1b[0mafter",
			want:  "beforeafter",
		},
		{
			name:  "multi-digit params",
			input: "\x1b[38;5;196mred\x1b[0m",
			want:  "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			if got != tt.want {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetPackageTestType(t *testing.T) {
	tests := []struct {
		name  string
		input []coretesting.TestReference
		want  string
	}{
		{
			name:  "empty slice",
			input: nil,
			want:  "",
		},
		{
			name:  "single test",
			input: []coretesting.TestReference{{Type: "gotest"}},
			want:  "gotest",
		},
		{
			name: "all same type",
			input: []coretesting.TestReference{
				{Type: "godog"},
				{Type: "godog"},
				{Type: "godog"},
			},
			want: "godog",
		},
		{
			name: "mixed types",
			input: []coretesting.TestReference{
				{Type: "gotest"},
				{Type: "godog"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPackageTestType(tt.input)
			if got != tt.want {
				t.Errorf("getPackageTestType() = %q, want %q", got, tt.want)
			}
		})
	}
}

