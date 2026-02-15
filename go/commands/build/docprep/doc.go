// Package docprep provides documentation preprocessing for MkDocs builds.
//
// It transforms source documentation (markdown, assets, diagrams) into a
// staging directory ready for MkDocs to render into HTML or PDF.
// All preprocessing is pure Go; containers only run mkdocs build.
//
// Usage:
//
//	mode := docprep.ModeFromString("site")
//	pctx := docprep.NewPreprocessContext(ctx, book, workspaceRoot, stagingDir, logWriter, mode)
//	pipeline := docprep.DefaultPipeline()
//	if err := pipeline.Execute(pctx); err != nil { ... }
package docprep
