package defaults

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
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

func TestMultipleServerDeployment(t *testing.T) {
	deployments := 6
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
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, deployments), env.Servers[deployments-1].DeploymentConfigs[0].Name)
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, deployments), cr.Spec.Objects.Servers[deployments-1].Name)
	assert.Nil(t, err)
}

func TestRHPAMTrialEnvironment(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
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
	assert.False(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[2])
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.Exec.Command[2])
}

func TestRHDMTrialEnvironment(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhdmTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
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
	assert.False(t, hasPort(pingService, 8888), "The ping service should not listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[2])
	assert.Equal(t, "curl --fail --silent -u \"${KIE_ADMIN_USER}\":\"${KIE_ADMIN_PWD}\" http://localhost:8080/kie-wb.jsp", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.Exec.Command[2])
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

	pingService := getService(env.Console.Services, "test-rhpamcentrmon-ping")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.True(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")
	for i := 0; i < len(env.Servers); i++ {
		assert.Equal(t, "PRODUCTION", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_MODE"))
	}
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
	assert.Equal(t, "test-amq", env.Others[0].StatefulSets[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < len(env.Servers); i++ {
		assert.Equal(t, "DEVELOPMENT", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_MODE"))
	}
}

func TestRhpamAuthoringHAEnvironment(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamAuthoringHA,
			CommonConfig: v1.CommonConfig{
				AMQPassword:        "amq",
				AMQClusterPassword: "cluster",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-amq", env.Others[0].StatefulSets[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	amqClusterPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_JMS_BROKER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	amqPassword := getEnvVariable(env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0], "AMQ_PASSWORD")
	assert.Equal(t, "amq", amqPassword, "Expected provided password to take effect, but found %v", amqPassword)
	amqClusterPassword = getEnvVariable(env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0], "AMQ_CLUSTER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	pingService := getService(env.Console.Services, "test-rhpamcentr-ping")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.True(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")

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

func TestInvalidBuildConfiguration(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(2),
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: v1.GitSource{
								URI:       "http://git.example.com",
								Reference: "somebranch",
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
	_, err := GetEnvironment(cr, test.MockService())
	assert.NotNil(t, err, "Expected error trying to deploy multiple builds of same type")
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
					{
						Build: &v1.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: v1.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "example",
							},
							Webhooks: []v1.WebhookSecret{
								{
									Type:   v1.GitHubWebhook,
									Secret: "s3cr3t",
								},
							},
						},
					},
					{
						From: &corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      "test",
							Namespace: "other-ns",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, 2, len(env.Servers))
	assert.Equal(t, "example", env.Servers[0].BuildConfigs[0].Spec.Source.ContextDir)

	// Server #0
	server := env.Servers[0]
	assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
	assert.Equal(t, "http://git.example.com", server.BuildConfigs[0].Spec.Source.Git.URI)
	assert.Equal(t, "somebranch", server.BuildConfigs[0].Spec.Source.Git.Ref)

	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
	assert.Equal(t, "https://maven.mirror.com/", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[1].Value)
	assert.Equal(t, "dir", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[2].Value)
	assert.Equal(t, "s3cr3t", server.BuildConfigs[0].Spec.Triggers[0].GitHubWebHook.Secret)
	assert.NotEmpty(t, server.BuildConfigs[0].Spec.Triggers[1].GenericWebHook.Secret)

	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, "test-kieserver:latest", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)

	// Server #1
	server = env.Servers[1]
	assert.Empty(t, server.ImageStreams)
	assert.Empty(t, server.BuildConfigs)
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, "test", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "other-ns", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
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
			CommonConfig: v1.CommonConfig{
				DBPassword: "Database",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting authoring environment")
	dbPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
	assert.Equal(t, "Database", dbPassword, "Expected provided password to take effect, but found %v", dbPassword)
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Spec.CommonConfig.ApplicationName), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, string(appsv1.DeploymentStrategyTypeRecreate), string(env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Strategy.Type), "The DC should use a Recreate strategy when using the H2 DB")
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Spec.CommonConfig.ApplicationName), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, v1.Environment{}, env, "Environment should not be empty")
}

func TestConstructConsoleObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name, 1)
	cr.Spec.Objects.Console.Replicas = Pint32(3)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	env = ConsolidateObjects(env, cr)
	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", name), env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, int32(1), env.Console.DeploymentConfigs[0].Spec.Replicas)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructSmartRouterObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name, 1)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	env = ConsolidateObjects(env, cr)
	assert.Equal(t, fmt.Sprintf("%s-smartrouter", name), env.SmartRouter.DeploymentConfigs[0].Name)
	assert.Equal(t, int32(1), env.SmartRouter.DeploymentConfigs[0].Spec.Replicas)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-smartrouter-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.SmartRouter.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructServerObject(t *testing.T) {
	name := "test"
	{
		cr := buildKieApp(name, 1)
		env, err := GetEnvironment(cr, test.MockService())
		assert.Nil(t, err)

		env = ConsolidateObjects(env, cr)
		assert.Equal(t, fmt.Sprintf("%s-kieserver", name), env.Servers[0].DeploymentConfigs[0].Name)
		assert.Equal(t, int32(1), env.Servers[0].DeploymentConfigs[0].Spec.Replicas)
		re := regexp.MustCompile("[0-9]+")
		assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
		for i := range sampleEnv {
			assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
		}
	}
	{
		cr := buildKieApp(name, 3)
		env, err := GetEnvironment(cr, test.MockService())
		assert.Nil(t, err)

		env = ConsolidateObjects(env, cr)
		for i, s := range env.Servers {
			if i == 0 {
				assert.Equal(t, fmt.Sprintf("%s-kieserver", name), s.DeploymentConfigs[0].Name)
			} else {
				assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", name, i+1), s.DeploymentConfigs[0].Name)
			}
			re := regexp.MustCompile("[0-9]+")
			assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
			for i := range sampleEnv {
				assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
			}
		}
	}
}

func TestSetReplicas(t *testing.T) {
	name := "test"
	replicas := Pint32(8)
	cr := buildKieApp(name, 3)
	cr.Spec.Objects.Console.Replicas = replicas
	cr.Spec.Objects.SmartRouter.Replicas = replicas
	for i := range cr.Spec.Objects.Servers {
		cr.Spec.Objects.Servers[i].Replicas = replicas
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	env = ConsolidateObjects(env, cr)
	assert.Equal(t, int32(1), env.Console.DeploymentConfigs[0].Spec.Replicas, "Replicas scaling should be denied and use default instead")
	assert.Equal(t, *replicas, env.SmartRouter.DeploymentConfigs[0].Spec.Replicas)
	for i, s := range env.Servers {
		if i == 0 {
			assert.Equal(t, fmt.Sprintf("%s-kieserver", name), s.DeploymentConfigs[0].Name)
		} else {
			assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", name, i+1), s.DeploymentConfigs[0].Name)
		}
		re := regexp.MustCompile("[0-9]+")
		assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
		assert.Equal(t, *replicas, s.DeploymentConfigs[0].Spec.Replicas)
		for i := range sampleEnv {
			assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
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

	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, fmt.Sprintf("envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Environment, cr.Name), err.Error())

	env = ConsolidateObjects(env, cr)
	assert.NotNil(t, err)

	log.Debug("Testing with environment ", cr.Spec.Environment)
	assert.Equal(t, v1.Environment{}, env, "Env object should be empty")
}

func TestTrialServerEnv(t *testing.T) {
	deployments := 6
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "replaced",
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
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						SecuredKieAppObject: v1.SecuredKieAppObject{
							KieAppObject: v1.KieAppObject{
								Env: []corev1.EnvVar{
									envReplace,
									envAddition,
								},
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env = append(env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition)
	env = ConsolidateObjects(env, cr)

	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, deployments), env.Servers[deployments-1].DeploymentConfigs[0].Name)
	pattern := regexp.MustCompile("[0-9]+")
	expectedISTagName := fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(pattern.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag)
	assert.Equal(t, expectedISTagName, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "replaced",
	})
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
}

func TestTrialServersEnv(t *testing.T) {
	deployments := 3
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "replaced",
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
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Name: "server-a",
						SecuredKieAppObject: v1.SecuredKieAppObject{
							KieAppObject: v1.KieAppObject{
								Env: []corev1.EnvVar{
									envReplace,
									envAddition,
									commonAddition,
								},
							},
						},
						Deployments: Pint(1),
					},
					{
						Name:        "server-b",
						Deployments: Pint(deployments),
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env = ConsolidateObjects(env, cr)

	assert.Len(t, env.Servers, 4)
	pattern := regexp.MustCompile("[0-9]+")
	expectedISTagName := fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(pattern.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag)
	for index := 0; index < 1; index++ {
		s := env.Servers[index]
		assert.Equal(t, expectedISTagName, s.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
		assert.Equal(t, cr.Spec.Objects.Servers[0].Name, s.DeploymentConfigs[0].Name)
		assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
		assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
		assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "KIE_ADMIN_PWD",
			Value: "replaced",
		})
		assert.Contains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
	}
	for index := 1; index < 1+deployments; index++ {
		s := env.Servers[index]
		assert.NotContains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
		assert.NotContains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
		assert.NotContains(t, s.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	}
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
			Environment: "rhdm-trial",
			CommonConfig: v1.CommonConfig{
				ApplicationName: "trial",
			},
			Objects: v1.KieAppObjects{
				Console: v1.SecuredKieAppObject{
					KieAppObject: v1.KieAppObject{
						Env: []corev1.EnvVar{
							envReplace,
							envAddition,
						},
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env = ConsolidateObjects(env, cr)

	assert.Equal(t, fmt.Sprintf("%s-rhdmcentr", cr.Spec.CommonConfig.ApplicationName), env.Console.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhdm%s-decisioncentral-openshift:%s", strings.Join(re.FindAllString(constants.ProductVersion, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
}

func TestKieAppDefaults(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects:     v1.KieAppObjects{},
		},
	}
	assert.NotContains(t, cr.Spec.Objects.Console.Env, corev1.EnvVar{
		Name: "empty",
	})
}

func TestMergeTrialAndCommonConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	// HTTP Routes are added
	assert.Equal(t, 2, len(env.Console.Routes), "Expected 2 routes. rhpamcentr (http + https)")
	assert.Equal(t, 2, len(env.Servers[0].Routes), "Expected 2 routes. kieserver[0] (http + https)")

	assert.Equal(t, "test-rhpamcentr", env.Console.Routes[0].Name)
	assert.Equal(t, "test-rhpamcentr-http", env.Console.Routes[1].Name)

	assert.Equal(t, "test-kieserver", env.Servers[0].Routes[0].Name)
	assert.Equal(t, "test-kieserver-http", env.Servers[0].Routes[1].Name)

	// Env vars overrides
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_SERVER_PROTOCOL",
		Value: "",
	})

	// H2 Volumes are mounted
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      "test-kieserver-h2-pvol",
		MountPath: "/opt/eap/standalone/data",
	})
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "test-kieserver-h2-pvol",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}

func TestServerConflict(t *testing.T) {
	deployments := 2
	name := "test"
	duplicate := "testing"
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Name: duplicate, Deployments: Pint(deployments)},
					{Name: duplicate},
				},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("duplicate kieserver name %s", duplicate))
}

