package ui

import (
	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidSchema(t *testing.T) {
	assert.NotEqual(t, spec.Schema{}, getSchema(), "Empty schema loaded")
}

func TestValidApiVersion(t *testing.T) {
	assert.NotEqual(t, "", getApiVersion(), "Empty API Version loaded")
}

func TestValidObjectKind(t *testing.T) {
	assert.NotEqual(t, "", getObjectKind(), "Empty object kind loaded")
}

func TestValidForm(t *testing.T) {
	assert.NotEqual(t, web.Form{}, getForm(), "Empty form loaded")
}
