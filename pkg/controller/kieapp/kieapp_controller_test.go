package kieapp

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestKieAppDefaults(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
			Objects: v1.KieAppObjects{
				Server: v1.KieAppObject{},
			},
		},
	}
	assert.Nil(t, cr.Spec.Objects.Server.Env)
	assert.NotContains(t, cr.Spec.Objects.Console.Env, corev1.EnvVar{
		Name: "empty",
	})
}

func TestUnknownEnvironmentObjects(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "unknown",
		},
	}

	env, err := defaults.GetEnvironment(cr, test.MockService())
	assert.Equal(t, fmt.Sprintf("envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Environment, cr.Name), err.Error())

	env = ConsolidateObjects(env, cr)
	assert.NotNil(t, err)

	log.Debug("Testing with environment ", cr.Spec.Environment)
	assert.Equal(t, v1.Environment{}, env, "Env object should be empty")
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
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
			Objects: v1.KieAppObjects{
				Console: v1.KieAppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}

	env, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env = ConsolidateObjects(env, cr)

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
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment:    "trial",
			KieDeployments: 3,
			Objects: v1.KieAppObjects{
				Server: v1.KieAppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}

	env, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env = append(env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition)
	env = ConsolidateObjects(env, cr)

	assert.Equal(t, cr.Spec.KieDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, cr.Spec.KieDeployments-1), env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Name)
	pattern := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(pattern.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
	assert.Contains(t, env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
}

func TestRhpamRegistry(t *testing.T) {
	registry1 := "registry1.test.com"
	os.Setenv("REGISTRY", registry1)
	defer os.Unsetenv("REGISTRY")
	os.Setenv("INSECURE", "true")
	defer os.Unsetenv("INSECURE")
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
		},
	}
	_, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, registry1, cr.Spec.RhpamRegistry.Registry)
	assert.Equal(t, true, cr.Spec.RhpamRegistry.Insecure)

	registry2 := "registry2.test.com:5000"
	cr2 := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
			RhpamRegistry: v1.KieAppRegistry{
				Registry: registry2,
			},
		},
	}
	_, err = defaults.GetEnvironment(cr2, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, registry2, cr2.Spec.RhpamRegistry.Registry)
	assert.Equal(t, false, cr2.Spec.RhpamRegistry.Insecure)
}

func TestGenerateSecret(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
		},
	}
	env, err := defaults.GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting a new environment")
	assert.Len(t, env.Console.Secrets, 0, "No secret is available when reading the trial workbench from yaml files")

	scheme, err := v1.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := &KieAppReconciler{mockService}
	env, _, err = reconciler.NewEnv(cr)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Len(t, env.Console.Secrets, 1, "One secret should be generated for the trial workbench")
}

func TestConsoleHost(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
		},
	}

	scheme, err := v1.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := &KieAppReconciler{mockService}
	_, _, err = reconciler.NewEnv(cr)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Equal(t, fmt.Sprintf("http://%s", cr.Name), cr.Status.ConsoleHost, "spec.commonConfig.consoleHost should be URL from the resulting workbench route host")
}

func TestMergeTrialAndCommonConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
		},
	}
	env, err := defaults.GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	// HTTP Routes are added
	assert.Equal(t, 2, len(env.Console.Routes), "Expected 2 routes. rhpamcentr (http + https)")
	assert.Equal(t, 2, len(env.Servers[0].Routes), "Expected 2 routes. kieserver[0] (http + https)")

	assert.Equal(t, "test-rhpamcentr", env.Console.Routes[0].Name)
	assert.Equal(t, "test-rhpamcentr-http", env.Console.Routes[1].Name)

	assert.Equal(t, "test-kieserver-0", env.Servers[0].Routes[0].Name)
	assert.Equal(t, "test-kieserver-0-http", env.Servers[0].Routes[1].Name)

	// Env vars overrides
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_SERVER_PROTOCOL",
		Value: "",
	})

	// H2 Volumes are mounted
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      "test-h2-pvol",
		MountPath: "/opt/eap/standalone/data",
	})
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "test-h2-pvol",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}
