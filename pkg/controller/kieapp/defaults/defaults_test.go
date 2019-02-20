package defaults

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
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

	_, err := GetEnvironment(cr, test.MockService())
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

	mockService := test.MockService()
	client := fake.NewFakeClient(cm)
	mockService.GetFunc = func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
		return client.Get(ctx, key, obj)
	}
	_, err := GetEnvironment(cr, mockService)
	assert.Equal(t, fmt.Sprintf("envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Environment, cr.Name), err.Error())
}

func TestInvalidServersConfiguration(t *testing.T) {
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
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 1,
				},
				Servers: []v1.KieServerSet{
					v1.KieServerSet{},
				},
			},
		},
	}

	_, err := GetEnvironment(cr, test.MockService())
	assert.Error(t, err, "invalid spec: provide either server or servers object")
}

func TestMultipleServerDeployment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
		}
	}()
	expectedDeployments := 6
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: expectedDeployments,
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, expectedDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, expectedDeployments-1), env.Servers[expectedDeployments-1].DeploymentConfigs[0].Name)
	assert.Nil(t, err)
}

func TestRHPAMTrialEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 2,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

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
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[2])
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.Exec.Command[2])
}

func TestRHDMTrialEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 2,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	wbServices := env.Console.Services
	mainService := getService(wbServices, "test-rhdmcentr")
	assert.NotNil(t, mainService, "rhdmcentr service not found")
	assert.Len(t, mainService.Spec.Ports, 3, "The rhdmcentr service should have three ports")
	assert.True(t, hasPort(mainService, 8001), "The rhdmcentr service should listen on port 8001")

	pingService := getService(wbServices, "test-rhdmcentr-ping")
	assert.NotNil(t, pingService, "Ping service not found")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.Equal(t, int32(8888), pingService.Spec.Ports[0].Port, "The ping service should listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-drools-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[2])
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-drools-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.Exec.Command[2])
}

func TestRhpamcentrMonitoringEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProduction,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhdmAuthoringHAEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamAuthoringHAEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhdmProdImmutableEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmProductionImmutable,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdImmutableEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestBuildConfiguration(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					v1.KieServerSet{
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/0",
							ArtifactDir:                  "dir0",
							GitSource: v1.GitSource{
								URI:        "http://git.example.com0",
								Reference:  "somebranch0",
								ContextDir: "example0",
							},
							Webhooks: []v1.WebhookSecret{
								{
									Type:   v1.GitHubWebhook,
									Secret: "s3cr3t0",
								},
							},
						},
					},
					v1.KieServerSet{
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.1-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/1",
							ArtifactDir:                  "dir1",
							GitSource: v1.GitSource{
								URI:        "http://git.example.com1",
								Reference:  "somebranch1",
								ContextDir: "example1",
							},
							Webhooks: []v1.WebhookSecret{
								{
									Type:   v1.GitHubWebhook,
									Secret: "s3cr3t1",
								},
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	for i, server := range env.Servers {
		assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
		assert.Equal(t, fmt.Sprintf("http://git.example.com%v", i), server.BuildConfigs[0].Spec.Source.Git.URI)
		assert.Equal(t, fmt.Sprintf("somebranch%v", i), server.BuildConfigs[0].Spec.Source.Git.Ref)
		assert.Equal(t, fmt.Sprintf("example%v", i), server.BuildConfigs[0].Spec.Source.ContextDir)

		assert.Equal(t, fmt.Sprintf("rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.%v-SNAPSHOT", i), server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
		assert.Equal(t, fmt.Sprintf("https://maven.mirror.com/%v", i), server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[1].Value)
		assert.Equal(t, fmt.Sprintf("dir%v", i), server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[2].Value)
		assert.Equal(t, fmt.Sprintf("s3cr3t%v", i), server.BuildConfigs[0].Spec.Triggers[0].GitHubWebHook.Secret)
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
			Environment: v1.RhpamAuthoring,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
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
			Environment: v1.RhpamAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting authoring-ha environment")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)-1), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, v1.Environment{}, env, "Environment should not be empty")
}

func TestConstructConsoleObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	object := shared.ConstructObject(env.Console, &cr.Spec.Objects.Console)
	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", name), object.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructSmartrouterObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	object := shared.ConstructObject(env.Smartrouter, &cr.Spec.Objects.Smartrouter)
	assert.Equal(t, fmt.Sprintf("%s-smartrouter", name), object.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-smartrouter-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Smartrouter.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructServerObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	for i := range env.Servers {
		object := shared.ConstructObject(env.Servers[i], &cr.Spec.Objects.Server.Spec)
		assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", name, i), object.DeploymentConfigs[0].Name)
		re := regexp.MustCompile("[0-9]+")
		assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
		for i := range sampleEnv {
			assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
		}
	}

}

var sampleEnv = []corev1.EnvVar{
	{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	},
	{
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
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Console: v1.KieAppObject{
					Env:       sampleEnv,
					Resources: sampleResources,
				},
				Server: &v1.CommonKieServerSet{
					Spec: v1.KieAppObject{

						Env:       sampleEnv,
						Resources: sampleResources,
					},
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

func TestPartialTemplateConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmAuthoring,
			CommonConfig: v1.CommonConfig{
				AdminPassword: "MyPassword",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting authoring environment")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, "MyPassword", adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Len(t, mavenPassword, 8, "Expected a password with length of 8 to be generated, but got %v", mavenPassword)
}

func getEnvVariable(container corev1.Container, name string) string {
	for _, env := range container.Env {
		if env.Name == name {
			return env.Value
		}
	}
	return ""
}

func TestOverwritePartialTrialPasswords(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			CommonConfig: v1.CommonConfig{
				AdminPassword: "MyPassword",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, "MyPassword", adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Equal(t, "RedHat", mavenPassword, "Expected default password of RedHat, but found %v", mavenPassword)
}

func TestDefaultKieServerNum(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
		},
	}
	_, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, constants.DefaultKieDeployments, cr.Spec.Objects.Server.Deployments, "Default number of kieserver deployments not being set in CR")
	assert.Len(t, cr.Spec.Objects.Servers, 0, "There should be zero custom kieservers being set by default")
}

func TestZeroKieServerDeployments(t *testing.T) {
	deployments := 0
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: deployments,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	for i := 0; i < deployments; i++ {
		kieServerID := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: fmt.Sprintf("test-kieserver-%v", i)}
		assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
	}
	assert.Equal(t, deployments, cr.Spec.Objects.Server.Deployments, "Number of kieserver deployments not set properly in CR")
	assert.Len(t, cr.Spec.Objects.Servers, 0, "There should be zero custom kieservers being set by default")
}
func TestDefaultKieServerID(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: deployments,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	for i := 0; i < deployments; i++ {
		kieServerID := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: fmt.Sprintf("test-kieserver-%v", i)}
		assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
	}
}

func TestSetKieServerID(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					v1.KieServerSet{
						Name: "kieserver-alpha",
					},
					v1.KieServerSet{
						Name: "kieserver-beta",
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	kieServerID := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "kieserver-alpha"}
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
	kieServerID = corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "kieserver-beta"}
	assert.Contains(t, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
}

func TestSetKieServerFrom(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					v1.KieServerSet{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "hello-rules:latest",
						},
					},
					v1.KieServerSet{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "bye-rules:latest",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, "hello-rules:latest", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "bye-rules:latest", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestSetKieServerFromBuild(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					v1.KieServerSet{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "hello-rules:latest",
						},
					},
					v1.KieServerSet{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "bye-rules:latest",
						},
						Build: &v1.KieAppBuildObject{},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, "hello-rules:latest", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "test-kieserver-1:latest", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestMultipleBuildConfigurations(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
							GitSource: v1.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "test",
							},
							Webhooks: []v1.WebhookSecret{
								{
									Type:   v1.GitHubWebhook,
									Secret: "s3cr3t",
								},
							},
							From: &corev1.ObjectReference{
								Kind:      "ImageStreamTag",
								Name:      "custom-kieserver",
								Namespace: "",
							},
						},
					},
					{
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
							GitSource: v1.GitSource{
								URI:        "http://git.example.com",
								Reference:  "anotherbranch",
								ContextDir: "test",
							},
							Webhooks: []v1.WebhookSecret{
								{
									Type:   v1.GitHubWebhook,
									Secret: "s3cr3t",
								},
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 2, "Expect two KIE Servers to be created based on provided build configs")
	assert.Equal(t, "somebranch", env.Servers[0].BuildConfigs[0].Spec.Source.Git.Ref)
	assert.Equal(t, "anotherbranch", env.Servers[1].BuildConfigs[0].Spec.Source.Git.Ref)

	assert.Equal(t, "ImageStreamTag", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Kind)
	assert.Equal(t, "custom-kieserver", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Namespace)

	assert.Equal(t, "ImageStreamTag", env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Kind)
	imgName := fmt.Sprintf("rhdm%v-kieserver-openshift:%v", cr.Spec.CommonConfig.Version, cr.Spec.CommonConfig.ImageTag)
	assert.Equal(t, imgName, env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "openshift", env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Namespace)
}
