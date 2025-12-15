package flags

import (
	"testing"
)

func TestParseDebugFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "no debug flag",
			args: []string{"arg1", "arg2"},
			want: false,
		},
		{
			name: "debug flag long form",
			args: []string{"arg1", "--debug", "arg2"},
			want: true,
		},
		{
			name: "debug flag short form",
			args: []string{"arg1", "-d", "arg2"},
			want: true,
		},
		{
			name: "debug flag at start",
			args: []string{"--debug", "arg1", "arg2"},
			want: true,
		},
		{
			name: "debug flag at end",
			args: []string{"arg1", "arg2", "-d"},
			want: true,
		},
		{
			name: "empty args",
			args: []string{},
			want: false,
		},
		{
			name: "similar but not debug",
			args: []string{"--debugger", "-debug"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDebugFlag(tt.args)
			if got != tt.want {
				t.Errorf("ParseDebugFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flag      string
		shorthand string
		want      bool
	}{
		{
			name:      "flag present - long form",
			args:      []string{"--all", "arg1"},
			flag:      "--all",
			shorthand: "-a",
			want:      true,
		},
		{
			name:      "flag present - short form",
			args:      []string{"-a", "arg1"},
			flag:      "--all",
			shorthand: "-a",
			want:      true,
		},
		{
			name:      "flag not present",
			args:      []string{"arg1", "arg2"},
			flag:      "--all",
			shorthand: "-a",
			want:      false,
		},
		{
			name:      "no shorthand",
			args:      []string{"--verbose"},
			flag:      "--verbose",
			shorthand: "",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasFlag(tt.args, tt.flag, tt.shorthand)
			if got != tt.want {
				t.Errorf("HasFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFlagValue(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		prefix string
		want   string
	}{
		{
			name:   "flag with value",
			args:   []string{"--target=main", "arg1"},
			prefix: "--target",
			want:   "main",
		},
		{
			name:   "flag not present",
			args:   []string{"arg1", "arg2"},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "flag without value",
			args:   []string{"--target", "arg1"},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "empty value",
			args:   []string{"--target="},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "value with equals",
			args:   []string{"--config=/path/to=file"},
			prefix: "--config",
			want:   "/path/to=file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFlagValue(tt.args, tt.prefix)
			if got != tt.want {
				t.Errorf("GetFlagValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPositionalArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "mixed args and flags",
			args: []string{"arg1", "--flag", "arg2", "-f"},
			want: []string{"arg1", "arg2"},
		},
		{
			name: "only positional args",
			args: []string{"arg1", "arg2", "arg3"},
			want: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "only flags",
			args: []string{"--flag1", "-f", "--flag2"},
			want: nil,
		},
		{
			name: "empty args",
			args: []string{},
			want: nil,
		},
		{
			name: "empty string in args",
			args: []string{"arg1", "", "arg2"},
			want: []string{"arg1", "arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPositionalArgs(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("GetPositionalArgs() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetPositionalArgs()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
