package defaults

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLoadUnknownEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
		}
	}()

	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "unknown",
		},
	}

	_, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Equal(t, fmt.Sprintf("envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Environment, cr.Name), err.Error())
}

func TestInaccessibleConfigMap(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "map-test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "na",
		},
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "test-ns",
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}

	client := fake.NewFakeClient(cm)
	_, err := GetEnvironment(cr, client)
	assert.Equal(t, fmt.Sprintf("%s/%s ConfigMap not yet accessible, '%s' KieApp not deployed. Retrying... ", cr.Namespace, constants.ConfigMapPrefix, cr.Name), err.Error())
}

func TestMultipleServerDeployment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
		}
	}()

	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment:    "trial",
			KieDeployments: 6,
		},
	}

	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Equal(t, cr.Spec.KieDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, cr.Spec.KieDeployments-1), env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Name)
	assert.Nil(t, err)
}

func TestTrialEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment:    "trial",
			KieDeployments: 2,
		},
	}
	env, err := GetEnvironment(cr, fake.NewFakeClient())

	assert.Nil(t, err, "Error getting trial environment")
	wbServices := env.Console.Services
	mainService := getService(wbServices, "test-rhpamcentr")
	assert.NotNil(t, mainService, "rhpamcentr service not found")
	assert.Len(t, mainService.Spec.Ports, 3, "The rhpamcentr service should have three ports")
	assert.True(t, hasPort(mainService, 8001), "The rhpamcentr service should listen on port 8001")

	pingService := getService(wbServices, "test-rhpamcentr-ping")
	assert.NotNil(t, pingService, "Ping service not found")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.Equal(t, int32(8888), pingService.Spec.Ports[0].Port, "The ping service should listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam72-businesscentral-openshift", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamcentrMonitoringEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment:    "production",
			KieDeployments: 2,
		},
	}
	env, err := GetEnvironment(cr, fake.NewFakeClient())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam72-businesscentral-monitoring-openshift", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestBuildConfiguration(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "immutable-kieserver",
			Objects: v1.KieAppObjects{
				Build: v1.KieAppBuildObject{
					KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
					GitSource: v1.GitSource{
						URI:        "http://git.example.com",
						Reference:  "somebranch",
						ContextDir: "test",
					},
					Webhooks: []v1.WebhookSecret{
						v1.WebhookSecret{
							Type:   v1.GitHubWebhook,
							Secret: "s3cr3t",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, fake.NewFakeClient())

	assert.Nil(t, err, "Error getting prod environment")
	for _, server := range env.Servers {
		assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
		assert.Equal(t, "http://git.example.com", server.BuildConfigs[0].Spec.Source.Git.URI)
		assert.Equal(t, "somebranch", server.BuildConfigs[0].Spec.Source.Git.Ref)
		assert.Equal(t, "test", server.BuildConfigs[0].Spec.Source.ContextDir)

		assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)

		assert.Equal(t, "s3cr3t", server.BuildConfigs[0].Spec.Triggers[0].GitHubWebHook.Secret)
		assert.NotEmpty(t, server.BuildConfigs[0].Spec.Triggers[1].GenericWebHook.Secret)
	}
}

func getService(services []corev1.Service, name string) corev1.Service {
	for _, service := range services {
		if service.Name == name {
			return service
		}
	}
	return corev1.Service{}
}

func hasPort(service corev1.Service, portNum int32) bool {
	for _, port := range service.Spec.Ports {
		if port.Port == portNum {
			return true
		}
	}
	return false
}

func TestAuthoringEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment:    "authoring",
			KieDeployments: 3,
		},
	}
	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err, "Error getting authoring environment")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, v1.Environment{}, env, "Environment should not be empty")
}

func TestAuthoringHAEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment:    "authoring-ha",
			KieDeployments: 3,
		},
	}
	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err, "Error getting authoring-ha environment")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, v1.Environment{}, env, "Environment should not be empty")
}

func TestConstructConsoleObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err)

	object := shared.ConstructObject(env.Console, &cr.Spec.Objects.Console)
	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", name), object.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructSmartrouterObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err)

	object := shared.ConstructObject(env.Smartrouter, &cr.Spec.Objects.Smartrouter)
	assert.Equal(t, fmt.Sprintf("%s-smartrouter", name), object.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-smartrouter-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Smartrouter.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructServerObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err)

	for i := range env.Servers {
		object := shared.ConstructObject(env.Servers[i], &cr.Spec.Objects.Server)
		assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", name, i), object.DeploymentConfigs[0].Name)
		re := regexp.MustCompile("[0-9]+")
		assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
		for i := range sampleEnv {
			assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
		}
	}

}

var sampleEnv = []corev1.EnvVar{
	corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	},
	corev1.EnvVar{
		Name:  "TEST_VAR",
		Value: "test",
	},
}

var sampleResources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		"memory": *resource.NewQuantity(1, "Mi"),
	},
}

func buildKieApp(name string) *v1.KieApp {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "trial",
			Objects: v1.KieAppObjects{
				Console: v1.KieAppObject{
					Env:       sampleEnv,
					Resources: sampleResources,
				},
				Server: v1.KieAppObject{
					Env:       sampleEnv,
					Resources: sampleResources,
				},
				Smartrouter: v1.KieAppObject{
					Env:       sampleEnv,
					Resources: sampleResources,
				},
			},
		},
	}
	return cr
}