func TestServerConflictGenerated(t *testing.T) {
	deployments := 2
	name := "test"
	duplicate := "test-kieserver-2"
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Name: duplicate},
					{Deployments: Pint(deployments)},
				},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.NotNil(t, err)
	assert.Error(t, err, fmt.Sprintf("duplicate kieserver name %s", duplicate))
}

// test-kieserver | test-kieserver-2 | test-kieserver-3
func TestServersDefaultNameDeployments(t *testing.T) {
	deployments := 3
	name := "test"
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, "test-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	for i := 1; i < deployments; i++ {
		assert.Equal(t, fmt.Sprintf("test-kieserver-%v", i+1), env.Servers[i].DeploymentConfigs[0].Name)
	}
}

// test-kieserver | test-kieserver2 | test-kieserver3
func TestServersDefaultNameArray(t *testing.T) {
	deployments := 3
	name := "test"
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{}, {}, {},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, "test-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	for i := 1; i < deployments; i++ {
		assert.Equal(t, fmt.Sprintf("test-kieserver%v", i+1), env.Servers[i].DeploymentConfigs[0].Name)
	}
}

// test-kieserver | test-kieserver-2 | test-kieserver-3
// test-kieserver2
// test-kieserver3 | test-kieserver3-2
// test-kieserver4
func TestServersDefaultNameMixed(t *testing.T) {
	deployments0 := 3
	deployments2 := 2
	deployments := 7
	name := "test"
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments0)},
					{},
					{Deployments: Pint(deployments2)},
					{},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, "test-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-kieserver2", env.Servers[deployments0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-kieserver3", env.Servers[deployments0+1].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-kieserver4", env.Servers[deployments0+1+deployments2].DeploymentConfigs[0].Name)
	for i := 1; i < deployments0; i++ {
		assert.Equal(t, fmt.Sprintf("test-kieserver-%v", i+1), env.Servers[i].DeploymentConfigs[0].Name)
	}
	for i := deployments0 + 1; i < deployments2; i++ {
		assert.Equal(t, fmt.Sprintf("test-kieserver3-%v", i+1), env.Servers[i].DeploymentConfigs[0].Name)
	}

}

