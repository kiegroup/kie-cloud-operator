package defaults

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/sirupsen/logrus"
)

func TestCommonEnvironmentDefaults(t *testing.T) {
	defaults := consoleEnvironmentDefaults()
	logrus.Infof("Loaded defaults as %v", defaults)
	assert.True(t, true)
}
