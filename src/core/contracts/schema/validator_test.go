//go:build L0 && ov

package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/src/core/contracts/schema"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getRepoRoot is a test helper that returns the repository root
func getRepoRoot(t *testing.T) string {
	t.Helper()
	repoRoot, err := repository.GetRepositoryRoot("")
	require.NoError(t, err)
	return repoRoot
}

func TestNewValidator(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestValidator_ValidateModulesYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual modules.yml from the repository
	modulesPath := filepath.Join(repoRoot, ".r2r", "eac", "repository", "modules.yml")
	data, err := os.ReadFile(modulesPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaModules, data)
	assert.NoError(t, err, "modules.yml should be valid against schema")
}

func TestValidator_ValidateEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual environments.yml from the repository
	envPath := filepath.Join(repoRoot, ".r2r", "eac", "repository", "environments.yml")
	data, err := os.ReadFile(envPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaEnvironments, data)
	assert.NoError(t, err, "environments.yml should be valid against schema")
}

func TestValidator_ValidateTestingTagsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual testing-tags.yml from the repository
	tagsPath := filepath.Join(repoRoot, ".r2r", "eac", "repository", "testing-tags.yml")
	data, err := os.ReadFile(tagsPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaTestingTags, data)
	assert.NoError(t, err, "testing-tags.yml should be valid against schema")
}

func TestValidator_ValidateTestingTaxonomyYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual testing-taxonomy.yml from the repository
	taxonomyPath := filepath.Join(repoRoot, ".r2r", "eac", "repository", "testing-taxonomy.yml")
	data, err := os.ReadFile(taxonomyPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaTestingTaxonomy, data)
	assert.NoError(t, err, "testing-taxonomy.yml should be valid against schema")
}

func TestValidator_InvalidModulesYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Missing required 'modules' field
	invalidYAML := []byte(`
something_else:
  - name: test
`)

	err = v.ValidateYAML(schema.SchemaModules, invalidYAML)
	assert.Error(t, err, "should fail validation for missing 'modules' field")

	var validErr *schema.ValidationError
	assert.ErrorAs(t, err, &validErr)
	assert.Equal(t, schema.SchemaModules, validErr.SchemaType)
}

func TestValidator_InvalidEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Missing required 'metadata' field
	invalidYAML := []byte(`
environments:
  - moniker: test
    name: Test Environment
    level: L0
    type: unit
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, invalidYAML)
	assert.Error(t, err, "should fail validation for missing 'metadata' field")

	var validErr *schema.ValidationError
	assert.ErrorAs(t, err, &validErr)
	assert.Equal(t, schema.SchemaEnvironments, validErr.SchemaType)
}

func TestValidator_InvalidEnvironmentLevel(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid level value
	invalidYAML := []byte(`
metadata:
  version: "0.1.0"
  description: Test
environments:
  - moniker: test
    name: Test Environment
    level: INVALID
    type: unit
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, invalidYAML)
	assert.Error(t, err, "should fail validation for invalid level")
}

func TestValidator_InvalidModuleMoniker(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid moniker (uppercase not allowed)
	invalidYAML := []byte(`
modules:
  - moniker: INVALID_UPPER
    name: Test Module
`)

	err = v.ValidateYAML(schema.SchemaModules, invalidYAML)
	assert.Error(t, err, "should fail validation for invalid moniker pattern")
}

func TestValidator_ValidMinimalModulesYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Minimal valid modules config
	validYAML := []byte(`
modules:
  - moniker: test-module
    name: Test Module
`)

	err = v.ValidateYAML(schema.SchemaModules, validYAML)
	assert.NoError(t, err, "minimal valid modules config should pass")
}

func TestValidator_ValidMinimalEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Minimal valid environments config
	validYAML := []byte(`
metadata:
  version: "0.1.0"
  description: Test environments
environments:
  - moniker: test-env
    name: Test Environment
    level: L0
    type: unit
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, validYAML)
	assert.NoError(t, err, "minimal valid environments config should pass")
}

func TestValidator_UnknownSchemaType(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaType("unknown"), []byte(`test: value`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown schema type")
}

func TestValidator_InvalidYAMLSyntax(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid YAML syntax
	invalidYAML := []byte(`
modules:
  - moniker: test
    name: [unclosed bracket
`)

	err = v.ValidateYAML(schema.SchemaModules, invalidYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestValidator_ValidateJSON(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	validJSON := []byte(`{
		"modules": [
			{
				"moniker": "test-module",
				"name": "Test Module"
			}
		]
	}`)

	err = v.ValidateJSON(schema.SchemaModules, validJSON)
	assert.NoError(t, err)
}

func TestValidator_InvalidJSON(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	invalidJSON := []byte(`{ invalid json }`)

	err = v.ValidateJSON(schema.SchemaModules, invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestGetSchemaTypes(t *testing.T) {
	types := schema.GetSchemaTypes()
	assert.Len(t, types, 5)
	assert.Contains(t, types, schema.SchemaModules)
	assert.Contains(t, types, schema.SchemaModuleTypes)
	assert.Contains(t, types, schema.SchemaEnvironments)
	assert.Contains(t, types, schema.SchemaTestingTags)
	assert.Contains(t, types, schema.SchemaTestingTaxonomy)
}

func TestValidator_GetSchemaPath(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	schemaPath := v.GetSchemaPath()
	assert.Contains(t, schemaPath, "contracts")
	assert.Contains(t, schemaPath, "src-core")
	assert.Contains(t, schemaPath, schema.ContractVersion)
}
