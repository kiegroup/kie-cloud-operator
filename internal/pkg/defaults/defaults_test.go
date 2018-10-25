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
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Name)
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
			Environment:    "trial",
			KieDeployments: 6,
		},
	}

	env, _, err := GetEnvironment(cr)
	assert.Equal(t, cr.Spec.KieDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, cr.Spec.KieDeployments-1), env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Name)
	assert.Nil(t, err)
}
