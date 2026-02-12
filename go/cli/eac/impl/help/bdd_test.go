package help

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	buildcmds "github.com/ready-to-release/eac/go/cli/eac/impl/build"
	commitmessagecmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/commit-message"
	createdesigncmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/design"
	riskassesscmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/risk-assess"
	createriskprofilecmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/risk-profile"
	speccmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/spec"
	squashmessagecmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/squash-message"
	describecmds "github.com/ready-to-release/eac/go/cli/eac/impl/describe"
	drawiocmds "github.com/ready-to-release/eac/go/cli/eac/impl/drawio"
	extensioncmds "github.com/ready-to-release/eac/go/cli/eac/impl/extension"
	getcmds "github.com/ready-to-release/eac/go/cli/eac/impl/get"
	initcmds "github.com/ready-to-release/eac/go/cli/eac/impl/init"
	lintcmds "github.com/ready-to-release/eac/go/cli/eac/impl/lint"
	showhelpcmds "github.com/ready-to-release/eac/go/cli/eac/impl/show/help"
	pipelinecmds "github.com/ready-to-release/eac/go/cli/eac/impl/pipeline"
	pipelinecicmds "github.com/ready-to-release/eac/go/cli/eac/impl/pipeline/ci"
	releasecmds "github.com/ready-to-release/eac/go/cli/eac/impl/release"
	scancmds "github.com/ready-to-release/eac/go/cli/eac/impl/scan"
	scanzapcmds "github.com/ready-to-release/eac/go/cli/eac/impl/scan/zap"
	servecmds "github.com/ready-to-release/eac/go/cli/eac/impl/serve"
	servedesigncmds "github.com/ready-to-release/eac/go/cli/eac/impl/serve/design"
	servegourcecmds "github.com/ready-to-release/eac/go/cli/eac/impl/serve/gource"
	showcmds "github.com/ready-to-release/eac/go/cli/eac/impl/show"
	specsunusedcmds "github.com/ready-to-release/eac/go/cli/eac/impl/specs/unused"
	templatescmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates"
	templatesaicmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/ai"
	templatesclaudecmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/claude"
	templatesdocscmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/docs"
	templatesreportscmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/reports"
	templatesspeccmds "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/specs"
	testcmds "github.com/ready-to-release/eac/go/cli/eac/impl/test"
	aisummarycmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/ai-summary"
	cacheclearcmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/cache-clear"
	updatedesigncmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/design"
	updatedocscmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/docs"
	docsmanifestcmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/docs-manifest"
	evidencecmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/evidence"
	gomodsumscmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/go-mod-sums"
	gotidycmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/go-tidy"
	pdfscreenshotscmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/pdf-screenshots"
	structurizrcmds "github.com/ready-to-release/eac/go/cli/eac/impl/update/structurizr"
	validatecmds "github.com/ready-to-release/eac/go/cli/eac/impl/validate"
	riskcatalogcmds "github.com/ready-to-release/eac/go/cli/eac/impl/validate/risk-catalog"
	validateriskprofilecmds "github.com/ready-to-release/eac/go/cli/eac/impl/validate/risk-profile"
	workcmds "github.com/ready-to-release/eac/go/cli/eac/impl/work"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/show-help," +
			specsBase + "/get-commands",
		DefaultReportName: "cucumber-help",
		RegisterSteps:     registerSteps,
	}

	opts := eacgodog.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: eacgodog.CreateScenarioInitializer(cfg),
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	reg := registry.NewCommandRegistry()
	for _, cmds := range [][]core.CommandPort{
		// This package's commands (help)
		Commands(),

		// Core groups
		getcmds.Commands(),
		showcmds.Commands(),
		validatecmds.Commands(),
		workcmds.Commands(),
		testcmds.Commands(),
		buildcmds.Commands(),
		lintcmds.Commands(),
		scancmds.Commands(),
		releasecmds.Commands(),
		pipelinecmds.Commands(),
		pipelinecicmds.Commands(),
		drawiocmds.Commands(),
		templatescmds.Commands(),

		// Standalone and single-command groups
		initcmds.Commands(),
		describecmds.Commands(),
		showhelpcmds.Commands(),
		extensioncmds.Commands(),
		servecmds.Commands(),
		servedesigncmds.Commands(),
		servegourcecmds.Commands(),
		specsunusedcmds.Commands(),

		// Scan subpackages
		scanzapcmds.Commands(),

		// Create subpackages
		commitmessagecmds.Commands(),
		createdesigncmds.Commands(),
		riskassesscmds.Commands(),
		createriskprofilecmds.Commands(),
		speccmds.Commands(),
		squashmessagecmds.Commands(),

		// Templates install subpackages
		templatesaicmds.Commands(),
		templatesclaudecmds.Commands(),
		templatesdocscmds.Commands(),
		templatesreportscmds.Commands(),
		templatesspeccmds.Commands(),

		// Update subpackages
		aisummarycmds.Commands(),
		cacheclearcmds.Commands(),
		updatedesigncmds.Commands(),
		updatedocscmds.Commands(),
		docsmanifestcmds.Commands(),
		evidencecmds.Commands(),
		gomodsumscmds.Commands(),
		gotidycmds.Commands(),
		pdfscreenshotscmds.Commands(),
		structurizrcmds.Commands(),

		// Validate subpackages
		riskcatalogcmds.Commands(),
		validateriskprofilecmds.Commands(),
	} {
		if err := reg.RegisterAll(cmds...); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to register commands: %v\n", err)
			os.Exit(1)
		}
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)
	os.Exit(m.Run())
}
