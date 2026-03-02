package deploy

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

func TestParseDeployArgs_ValidArgs(t *testing.T) {
	env := &environment.Env{}

	shared, deployFlags, module, envMoniker, err := parseDeployArgs(
		[]string{"infra", "development"},
		env,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if module != "infra" {
		t.Errorf("expected module 'infra', got %q", module)
	}
	if envMoniker != "development" {
		t.Errorf("expected env 'development', got %q", envMoniker)
	}
	if shared.DryRun {
		t.Error("expected DryRun=false")
	}
	if deployFlags.Component != "" {
		t.Errorf("expected empty component, got %q", deployFlags.Component)
	}
}

func TestParseDeployArgs_WithDryRun(t *testing.T) {
	env := &environment.Env{}

	shared, _, module, envMoniker, err := parseDeployArgs(
		[]string{"infra", "production", "--dry-run"},
		env,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if module != "infra" {
		t.Errorf("expected module 'infra', got %q", module)
	}
	if envMoniker != "production" {
		t.Errorf("expected env 'production', got %q", envMoniker)
	}
	if !shared.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestParseDeployArgs_WithComponent(t *testing.T) {
	env := &environment.Env{}

	_, deployFlags, _, _, err := parseDeployArgs(
		[]string{"infra", "development", "--component", "networking"},
		env,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deployFlags.Component != "networking" {
		t.Errorf("expected component 'networking', got %q", deployFlags.Component)
	}
}

func TestParseDeployArgs_MissingEnv(t *testing.T) {
	env := &environment.Env{}

	_, _, _, _, err := parseDeployArgs(
		[]string{"infra"},
		env,
	)
	if err == nil {
		t.Fatal("expected error for missing environment argument")
	}
}

func TestParseDeployArgs_TooManyArgs(t *testing.T) {
	env := &environment.Env{}

	_, _, _, _, err := parseDeployArgs(
		[]string{"infra", "development", "extra"},
		env,
	)
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
}

func TestParseDeployArgs_NoArgs(t *testing.T) {
	env := &environment.Env{}

	_, _, _, _, err := parseDeployArgs(
		[]string{},
		env,
	)
	if err == nil {
		t.Fatal("expected error for no arguments")
	}
}

func TestParseDeploySpecificFlags_ComponentEquals(t *testing.T) {
	f, positional, err := parseDeploySpecificFlags([]string{"--component=site"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Component != "site" {
		t.Errorf("expected component 'site', got %q", f.Component)
	}
	if len(positional) != 0 {
		t.Errorf("expected no positional args, got %v", positional)
	}
}

func TestParseDeploySpecificFlags_UnknownFlag(t *testing.T) {
	_, _, err := parseDeploySpecificFlags([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestCommands(t *testing.T) {
	cmds := Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Name() != "deploy" {
		t.Errorf("expected command name 'deploy', got %q", cmds[0].Name())
	}
}

func TestDeployCommandMetadata(t *testing.T) {
	cmd := &deployCommand{}
	meta := cmd.Metadata()
	if meta.CanonicalName != "deploy" {
		t.Errorf("expected canonical name 'deploy', got %q", meta.CanonicalName)
	}
	if meta.Args != "module environment" {
		t.Errorf("expected args 'module environment', got %q", meta.Args)
	}
	if len(meta.Flags) == 0 {
		t.Error("expected flags to be defined")
	}
}
