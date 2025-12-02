//go:build L1

package internal

import (
	"testing"
)

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flag      string
		shorthand string
		want      bool
	}{
		{"flag present", []string{"--debug", "value"}, "--debug", "-d", true},
		{"shorthand present", []string{"-d", "value"}, "--debug", "-d", true},
		{"neither present", []string{"value"}, "--debug", "-d", false},
		{"empty args", []string{}, "--debug", "-d", false},
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
		{"value present", []string{"--from=main", "other"}, "--from", "main"},
		{"value with equals", []string{"--message=test=value"}, "--message", "test=value"},
		{"no value", []string{"--from", "main"}, "--from", ""},
		{"not present", []string{"other"}, "--from", ""},
		{"empty args", []string{}, "--from", ""},
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
			"mixed args",
			[]string{"branch-name", "--debug", "value", "-d"},
			[]string{"branch-name", "value"},
		},
		{
			"only positional",
			[]string{"arg1", "arg2"},
			[]string{"arg1", "arg2"},
		},
		{
			"only flags",
			[]string{"--flag", "-f"},
			[]string{},
		},
		{
			"empty",
			[]string{},
			[]string{},
		},
		{
			"with empty strings",
			[]string{"", "value", ""},
			[]string{"value"},
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
