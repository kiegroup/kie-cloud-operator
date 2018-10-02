package defaults

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoadTrialEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err.(error))
		}
	}()

	env, err := GetEnvironment("trial")
	assert.Equal(t, env.Servers[0].DeploymentConfigs[0].ObjectMeta.Name, "trial-env-kieserver")
	assert.Nil(t, err)

	_, err = GetEnvironment("fdsfsd")
	assert.NotNil(t, err)
}

func TestDefaultConsole(t *testing.T) {
	object := GetConsoleObject()
	logrus.Infof("Object is %v", object)
	assert.Equal(t, "console-rhpamcentr", object.DeploymentConfigs[0].Name)
}

func TestDefaultServer(t *testing.T) {
	object := GetServerObject()
	logrus.Infof("Object is %v", object)
	assert.Equal(t, "default-kieserver", object.DeploymentConfigs[0].Name)
}