func TestImageRegistry(t *testing.T) {
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
			Environment: "rhpam-trial",
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, registry1, cr.Spec.ImageRegistry.Registry)
	assert.Equal(t, true, cr.Spec.ImageRegistry.Insecure)

	registry2 := "registry2.test.com:5000"
	cr2 := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			ImageRegistry: v1.KieAppRegistry{
				Registry: registry2,
			},
		},
	}
	_, err = GetEnvironment(cr2, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, registry2, cr2.Spec.ImageRegistry.Registry)
	assert.Equal(t, false, cr2.Spec.ImageRegistry.Insecure)
}

func buildKieApp(name string, deployments int) *v1.KieApp {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Console: v1.SecuredKieAppObject{
					KieAppObject: v1.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources,
					},
				},
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						SecuredKieAppObject: v1.SecuredKieAppObject{
							KieAppObject: v1.KieAppObject{
								Env:       sampleEnv,
								Resources: sampleResources,
							},
						},
					},
				},
				SmartRouter: v1.KieAppObject{
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
	assert.Equal(t, constants.DefaultKieDeployments, *cr.Spec.Objects.Servers[0].Deployments, "Default number of kieserver deployments not being set in CR")
	assert.Len(t, cr.Spec.Objects.Servers, 1, "There should be 1 custom kieserver being set by default")
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
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
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
	assert.Equal(t, deployments, *cr.Spec.Objects.Servers[0].Deployments, "Number of kieserver deployments not set properly in CR")
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
				Servers: []v1.KieServerSet{
					{Deployments: Pint(deployments)},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	kieServerID := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "test-kieserver"}
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
	kieServerID2 := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "test-kieserver-2"}
	assert.Contains(t, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID2)
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
					{
						Name: "alpha",
					},
					{
						Name: "beta",
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	kieServerID := corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "alpha"}
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, kieServerID)
	kieServerID = corev1.EnvVar{Name: "KIE_SERVER_ID", Value: "beta"}
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
					{
						Name: "one",
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "hello-rules:latest",
						},
					},
					{
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
					{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "hello-rules:latest",
						},
					},
					{
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
	assert.Equal(t, "test-kieserver:latest", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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

func TestExampleServerCommonConfig(t *testing.T) {
	kieApp := LoadKieApp(t, "examples", "server_config.yaml")
	env, err := GetEnvironment(&kieApp, test.MockService())
	assert.NoError(t, err, "Error getting environment for %v", kieApp.Spec.Environment)
	assert.Equal(t, 6, len(env.Servers), "Expect six servers")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].DeploymentConfigs[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].Services[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-ping", env.Servers[len(env.Servers)-2].Services[1].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].Routes[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-http", env.Servers[len(env.Servers)-2].Routes[1].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2", env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2", env.Servers[len(env.Servers)-1].Services[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2-ping", env.Servers[len(env.Servers)-1].Services[1].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2", env.Servers[len(env.Servers)-1].Routes[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2-http", env.Servers[len(env.Servers)-1].Routes[1].Name, "Unexpected name for object")
}

func TestGetKieSetIndex(t *testing.T) {
	assert.Equal(t, "", getKieSetIndex(0, 0))
	assert.Equal(t, "2", getKieSetIndex(1, 0))
	assert.Equal(t, "-2", getKieSetIndex(0, 1))
	assert.Equal(t, "2-3", getKieSetIndex(1, 2))
	assert.Equal(t, "3", getKieSetIndex(2, 0))
	assert.Equal(t, "-3", getKieSetIndex(0, 2))
}

func TestDatabaseExternalInvalid(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(2),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseExternal,
						},
					},
				},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "external database configuration is mandatory for external database type", "Expected database configuration error")
}

