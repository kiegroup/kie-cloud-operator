package defaults

import (
	"fmt"
	"testing"

	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoadTrialEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err.(error))
		}
	}()

	cr := &opv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: opv1.AppSpec{
			Environment: "trial",
		},
	}

	env, _, err := GetEnvironment(cr)
	assert.Equal(t, fmt.Sprintf("%s-kieserver-0", cr.Name), env.Servers[0].DeploymentConfigs[0].Name)
	assert.Nil(t, err)
}

func TestLoadUnknownEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err.(error))
		}
	}()

	cr := &opv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: opv1.AppSpec{
			Environment: "unknown",
		},
	}

	_, _, err := GetEnvironment(cr)
	assert.NotNil(t, err)
}
func TestMultipleServerDeployment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err.(error))
		}
	}()

	cr := &opv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: opv1.AppSpec{
			Environment:   "trial",
			NumKieServers: 2,
		},
	}

	env, _, err := GetEnvironment(cr)
	assert.Equal(t, fmt.Sprintf("%s-kieserver-1", cr.Name), env.Servers[1].DeploymentConfigs[0].Name)
	assert.Nil(t, err)
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
