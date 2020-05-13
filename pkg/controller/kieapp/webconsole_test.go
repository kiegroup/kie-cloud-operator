package kieapp

import (
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/semver"
)

func TestYamlSampleCreation(t *testing.T) {
	reconciler := &Reconciler{Service: test.MockService(), OcpVersion: semver.MajorMinor("v4.2")}
	err := reconciler.createConsoleYAMLSamples()
	assert.NotNil(t, err)
	assert.Equal(t, "console yaml samples not installed, incompatible ocp version", err.Error())
	reconciler.OcpVersion = semver.MajorMinor("v4.3")
	assert.Nil(t, reconciler.createConsoleYAMLSamples())
	reconciler.OcpVersion = semver.MajorMinor("")
	assert.Nil(t, reconciler.createConsoleYAMLSamples())
}
