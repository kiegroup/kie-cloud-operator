package kieapp

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	appv1alpha1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnknownEnvironmentObjects(t *testing.T) {
	cr := &appv1alpha1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: appv1alpha1.KieAppSpec{
			Environment: "unknown",
		},
	}

	env, common, err := defaults.GetEnvironment(cr)
	assert.NotNil(t, err)
	env = ConsolidateObjects(env, common, cr)

	logrus.Debugf("Testing with environment %v", cr.Spec.Environment)
	assert.Equal(t, appv1alpha1.Environment{}, env, "Env object should be empty")
	assert.Equal(t, fmt.Sprintf("envs/%s.yaml does not exist, environment not deployed", cr.Spec.Environment), err.Error())
}

func TestTrialConsoleEnv(t *testing.T) {
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	}
	envAddition := corev1.EnvVar{
		Name:  "CONSOLE_TEST",
		Value: "test",
	}
	cr := &appv1alpha1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appv1alpha1.KieAppSpec{
			Environment: "trial",
			Objects: appv1alpha1.KieAppObjects{
				Console: appv1alpha1.KieAppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}

	env, common, err := defaults.GetEnvironment(cr)
	if !assert.Nil(t, err, "error should be nil") {
		logrus.Error(err)
	}
	env = ConsolidateObjects(env, common, cr)

	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", cr.Name), env.Console.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
}

func TestTrialServerEnv(t *testing.T) {
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	}
	envAddition := corev1.EnvVar{
		Name:  "SERVER_TEST",
		Value: "test",
	}
	commonAddition := corev1.EnvVar{
		Name:  "COMMON_TEST",
		Value: "test",
	}
	cr := &appv1alpha1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appv1alpha1.KieAppSpec{
			Environment:    "trial",
			KieDeployments: 3,
			Objects: appv1alpha1.KieAppObjects{
				Server: appv1alpha1.KieAppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}

	env, common, err := defaults.GetEnvironment(cr)
	if !assert.Nil(t, err, "error should be nil") {
		logrus.Error(err)
	}
	common.Objects.Server.Env = append(common.Objects.Server.Env, commonAddition)
	env = ConsolidateObjects(env, common, cr)

	assert.Equal(t, cr.Spec.KieDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, cr.Spec.KieDeployments-1), env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
}
