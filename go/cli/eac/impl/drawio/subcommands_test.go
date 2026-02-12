//go:build L1 && ov

package drawio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommandsReturnsAllSubcommands verifies that Commands() returns all six drawio subcommands.
func TestCommandsReturnsAllSubcommands(t *testing.T) {
	cmds := Commands()

	assert.Len(t, cmds, 6, "expected 6 drawio subcommands")

	names := make(map[string]bool)
	for _, cmd := range cmds {
		names[cmd.Name()] = true
	}

	expected := []string{
		"drawio create",
		"drawio decode",
		"drawio embed",
		"drawio encode",
		"drawio info",
		"drawio render",
	}

	for _, name := range expected {
		assert.True(t, names[name], "expected command %q to be registered", name)
	}
}

// TestSubcommandMetadata verifies that each subcommand has required metadata fields.
func TestSubcommandMetadata(t *testing.T) {
	cmds := Commands()

	for _, cmd := range cmds {
		t.Run(cmd.Name(), func(t *testing.T) {
			meta := cmd.Metadata()

			assert.NotEmpty(t, meta.CanonicalName, "CanonicalName should not be empty")
			assert.NotEmpty(t, meta.Short, "Short description should not be empty")
			assert.NotEmpty(t, meta.Long, "Long description should not be empty")
		})
	}
}

// TestDecodeMetadataHasInputFlag verifies decode command declares an input flag.
func TestDecodeMetadataHasInputFlag(t *testing.T) {
	cmd := &drawioDecodeCommand{}
	meta := cmd.Metadata()

	found := false
	for _, flag := range meta.Flags {
		if flag.Name == "input" {
			found = true
			assert.Equal(t, "string", flag.Type)
			assert.Equal(t, "i", flag.Shorthand)
		}
	}
	assert.True(t, found, "decode command should have an 'input' flag")
}

// TestEncodeMetadataHasInputFlag verifies encode command declares an input flag.
func TestEncodeMetadataHasInputFlag(t *testing.T) {
	cmd := &drawioEncodeCommand{}
	meta := cmd.Metadata()

	found := false
	for _, flag := range meta.Flags {
		if flag.Name == "input" {
			found = true
			assert.Equal(t, "string", flag.Type)
		}
	}
	assert.True(t, found, "encode command should have an 'input' flag")
}

// TestEmbedMetadataHasPngFlag verifies embed command declares a required png flag.
func TestEmbedMetadataHasPngFlag(t *testing.T) {
	cmd := &drawioEmbedCommand{}
	meta := cmd.Metadata()

	found := false
	for _, flag := range meta.Flags {
		if flag.Name == "png" {
			found = true
			assert.True(t, flag.Required, "png flag should be required")
		}
	}
	assert.True(t, found, "embed command should have a 'png' flag")
}

// TestCreateMetadataHasOutputFlag verifies create command declares a required output flag.
func TestCreateMetadataHasOutputFlag(t *testing.T) {
	cmd := &drawioCreateCommand{}
	meta := cmd.Metadata()

	found := false
	for _, flag := range meta.Flags {
		if flag.Name == "output" {
			found = true
			assert.True(t, flag.Required, "output flag should be required")
			assert.Equal(t, "o", flag.Shorthand)
		}
	}
	assert.True(t, found, "create command should have an 'output' flag")
}

// TestRenderMetadataHasRequiredFlags verifies render command declares required input and output flags.
func TestRenderMetadataHasRequiredFlags(t *testing.T) {
	cmd := &drawioRenderCommand{}
	meta := cmd.Metadata()

	requiredFlags := map[string]bool{"input": false, "output": false}
	for _, flag := range meta.Flags {
		if _, expected := requiredFlags[flag.Name]; expected {
			requiredFlags[flag.Name] = true
			assert.True(t, flag.Required, "%s flag should be required", flag.Name)
		}
	}
	for name, found := range requiredFlags {
		assert.True(t, found, "render command should have a '%s' flag", name)
	}
}

// TestInfoMetadataHasJsonFlag verifies info command declares a json flag.
func TestInfoMetadataHasJsonFlag(t *testing.T) {
	cmd := &drawioInfoCommand{}
	meta := cmd.Metadata()

	found := false
	for _, flag := range meta.Flags {
		if flag.Name == "json" {
			found = true
			assert.Equal(t, "bool", flag.Type)
		}
	}
	assert.True(t, found, "info command should have a 'json' flag")
}

// TestCanonicalNames verifies canonical names follow the expected pattern.
func TestCanonicalNames(t *testing.T) {
	expected := map[string]string{
		"drawio create": "drawio-create",
		"drawio decode": "drawio-decode",
		"drawio embed":  "drawio-embed",
		"drawio encode": "drawio-encode",
		"drawio info":   "drawio-info",
		"drawio render": "drawio-render",
	}

	cmds := Commands()
	for _, cmd := range cmds {
		t.Run(cmd.Name(), func(t *testing.T) {
			meta := cmd.Metadata()
			assert.Equal(t, expected[cmd.Name()], meta.CanonicalName)
		})
	}
}
