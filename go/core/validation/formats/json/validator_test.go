package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator_InvalidSchema(t *testing.T) {
	_, err := NewValidator("/non/existent/schema.json")
	assert.Error(t, err, "should fail for non-existent schema")
}

func TestVerifyImplementation_NilSchema(t *testing.T) {
	v := &Validator{schemaPath: "test.json"}
	errs := v.VerifyImplementation()
	assert.Len(t, errs, 1, "should report error when schema is nil")
}

func TestPathToFileURL(t *testing.T) {
	url := pathToFileURL("/tmp/schema.json")
	assert.Contains(t, url, "file://")
	assert.Contains(t, url, "schema.json")
}