func TestDatabaseExternal(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseExternal,
							ExternalConfig: &v1.ExternalDatabaseObject{
								Dialect:              "org.hibernate.dialect.Oracle10gDialect",
								JndiName:             "java:jboss/OracleDS",
								Driver:               "oracle",
								ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker",
								ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter",
								BackgroundValidation: "false",
								Username:             "oracleUser",
								Password:             "oraclePwd",
								JdbcURL:              "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7",
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, 0, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, "RHPAM", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "DATASOURCES"))
		assert.Equal(t, "true", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_JTA"))
		assert.Equal(t, "10000", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "TIMER_SERVICE_DATA_STORE_REFRESH_INTERVAL"))
		assert.Equal(t, "oracle", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))
		assert.Equal(t, "oracleUser", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_USERNAME"))
		assert.Equal(t, "oraclePwd", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD"))
		assert.Equal(t, "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_URL"))
		assert.Equal(t, "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_XA_CONNECTION_PROPERTY_URL"))
		assert.Equal(t, "false", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_BACKGROUND_VALIDATION"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_VALIDATION_MILLIS"))
		assert.Equal(t, "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_CONNECTION_CHECKER"))
		assert.Equal(t, "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_EXCEPTION_SORTER"))
		assert.Equal(t, "java:jboss/OracleDS", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_JNDI"))
		assert.Equal(t, "org.hibernate.dialect.Oracle10gDialect", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_PERSISTENCE_DIALECT"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DATABASE"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_SERVICE_HOST"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_SERVICE_PORT"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_NONXA"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_MIN_POOL_SIZE"))
		assert.Equal(t, "", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_MAX_POOL_SIZE"))

	}
}

