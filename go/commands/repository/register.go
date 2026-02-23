// Package repository aggregates command providers for repository-scoped CLI commands.
// It registers itself as a command provider using the database/sql driver pattern:
// importing this package causes its init() to register all commands with the global registry.
package repository

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/registry"

	createcommitmessage "github.com/ready-to-release/eac/go/commands/repository/create/commit-message"
	createdesign "github.com/ready-to-release/eac/go/commands/repository/create/design"
	createriskassess "github.com/ready-to-release/eac/go/commands/repository/create/risk-assess"
	createriskprofile "github.com/ready-to-release/eac/go/commands/repository/create/risk-profile"
	createspec "github.com/ready-to-release/eac/go/commands/repository/create/spec"
	createsquashmessage "github.com/ready-to-release/eac/go/commands/repository/create/squash-message"
	"github.com/ready-to-release/eac/go/commands/repository/describe"
	"github.com/ready-to-release/eac/go/commands/repository/drawio"
	"github.com/ready-to-release/eac/go/commands/repository/extension"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/commands/repository/help"
	initcmd "github.com/ready-to-release/eac/go/commands/repository/init"
	"github.com/ready-to-release/eac/go/commands/repository/pipeline"
	pipelineci "github.com/ready-to-release/eac/go/commands/repository/pipeline/ci"
	"github.com/ready-to-release/eac/go/commands/repository/release"
	"github.com/ready-to-release/eac/go/commands/repository/serve"
	servedesign "github.com/ready-to-release/eac/go/commands/repository/serve/design"
	servegource "github.com/ready-to-release/eac/go/commands/repository/serve/gource"
	"github.com/ready-to-release/eac/go/commands/repository/show"
	showhelp "github.com/ready-to-release/eac/go/commands/repository/show/help"
	specsunused "github.com/ready-to-release/eac/go/commands/repository/specs/unused"
	"github.com/ready-to-release/eac/go/commands/repository/templates"
	templatesai "github.com/ready-to-release/eac/go/commands/repository/templates/install/ai"
	templatesclaude "github.com/ready-to-release/eac/go/commands/repository/templates/install/claude"
	templatesdocs "github.com/ready-to-release/eac/go/commands/repository/templates/install/docs"
	templatesreports "github.com/ready-to-release/eac/go/commands/repository/templates/install/reports"
	templatesspecs "github.com/ready-to-release/eac/go/commands/repository/templates/install/specs"
	updateaisummary "github.com/ready-to-release/eac/go/commands/repository/update/ai-summary"
	updatedependabot "github.com/ready-to-release/eac/go/commands/repository/update/dependabot"
	updatecacheclear "github.com/ready-to-release/eac/go/commands/repository/update/cache-clear"
	updatedesign "github.com/ready-to-release/eac/go/commands/repository/update/design"
	updatedocsmanifest "github.com/ready-to-release/eac/go/commands/repository/update/docs-manifest"
	updategomodsums "github.com/ready-to-release/eac/go/commands/repository/update/go-mod-sums"
	updategotidy "github.com/ready-to-release/eac/go/commands/repository/update/go-tidy"
	updatepdfscreenshots "github.com/ready-to-release/eac/go/commands/repository/update/pdf-screenshots"
	updatestructurizr "github.com/ready-to-release/eac/go/commands/repository/update/structurizr"
	"github.com/ready-to-release/eac/go/commands/repository/validate"
	validateriskcatalog "github.com/ready-to-release/eac/go/commands/repository/validate/risk-catalog"
	validateriskprofile "github.com/ready-to-release/eac/go/commands/repository/validate/risk-profile"
	"github.com/ready-to-release/eac/go/commands/repository/work"
)

func init() {
	registry.RegisterProvider(Commands)
}

// Commands returns all command ports from all repository command packages.
func Commands() []core.CommandPort {
	var all []core.CommandPort
	for _, cmds := range [][]core.CommandPort{
		// Core groups
		get.Commands(),
		show.Commands(),
		validate.Commands(),
		work.Commands(),
		release.Commands(),
		pipeline.Commands(),
		pipelineci.Commands(),
		drawio.Commands(),
		templates.Commands(),
		help.Commands(),

		// Standalone and single-command groups
		initcmd.Commands(),
		describe.Commands(),
		showhelp.Commands(),
		extension.Commands(),
		serve.Commands(),
		servedesign.Commands(),
		servegource.Commands(),
		specsunused.Commands(),

		// Create subpackages
		createcommitmessage.Commands(),
		createdesign.Commands(),
		createriskassess.Commands(),
		createriskprofile.Commands(),
		createspec.Commands(),
		createsquashmessage.Commands(),

		// Templates install subpackages
		templatesai.Commands(),
		templatesclaude.Commands(),
		templatesdocs.Commands(),
		templatesreports.Commands(),
		templatesspecs.Commands(),

		// Update subpackages
		updateaisummary.Commands(),
		updatedependabot.Commands(),
		updatecacheclear.Commands(),
		updatedesign.Commands(),
		updatedocsmanifest.Commands(),
		updategomodsums.Commands(),
		updategotidy.Commands(),
		updatepdfscreenshots.Commands(),
		updatestructurizr.Commands(),

		// Validate subpackages
		validateriskcatalog.Commands(),
		validateriskprofile.Commands(),
	} {
		all = append(all, cmds...)
	}
	return all
}
