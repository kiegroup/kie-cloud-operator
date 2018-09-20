package defaults

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadTrialEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if (err != nil) {
			logrus.Error(err.(error))
		}
	}()

	env := GetTrialEnvironment()
	assert.Equal(t, env.Servers[0].DeploymentConfig.ObjectMeta.Name, "trial-env-kieserver")
}

func TestDefaultServer(t *testing.T) {
	object := GetServerObject()
	logrus.Infof("Object is %v", object)
	assert.Equal(t, "default-kieserver", object.DeploymentConfig.Name)
}