func TestDatabaseH2(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseH2,
							Size: "10Mi",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-claim", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}

func TestDatabaseH2Ephemeral(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseH2,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-h2-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].Name)
		assert.NotNil(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].EmptyDir)
		assert.Equal(t, 0, len(env.Servers[i].PersistentVolumeClaims))
	}
}

func TestDatabaseMySQL(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseMySQL,
							Size: "10Mi",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mysql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}

func TestDatabaseMySQLDefaultSize(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseMySQL,
							Size: "",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mysql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Credentials
		adminUser := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_USERNAME")
		assert.NotEmpty(t, adminUser, "The admin user must not be empty")
		assert.Equal(t, adminUser, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "MYSQL_USER"))
		adminPwd := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
		assert.NotEmpty(t, adminPwd, "The admin password should have been generated")
		assert.Equal(t, adminPwd, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "MYSQL_PASSWORD"))
		dbName := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DATABASE")
		assert.NotEmpty(t, dbName, "The Database Name must not be empty")
		assert.Equal(t, dbName, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "MYSQL_DATABASE"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("1Gi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}
func TestDatabaseMySQLTrialEphemeral(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabaseMySQL,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mysql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.NotNil(t, env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].EmptyDir)
		assert.Equal(t, 0, len(env.Servers[i].PersistentVolumeClaims))
	}
}

func TestDatabasePostgresql(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabasePostgreSQL,
							Size: "10Mi",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}

func TestDatabasePostgresqlDefaultSize(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamProductionImmutable,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabasePostgreSQL,
							Size: "",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-monitoring-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Credentials
		adminUser := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_USERNAME")
		assert.NotEmpty(t, adminUser, "The admin user must not be empty")
		assert.Equal(t, adminUser, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "POSTGRESQL_USER"))
		adminPwd := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
		assert.NotEmpty(t, adminPwd, "The admin password should have been generated")
		assert.Equal(t, adminPwd, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "POSTGRESQL_PASSWORD"))
		dbName := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DATABASE")
		assert.NotEmpty(t, dbName, "The Database Name must not be empty")
		assert.Equal(t, dbName, getEnvVariable(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0], "POSTGRESQL_DATABASE"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("1Gi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}
func TestDatabasePostgresqlTrialEphemeral(t *testing.T) {
	deployments := 2
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: v1.RhpamTrial,
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &v1.DatabaseObject{
							Type: v1.DatabasePostgreSQL,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift", cr.Spec.CommonConfig.Version), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 3, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].Services[2].ObjectMeta.Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Servers[i].DeploymentConfigs[1].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].Name)
		assert.NotNil(t, env.Servers[i].DeploymentConfigs[1].Spec.Template.Spec.Volumes[0].EmptyDir)
		assert.Equal(t, 0, len(env.Servers[i].PersistentVolumeClaims))
	}
}

func LoadKieApp(t *testing.T, folder string, fileName string) v1.KieApp {
	box := packr.NewBox("../../../../deploy/" + folder)
	yamlString, err := box.FindString(fileName)
	assert.NoError(t, err, "Error reading yaml %v/%v", folder, fileName)
	var kieApp v1.KieApp
	err = yaml.Unmarshal([]byte(yamlString), &kieApp)
	assert.NoError(t, err, "Error parsing yaml %v/%v", folder, fileName)
	return kieApp
}
