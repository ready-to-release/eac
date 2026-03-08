// Usage: serve docs [flags]
package docs

import (
	"context"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/serve"
)

type serveDocsCommand struct{}

var _ core.SimpleCommandPort = (*serveDocsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&serveDocsCommand{},
	}
}

func (c *serveDocsCommand) Name() string { return "serve docs" }

func (c *serveDocsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "serve-docs",
		Short:         "Start or stop MkDocs documentation server",
		Long:          "Starts a Docker container serving the documentation site with live reload.\nEdits to markdown files in docs/ are reflected immediately.\nUse --stop to stop the running server.",
		Flags:         serve.DocsFlags(),
	}
}

func (c *serveDocsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return serve.Serve()
}
