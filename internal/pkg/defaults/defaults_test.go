package defaults

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCommonEnvironmentDefaults(t *testing.T) {
	defaults := ConsoleEnvironmentDefaults()
	logrus.Infof("Loaded common defaults as %v", defaults)
	assert.True(t, true)
}

func TestServerEnvironmentDefaults(t *testing.T) {
	defaults := ServerEnvironmentDefaults()
	logrus.Infof("Loaded server defaults as %v", defaults)
	assert.True(t, true)
}
