package defaults

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var bcmImage = constants.RhpamPrefix + "-businesscentral-monitoring" + constants.RhelVersion

func TestLoadUnknownEnvironment(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
		}
	}()

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: "unknown",
		},
	}

	_, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Status.Applied.Version, cr.Spec.Environment, cr.Name), err.Error())
}

func TestInaccessibleConfigMap(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "map-test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
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
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Status.Applied.Version, cr.Spec.Environment, cr.Name), err.Error())
	assert.NotNil(t, cr.Status.Applied.Objects.Servers)
	assert.Len(t, cr.Status.Applied.Objects.Servers, 1)
	assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].Replicas)
}

func TestMultipleServerDeployment(t *testing.T) {
	deployments := 6
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
		}
	}()
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: Pint(deployments)},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, deployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, deployments), env.Servers[deployments-1].DeploymentConfigs[0].Name)
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, deployments), cr.Spec.Objects.Servers[deployments-1].Name)
	assert.Nil(t, err)
}

func TestRHPAMTrialEnvironment(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	assert.Len(t, mainService.Spec.Ports, 2, "The rhpamcentr service should have two ports")
	assert.False(t, hasPort(mainService, 8001), "The rhpamcentr service should NOT listen on port 8001")

	pingService := getService(wbServices, "test-rhpamcentr-ping")
	assert.NotNil(t, pingService, "Ping service not found")
	assert.False(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, getLivenessReadiness("/rest/ready"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/rest/healthy"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
}

func TestRHDMTrialEnvironment(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	assert.Len(t, mainService.Spec.Ports, 2, "The rhdmcentr service should have three ports")
	assert.False(t, hasPort(mainService, 8001), "The rhdmcentr service should NOT listen on port 8001")

	pingService := getService(wbServices, "test-rhdmcentr-ping")
	assert.False(t, hasPort(pingService, 8888), "The ping service should not listen on port 8888")
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhdm-decisioncentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	assert.Equal(t, getLivenessReadiness("/rest/ready"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/rest/healthy"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
}

func TestRhpamcentrMonitoringEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	adminPassword := cr.Status.Applied.CommonConfig.AdminPassword

	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, adminPassword, cr.Status.Applied.CommonConfig.AdminPassword)
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	pingService := getService(env.Console.Services, "test-rhpamcentrmon-ping")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.True(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")
	for i := 0; i < len(env.Servers); i++ {
		assert.Equal(t, "PRODUCTION", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_MODE"))
	}
}

func TestRhdmAuthoringHAEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	var partitionValue int32
	partitionValue = 0

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
	assert.Equal(t, "test-datagrid", env.Others[0].StatefulSets[0].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[0].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[0].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, "test-amq", env.Others[0].StatefulSets[1].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[1].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[1].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/jboss-datagrid-7/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/amq7/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].BrokerImage, constants.VersionConstants[cr.Status.Applied.Version].BrokerImageTag), env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "rhdm-decisioncentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < len(env.Servers); i++ {
		assert.Equal(t, "DEVELOPMENT", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_MODE"))
	}
}

func TestRhpamAuthoringHAEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			CommonConfig: api.CommonConfig{
				AdminPassword:      "admin",
				AMQPassword:        "amq",
				AMQClusterPassword: "cluster",
			},
		},
		Status: api.KieAppStatus{
			Applied: api.KieAppSpec{
				CommonConfig: api.CommonConfig{
					AdminPassword:      "RedHat",
					AMQPassword:        "RedHat",
					AMQClusterPassword: "RedHat",
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	var partitionValue int32
	partitionValue = 0

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
	assert.Equal(t, "test-datagrid", env.Others[0].StatefulSets[0].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[0].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[0].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, "test-amq", env.Others[0].StatefulSets[1].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[1].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[1].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/jboss-datagrid-7/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/amq7/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].BrokerImage, constants.VersionConstants[cr.Status.Applied.Version].BrokerImageTag), env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	amqClusterPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_JMS_BROKER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	amqPassword := getEnvVariable(env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0], "AMQ_PASSWORD")
	assert.Equal(t, "amq", amqPassword, "Expected provided password to take effect, but found %v", amqPassword)
	adminPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, "admin", adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	amqClusterPassword = getEnvVariable(env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0], "AMQ_CLUSTER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	pingService := getService(env.Console.Services, "test-rhpamcentr-ping")
	assert.Len(t, pingService.Spec.Ports, 1, "The ping service should have only one port")
	assert.True(t, hasPort(pingService, 8888), "The ping service should listen on port 8888")

}

func TestRhdmProdImmutableEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PORT"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhdm-decisioncentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdwSmartRouter(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.False(t, env.SmartRouter.Omit, "SmarterRouter should not be omitted")
	assert.Equal(t, "test-smartrouter", env.SmartRouter.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-smartrouter", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_SERVICE"), "Variable should exist")
	assert.Equal(t, "9000", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PORT"), "Variable should exist")
	assert.Equal(t, "http", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "", getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_ROUTE_NAME"), "Variable should exist")
	assert.Equal(t, env.SmartRouter.DeploymentConfigs[0].Spec.Strategy.Type, appsv1.DeploymentStrategyTypeRolling)
	assert.Equal(t, env.SmartRouter.DeploymentConfigs[0].Spec.Strategy.RollingParams.MaxSurge, &intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "100%"})

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-smartrouter", env.SmartRouter.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRolling, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
	assert.Equal(t, &intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "100%"}, env.Console.DeploymentConfigs[0].Spec.Strategy.RollingParams.MaxSurge)
}

func TestRhpamProdSmartRouterWithSSL(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{
					Protocol:         "https",
					UseExternalRoute: true,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.False(t, env.SmartRouter.Omit, "SmarterRouter should not be omitted")
	assert.Equal(t, "test-smartrouter", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_SERVICE"), "Variable should exist")
	assert.Equal(t, "9443", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PORT"), "Variable should exist")
	assert.Equal(t, "https", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-smartrouter", env.SmartRouter.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-smartrouter", getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_ROUTE_NAME"), "Variable should exist")
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhdmProdImmutableJMSEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Jms: &api.KieAppJmsObject{
							EnableIntegration:  true,
							ExecutorTransacted: true,
							Username:           "adminUser",
							Password:           "adminPassword",
							AuditTransacted:    Pbool(false),
							EnableAudit:        true,
							QueueAudit:         "queue/CUSTOM.KIE.SERVER.AUDIT",
							EnableSignal:       true,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Equal(t, "test-jms-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "rhdm-decisioncentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdImmutableEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")

	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdImmutableJMSEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Jms: &api.KieAppJmsObject{
							EnableIntegration:  true,
							ExecutorTransacted: true,
							Username:           "adminUser",
							Password:           "adminPassword",
							AuditTransacted:    Pbool(false),
							EnableAudit:        true,
							QueueAudit:         "queue/CUSTOM.KIE.SERVER.AUDIT",
							EnableSignal:       true,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	cr.Spec.Version = "7.7.1"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[2].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[2].Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdImmutableJMSEnvironmentWithSSL(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Jms: &api.KieAppJmsObject{
							EnableIntegration:     true,
							ExecutorTransacted:    true,
							Username:              "adminUser",
							Password:              "adminPassword",
							AuditTransacted:       Pbool(false),
							EnableAudit:           true,
							QueueAudit:            "queue/CUSTOM.KIE.SERVER.AUDIT",
							EnableSignal:          true,
							AMQSecretName:         "broker-secret",
							AMQTruststoreName:     "broker.ts",
							AMQTruststorePassword: "changeme",
							AMQKeystoreName:       "broker.ks",
							AMQKeystorePassword:   "changeme",
						},
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.False(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	assert.Equal(t, "amq-tcp-ssl", env.Servers[0].Routes[2].Name)
	assert.False(t, env.Servers[0].Routes[2].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, cr.Status.Applied.Objects.Servers[0].Jms.AMQEnableSSL)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

}

func TestRhpamProdImmutableJMSEnvironmentExecutorDisabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Jms: &api.KieAppJmsObject{
							EnableIntegration:  true,
							Executor:           Pbool(false),
							ExecutorTransacted: true,
							EnableAudit:        true,
							QueueAudit:         "queue/CUSTOM.KIE.SERVER.AUDIT",
							EnableSignal:       true,
							QueueSignal:        "queue/CUSTOM.KIE.SERVER.SIGNAL",
						},
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Empty(t, cr.Spec.Objects.Servers[0].Jms.Username)
	assert.Empty(t, cr.Spec.Objects.Servers[0].Jms.Password)
	assert.NotEmpty(t, cr.Status.Applied.Objects.Servers[0].Jms.Username)
	assert.NotEmpty(t, cr.Status.Applied.Objects.Servers[0].Jms.Password)
	user := cr.Status.Applied.Objects.Servers[0].Jms.Username
	password := cr.Status.Applied.Objects.Servers[0].Jms.Password
	_, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, user, cr.Status.Applied.Objects.Servers[0].Jms.Username)
	assert.Equal(t, password, cr.Status.Applied.Objects.Servers[0].Jms.Password)

	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_EXECUTOR_JMS"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_EXECUTOR_JMS_TRANSACTED"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_ENABLE_AUDIT"), "Variable should exist")
	assert.Equal(t, "queue/CUSTOM.KIE.SERVER.AUDIT", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_QUEUE_AUDIT"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_ENABLE_SIGNAL"), "Variable should exist")
	assert.Equal(t, "queue/CUSTOM.KIE.SERVER.SIGNAL", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_QUEUE_SIGNAL"), "Variable should exist")
	assert.Equal(t, "queue/KIE.SERVER.REQUEST, queue/KIE.SERVER.RESPONSE, queue/CUSTOM.KIE.SERVER.SIGNAL, queue/CUSTOM.KIE.SERVER.AUDIT", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "AMQ_QUEUES"), "Variable should exist")

	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

}

func testAMQEnvs(t *testing.T, kieserverEnvs []corev1.EnvVar, amqEnvs []corev1.EnvVar) {
	for _, env := range kieserverEnvs {
		switch e := env.Name; e {
		case "KIE_SERVER_JMS_QUEUE_REQUEST":
			assert.Equal(t, "queue/KIE.SERVER.REQUEST", env.Value)

		case "KIE_SERVER_JMS_QUEUE_RESPONSE":
			assert.Equal(t, "queue/KIE.SERVER.RESPONSE", env.Value)

		case "KIE_SERVER_JMS_QUEUE_EXECUTOR":
			assert.Equal(t, "queue/KIE.SERVER.EXECUTOR", env.Value)

		case "KIE_SERVER_JMS_QUEUE_SIGNAL":
			assert.Equal(t, "queue/KIE.SERVER.SIGNAL", env.Value)

		case "KIE_SERVER_JMS_QUEUE_AUDIT":
			assert.Equal(t, "queue/CUSTOM.KIE.SERVER.AUDIT", env.Value)

		case "KIE_SERVER_JMS_ENABLE_SIGNAL":
			assert.Equal(t, "true", env.Value)

		case "KIE_SERVER_JMS_ENABLE_AUDIT":
			assert.Equal(t, "true", env.Value)

		case "KIE_SERVER_JMS_AUDIT_TRANSACTED":
			assert.Equal(t, "false", env.Value)

		case "KIE_SERVER_EXECUTOR_JMS_TRANSACTED":
			assert.Equal(t, "true", env.Value)

		case "KIE_SERVER_EXECUTOR_JMS":
			assert.Equal(t, "true", env.Value)

		case "MQ_SERVICE_PREFIX_MAPPING":
			assert.Equal(t, "test-jms-kieserver-amq7=AMQ", env.Value)

		case "AMQ_USERNAME":
			assert.Equal(t, "adminUser", env.Value)

		case "AMQ_PASSWORD":
			assert.Equal(t, "adminPassword", env.Value)

		case "AMQ_PROTOCOL":
			assert.Equal(t, "tcp", env.Value)

		case "AMQ_QUEUES":
			assert.Equal(t, "queue/KIE.SERVER.EXECUTOR, queue/KIE.SERVER.REQUEST, queue/KIE.SERVER.RESPONSE, queue/KIE.SERVER.SIGNAL, queue/CUSTOM.KIE.SERVER.AUDIT", env.Value)
		}
	}

	for _, env := range amqEnvs {
		switch e := env.Name; e {
		case "AMQ_USER":
			assert.Equal(t, "adminUser", env.Value)

		case "AMQ_PASSWORD":
			assert.Equal(t, "adminPassword", env.Value)

		case "AMQ_ROLE":
			assert.Equal(t, "admin", env.Value)

		case "AMQ_NAME":
			assert.Equal(t, "broker", env.Value)

		case "AMQ_TRANSPORTS":
			assert.Equal(t, "openwire", env.Value)

		case "AMQ_REQUIRE_LOGIN":
			assert.Equal(t, "true", env.Value)

		case "AMQ_QUEUES":
			assert.Equal(t, "queue/KIE.SERVER.EXECUTOR, queue/KIE.SERVER.REQUEST, queue/KIE.SERVER.RESPONSE, queue/KIE.SERVER.SIGNAL, queue/CUSTOM.KIE.SERVER.AUDIT", env.Value)
		}
	}
}

func createJvmTestObject() *api.JvmObject {
	jvmObject := api.JvmObject{
		JavaOptsAppend:             "-Dsome.property=foo",
		JavaMaxMemRatio:            Pint32(50),
		JavaInitialMemRatio:        Pint32(25),
		JavaMaxInitialMem:          Pint32(4096),
		JavaDiagnostics:            Pbool(true),
		JavaDebug:                  Pbool(true),
		JavaDebugPort:              Pint32(8787),
		GcMinHeapFreeRatio:         Pint32(20),
		GcMaxHeapFreeRatio:         Pint32(40),
		GcTimeRatio:                Pint32(4),
		GcAdaptiveSizePolicyWeight: Pint32(90),
		GcMaxMetaspaceSize:         Pint32(100),
		GcContainerOptions:         "-XX:+UseG1GC",
	}
	return &jvmObject
}

func testJvmEnv(t *testing.T, envs []corev1.EnvVar) {
	for _, env := range envs {
		switch e := env.Name; e {
		case "JAVA_OPTS_APPEND":
			assert.Equal(t, "-Dsome.property=foo", env.Value)

		case "JAVA_MAX_MEM_RATIO":
			assert.Equal(t, "50", env.Value)

		case "JAVA_INITIAL_MEM_RATIO":
			assert.Equal(t, "25", env.Value)

		case "JAVA_MAX_INITIAL_MEM":
			assert.Equal(t, "4096", env.Value)

		case "JAVA_DIAGNOSTICS":
			assert.Equal(t, "true", env.Value)

		case "JAVA_DEBUG":
			assert.Equal(t, "true", env.Value)

		case "JAVA_DEBUG_PORT":
			assert.Equal(t, "8787", env.Value)

		case "GC_MIN_HEAP_FREE_RATIO":
			assert.Equal(t, "20", env.Value)

		case "GC_MAX_HEAP_FREE_RATIO":
			assert.Equal(t, "40", env.Value)

		case "GC_TIME_RATIO":
			assert.Equal(t, "4", env.Value)

		case "GC_ADAPTIVE_SIZE_POLICY_WEIGHT":
			assert.Equal(t, "90", env.Value)

		case "GC_MAX_METASPACE_SIZE":
			assert.Equal(t, "100", env.Value)

		case "GC_CONTAINER_OPTIONS":
			assert.Equal(t, "-XX:+UseG1GC", env.Value)
		}
	}
}

func TestInvalidBuildConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(2),
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: api.GitSource{
								URI:       "http://git.example.com",
								Reference: "somebranch",
							},
							Webhooks: []api.WebhookSecret{
								{
									Type:   api.GitHubWebhook,
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

func TestExtensionImageBuildConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
							ExternalConfig: &api.ExternalDatabaseObject{
								CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
									Driver:               "mssql",
									ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLValidConnectionChecker",
									ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLExceptionSorter",
									BackgroundValidation: "true",
									MinPoolSize:          "10",
									MaxPoolSize:          "10",
									Username:             "sqlserverUser",
									Password:             "sqlserverPwd",
									JdbcURL:              "jdbc:sqlserver://192.168.1.129:1433;DatabaseName=rhpam",
								},
								Dialect: "org.hibernate.dialect.SQLServerDialect",
							},
						},
						Build: &api.KieAppBuildObject{
							ExtensionImageStreamTag: "test-sqlserver:1.0",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, 1, len(env.Servers))
	assert.Equal(t, "openshift", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].From.Namespace)
	assert.Equal(t, "test-sqlserver:1.0", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].From.Name)
	assert.Equal(t, "./extensions/extras", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].Paths[0].DestinationDir)
	assert.Equal(t, "/extensions/.", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].Paths[0].SourcePath)
	server := env.Servers[0]
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, "test-kieserver:latest", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestExtensionImageBuildWithCustomConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
							ExternalConfig: &api.ExternalDatabaseObject{
								CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
									Driver:               "mssql",
									ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLValidConnectionChecker",
									ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLExceptionSorter",
									BackgroundValidation: "true",
									MinPoolSize:          "10",
									MaxPoolSize:          "10",
									Username:             "sqlserverUser",
									Password:             "sqlserverPwd",
									JdbcURL:              "jdbc:sqlserver://192.168.1.129:1433;DatabaseName=rhpam",
								},
								Dialect: "org.hibernate.dialect.SQLServerDialect",
							},
						},
						Build: &api.KieAppBuildObject{
							ExtensionImageStreamTag:          "test-sqlserver:1.0",
							ExtensionImageStreamTagNamespace: "hello-world-namespace",
							ExtensionImageInstallDir:         "/tmp/test/tested",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, 1, len(env.Servers))
	assert.Equal(t, "hello-world-namespace", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].From.Namespace)
	assert.Equal(t, "test-sqlserver:1.0", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].From.Name)
	assert.Equal(t, "./extensions/extras", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].Paths[0].DestinationDir)
	assert.Equal(t, "/tmp/test/tested/.", env.Servers[0].BuildConfigs[0].Spec.Source.Images[0].Paths[0].SourcePath)
	server := env.Servers[0]
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, "test-kieserver:latest", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestBuildConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "example",
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
	assert.Len(t, cr.Spec.Objects.Servers[0].Build.Webhooks, 0)
	var secret string
	for _, s := range cr.Status.Applied.Objects.Servers[0].Build.Webhooks {
		if s.Type == api.GitHubWebhook {
			secret = s.Secret
		}
	}
	checkWebhooks(t, secret, cr, env)

	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, cr.Spec.Objects.Servers[0].Build.Webhooks, 0)
	checkWebhooks(t, secret, cr, env)

	secret = "s3cr3t"
	cr.Spec.Objects.Servers[0].Build.Webhooks = []api.WebhookSecret{{Type: api.GitHubWebhook, Secret: secret}}
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, cr.Spec.Objects.Servers[0].Build.Webhooks, 1)
	assert.Equal(t, secret, cr.Spec.Objects.Servers[0].Build.Webhooks[0].Secret)
	assert.Equal(t, 2, len(env.Servers))
	assert.Equal(t, "example", env.Servers[0].BuildConfigs[0].Spec.Source.ContextDir)
	checkWebhooks(t, secret, cr, env)

	// Server #0
	server := env.Servers[0]
	assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
	assert.Equal(t, "http://git.example.com", server.BuildConfigs[0].Spec.Source.Git.URI)
	assert.Equal(t, "somebranch", server.BuildConfigs[0].Spec.Source.Git.Ref)

	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
	assert.Equal(t, "https://maven.mirror.com/", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[1].Value)
	assert.Equal(t, "dir", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[2].Value)
	for _, s := range server.BuildConfigs[0].Spec.Triggers {
		if s.GitHubWebHook != nil {
			assert.NotEmpty(t, s.GitHubWebHook.Secret)
			assert.Equal(t, secret, s.GitHubWebHook.Secret)
		}
	}

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

func checkWebhooks(t *testing.T, secret string, cr *api.KieApp, env api.Environment) {
	assert.Len(t, cr.Status.Applied.Objects.Servers[0].Build.Webhooks, 2)
	for _, webhook := range cr.Status.Applied.Objects.Servers[0].Build.Webhooks {
		if webhook.Type == api.GitHubWebhook {
			assert.NotEmpty(t, webhook.Secret)
			assert.Equal(t, secret, webhook.Secret)
		}
		if webhook.Type == api.GenericWebhook {
			assert.NotEmpty(t, webhook.Secret)
		}
	}
	assert.Len(t, env.Servers[0].BuildConfigs[0].Spec.Triggers, 4)
	for _, trigger := range env.Servers[0].BuildConfigs[0].Spec.Triggers {
		if trigger.GitHubWebHook != nil {
			assert.NotEmpty(t, trigger.GitHubWebHook.Secret)
			assert.Equal(t, secret, trigger.GitHubWebHook.Secret)
		}
		if trigger.GenericWebHook != nil {
			assert.NotEmpty(t, trigger.GenericWebHook.Secret)
		}
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			CommonConfig: api.CommonConfig{
				DBPassword: "Database",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Nil(t, err, "Error getting authoring environment")
	dbPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
	assert.Equal(t, "Database", dbPassword, "Expected provided password to take effect, but found %v", dbPassword)
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Name), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, string(appsv1.DeploymentStrategyTypeRolling), string(env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Strategy.Type), "The DC should use a Rolling strategy when using the H2 DB")
	assert.NotEqual(t, api.Environment{}, env, "Environment should not be empty")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

	// test kieserver probes
	assert.Equal(t, getLivenessReadiness("/services/rest/server/readycheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/services/rest/server/healthcheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
}

func TestAuthoringHAEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Nil(t, err, "Error getting authoring-ha environment")
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Name), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, api.Environment{}, env, "Environment should not be empty")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
}

func TestConstructConsoleObject(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Console.Replicas)

	cr.Spec.Objects = api.KieAppObjects{Console: api.ConsoleObject{}}
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Nil(t, cr.Spec.Objects.Servers)
	assert.Nil(t, cr.Spec.Objects.Console.Replicas)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Console.Replicas)

	cr.Spec.Objects.Console.Replicas = Pint32(1)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Console.Replicas)
	assert.Equal(t, Pint32(1), cr.Spec.Objects.Console.Replicas)

	cr.Spec.Objects.Console.Replicas = Pint32(3)
	cr.Spec.Objects.Console.Image = "test"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "test", cr.Spec.Objects.Console.Image)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Console.Replicas)
	assert.Equal(t, Pint32(3), cr.Spec.Objects.Console.Replicas)

	cr = buildKieApp(name, 1)
	cr.Spec.Objects.Console.Replicas = Pint32(3)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Console.Replicas)
	assert.Equal(t, Pint32(3), cr.Spec.Objects.Console.Replicas)

	env = ConsolidateObjects(env, cr)
	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", name), env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, int32(1), env.Console.DeploymentConfigs[0].Spec.Replicas)
	assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Status.Applied.Version), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	for i := range sampleEnv {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, sampleEnv[i], "Environment merge not functional. Expecting: %v", sampleEnv[i])
	}
}

func TestConstructSmartRouterObject(t *testing.T) {
	name := "test"
	cr := buildKieApp(name, 1)
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter.Resources)
	assert.Nil(t, cr.Spec.Objects.SmartRouter.Replicas)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter.Replicas)

	cr.Spec.Objects.SmartRouter.Replicas = Pint32(2)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter.Resources)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter.Replicas)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter)

	env = ConsolidateObjects(env, cr)
	assert.Equal(t, fmt.Sprintf("%s-smartrouter", name), env.SmartRouter.DeploymentConfigs[0].Name)
	assert.Equal(t, int32(2), env.SmartRouter.DeploymentConfigs[0].Spec.Replicas)
	assert.Equal(t, fmt.Sprintf("rhpam-smartrouter-rhel8:%s", cr.Status.Applied.Version), env.SmartRouter.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
		assert.NotNil(t, cr.Spec.Objects.Servers)
		assert.NotNil(t, cr.Spec.Objects.Servers[0].Resources)
		assert.NotNil(t, cr.Status.Applied.Objects.Servers)
		assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].Replicas)
		assert.Equal(t, cr.Spec.Objects.Servers[0].Resources, cr.Status.Applied.Objects.Servers[0].Resources)

		env = ConsolidateObjects(env, cr)
		assert.Equal(t, fmt.Sprintf("%s-kieserver", name), env.Servers[0].DeploymentConfigs[0].Name)
		assert.Equal(t, int32(1), env.Servers[0].DeploymentConfigs[0].Spec.Replicas)
		assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Status.Applied.Version), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
			assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Status.Applied.Version), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
		assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Status.Applied.Version), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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

var sampleResources = &corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		"memory": *resource.NewQuantity(1, "Mi"),
	},
}

func TestUnknownEnvironmentObjects(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: "unknown",
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Status.Applied.Version, cr.Spec.Environment, cr.Name), err.Error())

	env = ConsolidateObjects(env, cr)
	assert.NotNil(t, err)

	log.Debug("Testing with environment ", cr.Spec.Environment)
	assert.Equal(t, api.Environment{}, env, "Env object should be empty")
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{
					Jvm: &api.JvmObject{
						JavaOptsAppend:     "",
						GcContainerOptions: "",
						JavaDebug:          Pbool(false),
					},
				},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						KieAppObject: api.KieAppObject{
							Env: []corev1.EnvVar{
								envReplace,
								envAddition,
							},
						},
						Jvm: createJvmTestObject(),
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, deployments), env.Servers[deployments-1].DeploymentConfigs[0].Name)
	assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Status.Applied.Version), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "replaced",
	})
	assert.Contains(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition, "Environment additions not functional")
	testJvmEnv(t, env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "JAVA_DEBUG", Value: strconv.FormatBool(*cr.Spec.Objects.Console.Jvm.JavaDebug)})
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "JAVA_OPTS_APPEND", Value: cr.Spec.Objects.Console.Jvm.JavaOptsAppend})
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "GC_CONTAINER_OPTIONS", Value: cr.Spec.Objects.Console.Jvm.GcContainerOptions})
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name: "server-a",
						KieAppObject: api.KieAppObject{
							Env: []corev1.EnvVar{
								envReplace,
								envAddition,
								commonAddition,
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
	for index := 0; index < 1; index++ {
		s := env.Servers[index]
		assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Status.Applied.Version), s.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			CommonConfig: api.CommonConfig{
				ApplicationName: "trial",
			},
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Env: []corev1.EnvVar{
							envReplace,
							envAddition,
						},
					},
					Jvm: createJvmTestObject(),
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
	assert.Equal(t, fmt.Sprintf("rhdm-decisioncentral-rhel8:%s", cr.Status.Applied.Version), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	adminUser := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_USER")
	assert.Equal(t, constants.DefaultAdminUser, adminUser, "AdminUser default not being set correctly")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})

	testJvmEnv(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestKieAppDefaults(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	assert.False(t, cr.Spec.Upgrades.Enabled)
	assert.Empty(t, cr.Spec.CommonConfig.ApplicationName)
	assert.Nil(t, cr.Spec.Objects.Console.Replicas)
	assert.Nil(t, cr.Spec.Objects.Servers)
	assert.NotEmpty(t, cr.Status.Applied.CommonConfig.ApplicationName)
	assert.NotNil(t, cr.Status.Applied.Objects.Console.Replicas)
	assert.Len(t, cr.Status.Applied.Objects.Servers, 1)
}

func TestMergeTrialAndCommonConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
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
		Name:      "test-kieserver-kie-pvol",
		MountPath: "/opt/kie/data",
	})
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "test-kieserver-kie-pvol",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}

func TestServerConflict(t *testing.T) {
	deployments := 2
	name := "test"
	duplicate := "testing"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	os.Setenv("INSECURE", "true")
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Nil(t, cr.Spec.ImageRegistry)

	registry2 := "registry2.test.com:5000"
	cr2 := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			ImageRegistry: &api.KieAppRegistry{
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
	os.Clearenv()
}

func buildKieApp(name string, deployments int) *api.KieApp {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources,
					},
				},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						KieAppObject: api.KieAppObject{
							Env:       sampleEnv,
							Resources: sampleResources,
						},
					},
				},
				SmartRouter: &api.SmartRouterObject{
					KieAppObject: api.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources,
					},
				},
			},
		},
	}
	return cr
}

func TestPartialTemplateConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmAuthoring,
			CommonConfig: api.CommonConfig{
				AdminUser:     "NewAdmin",
				AdminPassword: "MyPassword",
			},
		},
		Status: api.KieAppStatus{
			Applied: api.KieAppSpec{
				Environment: api.RhdmAuthoring,
				CommonConfig: api.CommonConfig{
					AdminPassword: "RedHat",
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting authoring environment")
	adminUser := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_USER")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, cr.Spec.CommonConfig.AdminUser, adminUser, "Expected provided user to take effect, but found %v", adminUser)
	assert.Equal(t, cr.Spec.CommonConfig.AdminPassword, adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	assert.Equal(t, cr.Spec.CommonConfig.AdminPassword, cr.Status.Applied.CommonConfig.AdminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Equal(t, "MyPassword", mavenPassword, "Expected default password of RedHat, but found %v", mavenPassword)

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			CommonConfig: api.CommonConfig{
				AdminPassword: "MyPassword",
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, "MyPassword", adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Equal(t, "MyPassword", mavenPassword, "Expected default password of RedHat, but found %v", mavenPassword)

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
}

func TestDefaultKieServerNum(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
		},
	}
	_, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, constants.DefaultKieDeployments, *cr.Status.Applied.Objects.Servers[0].Deployments, "Default number of kieserver deployments not being set in CR")
	assert.Len(t, cr.Status.Applied.Objects.Servers, 1, "There should be 1 custom kieserver being set by default")
}

func TestZeroKieServerDeployments(t *testing.T) {
	deployments := 0
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	assert.NotNil(t, cr.Status.Applied.Objects.Servers)
	assert.Equal(t, *cr.Spec.Objects.Servers[0].Deployments, *cr.Status.Applied.Objects.Servers[0].Deployments)
	assert.Equal(t, deployments, *cr.Spec.Objects.Servers[0].Deployments, "Number of kieserver deployments not set properly in CR")
}

func TestDefaultKieServerID(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: Pint(deployments)},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, env.Servers[0].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], cr.Status.Applied.Objects.Servers[0].Name)
	assert.Equal(t, env.Servers[1].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], strings.Join([]string{cr.Status.Applied.Objects.Servers[0].Name, "2"}, "-"))
}

func TestSetKieServerID(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name: "alpha",
						ID:   "omega",
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
	assert.Equal(t, env.Servers[0].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], cr.Spec.Objects.Servers[0].ID)
	assert.Equal(t, env.Servers[1].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], cr.Spec.Objects.Servers[1].Name)
}

func TestSetKieServerFrom(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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
						Build: &api.KieAppBuildObject{},
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "test",
							},
							Webhooks: []api.WebhookSecret{
								{
									Type:   api.GitHubWebhook,
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
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "anotherbranch",
								ContextDir: "test",
							},
							Webhooks: []api.WebhookSecret{
								{
									Type:   api.GitHubWebhook,
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
	assert.Equal(t, fmt.Sprintf("rhdm-kieserver-rhel8:%v", cr.Status.Applied.Version), env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "openshift", env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Namespace)
}

func TestExampleServerCommonConfig(t *testing.T) {
	kieApp := LoadKieApp(t, "examples/"+api.SchemeGroupVersion.Version, "server_config.yaml")
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(2),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
							ExternalConfig: &api.ExternalDatabaseObject{
								CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
									Driver:               "oracle",
									ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker",
									ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter",
									BackgroundValidation: "false",
									Username:             "oracleUser",
									Password:             "oraclePwd",
									JdbcURL:              "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7",
								},
								Dialect: "org.hibernate.dialect.Oracle10gDialect",
							},
						},
						KieAppObject: api.KieAppObject{
							Env: []corev1.EnvVar{
								{
									Name:  "RHPAM_JNDI",
									Value: "java:jboss/OracleDS",
								},
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	env = ConsolidateObjects(env, cr)

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseH2,
								Size: "10Mi",
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
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-claim", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Servers[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-claim", idx), env.Servers[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Servers[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}

func TestDefaultVersioning(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Applied.Version)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Applied.Version)
	assert.True(t, checkVersion(cr.Status.Applied.Version))
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestConfigVersioning(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     "6.3.1",
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Error(t, err, "Incompatible product versions should throw an error")
	assert.Equal(t, fmt.Sprintf("Product version %s is not allowed. The following versions are allowed - %s", cr.Status.Applied.Version, constants.SupportedVersions), err.Error())
	assert.Equal(t, "6.3.1", cr.Status.Applied.Version)
	major, minor, micro := MajorMinorMicro(cr.Status.Applied.Version)
	assert.Equal(t, "6", major)
	assert.Equal(t, "3", minor)
	assert.Equal(t, "1", micro)
	assert.False(t, checkVersion(cr.Status.Applied.Version))
}

func TestConfigMapNames(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	filepath := "envs/rhpam-trial.yaml"
	filename := strings.Join([]string{cr.Status.Applied.Version, filepath}, "/")
	cmNameT, fileT := convertToConfigMapName(filename)

	fileslice := strings.Split(filepath, "/")
	file := fileslice[len(fileslice)-1]
	assert.Equal(t, file, fileT)

	cmName := strings.Join([]string{constants.ConfigMapPrefix, cr.Status.Applied.Version, fileslice[0]}, "-")
	assert.Equal(t, cmName, cmNameT)
}

func TestDatabaseH2Ephemeral(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseH2,
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].Name)
		assert.Equal(t, 2, len(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-kie-pvol", idx), env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].Name)
		assert.NotNil(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[1].EmptyDir)
		assert.Equal(t, 0, len(env.Servers[i].PersistentVolumeClaims))
	}
}

func TestDatabaseMySQL(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseMySQL,
								Size: "10Mi",
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
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Databases[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Databases[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Databases[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
	cr.Spec.Version = "7.7.1"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseMySQL,
								Size: "",
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
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Credentials
		adminUser := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_USERNAME")
		assert.NotEmpty(t, adminUser, "The admin user must not be empty")
		assert.Equal(t, adminUser, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "MYSQL_USER"))
		adminPwd := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
		assert.NotEmpty(t, adminPwd, "The admin password should have been generated")
		assert.Equal(t, adminPwd, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "MYSQL_PASSWORD"))
		dbName := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DATABASE")
		assert.NotEmpty(t, dbName, "The Database Name must not be empty")
		assert.Equal(t, dbName, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "MYSQL_DATABASE"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Databases[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-claim", idx), env.Databases[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("1Gi"), env.Databases[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}

	cr.Spec.Version = "7.7.1"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseMySQL,
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// MYSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-mysql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.NotNil(t, env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].EmptyDir)
		assert.Equal(t, 0, len(env.Databases[i].PersistentVolumeClaims))
	}
	cr.Spec.Version = "7.7.1"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

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
		assert.Equal(t, "mariadb", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabasePostgreSQL,
								Size: "10Mi",
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
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Databases[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Databases[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("10Mi"), env.Databases[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
	cr.Spec.Version = "7.7.1"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabasePostgreSQL,
								Size: "",
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
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Credentials
		adminUser := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_USERNAME")
		assert.NotEmpty(t, adminUser, "The admin user must not be empty")
		assert.Equal(t, adminUser, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "POSTGRESQL_USER"))
		adminPwd := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
		assert.NotEmpty(t, adminPwd, "The admin password should have been generated")
		assert.Equal(t, adminPwd, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "POSTGRESQL_PASSWORD"))
		dbName := getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DATABASE")
		assert.NotEmpty(t, dbName, "The Database Name must not be empty")
		assert.Equal(t, dbName, getEnvVariable(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "POSTGRESQL_DATABASE"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, 1, len(env.Databases[i].PersistentVolumeClaims))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-claim", idx), env.Databases[i].PersistentVolumeClaims[0].Name)
		assert.Equal(t, resource.MustParse("1Gi"), env.Databases[i].PersistentVolumeClaims[0].Spec.Resources.Requests["storage"])
	}
}
func TestDatabasePostgresqlTrialEphemeral(t *testing.T) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						Database: &api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabasePostgreSQL,
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 2, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-ping", idx), env.Servers[i].Services[1].ObjectMeta.Name)
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].Services[0].ObjectMeta.Name)
		assert.Equal(t, 1, len(env.Servers[i].DeploymentConfigs))
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].DeploymentConfigs[0].Name)
		assert.Equal(t, "postgresql", getEnvVariable(env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_DRIVER"))

		// PostgreSQL Deployment
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql", idx), env.Databases[i].DeploymentConfigs[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, 1, len(env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s-postgresql-pvol", idx), env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].Name)
		assert.NotNil(t, env.Databases[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes[0].EmptyDir)
		assert.Equal(t, 0, len(env.Databases[i].PersistentVolumeClaims))
	}
}

func TestEnvCustomImageTag(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(2),
					},
				},
				SmartRouter:      &api.SmartRouterObject{},
				ProcessMigration: &api.ProcessMigrationObject{},
			},
		},
	}
	// test setting image with env vars
	image := "testing@sha256"
	imageTag := "e1168e1a1c6e4f248"
	imageName := image + ":" + imageTag
	imageURL := "image-registry.openshift-image-registry.svc:5000/openshift/" + image
	envConstants := constants.EnvironmentConstants[cr.Spec.Environment]
	os.Setenv(envConstants.App.ImageVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamKieImageVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamSmartRouterVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamProcessMigrationVar+constants.CurrentVersion, imageURL)
	checkImageNames(cr, imageName, imageURL, t)

	// test setting image with env vars, DM product, and different image url
	cr.Spec.Environment = api.RhdmAuthoring
	image = "testing"
	imageTag = cr.Status.Applied.Version
	imageName = image + ":" + imageTag
	imageURL = "image-registry.openshift-image-registry.svc:5000/openshift/" + imageName
	envConstants = constants.EnvironmentConstants[cr.Spec.Environment]
	os.Setenv(envConstants.App.ImageVar+cr.Status.Applied.Version, imageURL)
	os.Setenv(constants.DmKieImageVar+cr.Status.Applied.Version, imageURL)
	os.Setenv(constants.PamSmartRouterVar+cr.Status.Applied.Version, imageURL)
	checkImageNames(cr, imageName, imageURL, t)

	// test setting image with env vars, DM product, and different image url
	image = "testing"
	imageTag = "1.6"
	imageName = image + ":" + imageTag
	imageURL = "registry.redhat.io/openshift/" + imageName
	os.Setenv(envConstants.App.ImageVar+cr.Status.Applied.Version, imageURL)
	os.Setenv(constants.DmKieImageVar+cr.Status.Applied.Version, imageURL)
	os.Setenv(constants.PamSmartRouterVar+cr.Status.Applied.Version, imageURL)
	checkImageNames(cr, imageName, imageURL, t)

	// test useImageTags = true
	cr.Spec.UseImageTags = true
	env, err := GetEnvironment(cr, test.MockService())
	assert.Equal(t, constants.RhdmPrefix+"-kieserver"+constants.RhelVersion+":"+cr.Status.Applied.Version, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	// test that setting imagetag in CR overrides env vars
	cr.Spec.Environment = api.RhpamAuthoring
	imageURL = image + ":" + imageTag
	os.Setenv(envConstants.App.ImageVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamKieImageVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamBusinessCentralVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamSmartRouterVar+constants.CurrentVersion, imageURL)
	os.Setenv(constants.PamProcessMigrationVar+constants.CurrentVersion, imageURL)

	cr.Spec.UseImageTags = false
	imageTag = "1.5"
	imageName = image + ":" + imageTag
	cr.Spec.Objects.Console.ImageTag = imageTag
	cr.Spec.Objects.Servers[0].ImageTag = imageTag
	cr.Spec.Objects.SmartRouter.ImageTag = imageTag
	cr.Spec.Objects.ProcessMigration.ImageTag = imageTag
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 2, "Expect two KIE Servers to be created based on provided build configs")
	if isTagImage := getImageChangeName(env.Console.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[0].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[1].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.SmartRouter.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.ProcessMigration.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	// as versions progress and configs evolve, the following 4 tests should change from "imageName/image" to "imageURL" as the above tests do
	assert.Equal(t, imageName, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	// test that setting image in CR overrides env vars
	image = "testing-images"
	imageName = image + ":" + imageTag
	cr.Spec.Objects.Console.Image = image
	cr.Spec.Objects.Servers[0].Image = image
	cr.Spec.Objects.SmartRouter.Image = image
	cr.Spec.Objects.ProcessMigration.Image = image
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 2, "Expect two KIE Servers to be created based on provided build configs")
	if isTagImage := getImageChangeName(env.Console.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[0].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[1].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.SmartRouter.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.ProcessMigration.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	// as versions progress and configs evolve, the following 4 tests should change from "imageName/image" to "imageURL" as the above tests do
	assert.Equal(t, imageName, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageName, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	os.Clearenv()
}

func TestStorageClassName(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(1),
						Database:    &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabaseMySQL}},
					},
					{
						Deployments: Pint(1),
						Database:    &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabasePostgreSQL}},
					},
					{
						Deployments: Pint(1),
						Database:    &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabaseH2}},
					},
				},
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Nil(t, env.Console.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.SmartRouter.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Len(t, env.Others[0].StatefulSets, 0)

	cr.Spec.Objects.Servers[0].Database.Size = "10Gi"
	cr.Spec.Objects.Servers[1].Database.Size = "10Gi"
	cr.Spec.Objects.Servers[2].Database.Size = "10Gi"
	cr.Spec.Objects.Servers[0].StorageClassName = "silver"
	cr.Spec.Objects.Servers[1].StorageClassName = "silver"
	cr.Spec.Objects.Servers[2].StorageClassName = "silver"
	cr.Spec.Objects.Console = api.ConsoleObject{
		KieAppObject: api.KieAppObject{StorageClassName: "fast"},
	}
	cr.Spec.Objects.SmartRouter.StorageClassName = "slow"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "fast", *env.Console.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Nil(t, env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Len(t, env.Others[0].StatefulSets, 0)
	assert.Equal(t, "slow", *env.SmartRouter.PersistentVolumeClaims[0].Spec.StorageClassName)

	cr.Spec.Objects.Servers[0].Database.StorageClassName = "gold"
	cr.Spec.Objects.Servers[1].Database.StorageClassName = "gold1"
	cr.Spec.Objects.Servers[2].Database.StorageClassName = "gold2"
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "gold", *env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold1", *env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold2", *env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Len(t, env.Others[0].StatefulSets, 0)

	cr.Spec.Environment = api.RhpamAuthoringHA
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "fast", *env.Console.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold", *env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold1", *env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold2", *env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "slow", *env.SmartRouter.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "fast", *env.Others[0].StatefulSets[0].Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
	assert.Equal(t, "fast", *env.Others[0].StatefulSets[1].Spec.VolumeClaimTemplates[0].Spec.StorageClassName)

	cr.Spec.Environment = api.RhdmAuthoring
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "fast", *env.Console.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold", *env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold1", *env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold2", *env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "slow", *env.SmartRouter.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Len(t, env.Others[0].StatefulSets, 0)

	cr.Spec.Environment = api.RhdmAuthoringHA
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "fast", *env.Console.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold", *env.Databases[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold1", *env.Databases[1].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "gold2", *env.Servers[2].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "slow", *env.SmartRouter.PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, "fast", *env.Others[0].StatefulSets[0].Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
	assert.Equal(t, "fast", *env.Others[0].StatefulSets[1].Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
}

func TestGitHooks(t *testing.T) {
	var defaultMode int32 = 420
	tests := []struct {
		name                string
		gitHooks            *api.GitHooksVolume
		expectedVolumeMount *corev1.VolumeMount
		expectedVolume      *corev1.Volume
		expectedPath        string
	}{{
		name:                "GitHooks EnvVar is not present",
		gitHooks:            nil,
		expectedVolumeMount: nil,
		expectedVolume:      nil,
		expectedPath:        "",
	}, {
		name:                "GitHooks EnvVar is present and has default value",
		gitHooks:            &api.GitHooksVolume{},
		expectedVolumeMount: nil,
		expectedVolume:      nil,
		expectedPath:        constants.GitHooksDefaultDir,
	}, {
		name: "GitHooks DIR is configured",
		gitHooks: &api.GitHooksVolume{
			MountPath: "/some/path",
		},
		expectedVolumeMount: nil,
		expectedVolume:      nil,
		expectedPath:        "/some/path",
	}, {
		name: "ConfigMap GitHooks are configured",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Kind: "ConfigMap",
				Name: "test-cm",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.GitHooksVolume,
			MountPath: constants.GitHooksDefaultDir,
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.GitHooksVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-cm",
					},
					DefaultMode: &defaultMode,
				},
			},
		},
		expectedPath: constants.GitHooksDefaultDir,
	}, {
		name: "Secret GitHooks are configured",
		gitHooks: &api.GitHooksVolume{
			MountPath: "/some/path",
			From: &corev1.ObjectReference{
				Kind: "Secret",
				Name: "test-secret",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.GitHooksVolume,
			MountPath: "/some/path",
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.GitHooksVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "test-secret",
				},
			},
		},
		expectedPath: "/some/path",
	}, {
		name: "PersistentVolumeClaim GitHooks are configured",
		gitHooks: &api.GitHooksVolume{
			MountPath: "/some/path",
			From: &corev1.ObjectReference{
				Kind: "PersistentVolumeClaim",
				Name: "test-pvc",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.GitHooksVolume,
			MountPath: "/some/path",
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.GitHooksVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "test-pvc",
				},
			},
		},
		expectedPath: "/some/path",
	},
	}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{},
			},
		},
	}

	for _, item := range tests {
		expectedEnv := corev1.EnvVar{Name: "GIT_HOOKS_DIR", Value: item.expectedPath}
		cr.Spec.Objects.Console.GitHooks = item.gitHooks
		env, err := GetEnvironment(cr, test.MockService())
		assert.Nil(t, err, "Error getting prod environment")
		if item.expectedPath != "" {
			assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Test %s failed", item.name)
		} else {
			expectedEnv.Value = constants.GitHooksDefaultDir
			assert.NotContainsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Test %s failed", item.name)
		}
		if item.expectedVolumeMount != nil {
			assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, *item.expectedVolumeMount, "Test %s failed", item.name)
		}
		if item.expectedVolume != nil {
			assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, *item.expectedVolume, "Test %s failed", item.name)
		}
	}
}

func getImageChangeName(dc appsv1.DeploymentConfig) string {
	for _, trigger := range dc.Spec.Triggers {
		if trigger.Type == appsv1.DeploymentTriggerOnImageChange {
			return trigger.ImageChangeParams.From.Name
		}
	}
	return ""
}

func LoadKieApp(t *testing.T, folder string, fileName string) api.KieApp {
	box := packr.New("deploy/"+folder, "../../../../deploy/"+folder)
	yamlString, err := box.FindString(fileName)
	assert.NoError(t, err, "Error reading yaml %v/%v", folder, fileName)
	var kieApp api.KieApp
	err = yaml.Unmarshal([]byte(yamlString), &kieApp)
	assert.NoError(t, err, "Error parsing yaml %v/%v", folder, fileName)
	return kieApp
}

func getLivenessReadiness(uri string) *corev1.HTTPGetAction {
	return &corev1.HTTPGetAction{
		Scheme: corev1.URIScheme("HTTP"),
		Host:   "",
		Port: intstr.IntOrString{
			Type:   0,
			IntVal: 8080},
		Path: uri,
	}
}

func checkImageNames(cr *api.KieApp, imageName, imageURL string, t *testing.T) {
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 2, "Expect two KIE Servers to be created based on provided build configs")
	if isTagImage := getImageChangeName(env.Console.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[0].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.Servers[1].DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	if isTagImage := getImageChangeName(env.SmartRouter.DeploymentConfigs[0]); isTagImage != "" {
		assert.Equal(t, imageName, isTagImage)
	}
	assert.Equal(t, imageURL, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageURL, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageURL, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, imageURL, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	if len(env.ProcessMigration.DeploymentConfigs) > 0 {
		if isTagImage := getImageChangeName(env.ProcessMigration.DeploymentConfigs[0]); isTagImage != "" {
			assert.Equal(t, imageName, isTagImage)
		}
		assert.Equal(t, imageURL, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	}
}

func TestDeployProcessMigration(t *testing.T) {
	type args struct {
		cr *api.KieApp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"RhpamTrial",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			true,
		},
		{
			"RhpamProduction",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamProduction,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			true,
		},
		{
			"RhpamProductionImmutable",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamProductionImmutable,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			true,
		},
		{
			"RhpamAuthoring",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamAuthoring,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			true,
		},
		{
			"RhpamAuthoringHA",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamAuthoringHA,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			true,
		},
		{
			"RhdmTrial",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			false,
		},
		{
			"RhdmProductionImmutable",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmProductionImmutable,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			false,
		},
		{
			"RhdmAuthoring",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmAuthoring,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			false,
		},
		{
			"RhdmAuthoringHA",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmAuthoringHA,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
			},
			false,
		},
		{
			"RhpamTrial_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
					},
				},
			},
			false,
		},
		{
			"RhpamProduction_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamProduction,
					},
				},
			},
			false,
		},
		{
			"RhpamProductionImmutable_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamProductionImmutable,
					},
				},
			},
			false,
		},
		{
			"RhpamAuthoring_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamAuthoring,
					},
				},
			},
			false,
		},
		{
			"RhpamAuthoringHA_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamAuthoringHA,
					},
				},
			},
			false,
		},
		{
			"RhdmTrial_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmTrial,
					},
				},
			},
			false,
		},
		{
			"RhdmProductionImmutable_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmProductionImmutable,
					},
				},
			},
			false,
		},
		{
			"RhdmAuthoring_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmAuthoring,
					},
				},
			},
			false,
		},
		{
			"RhdmAuthoringHA_NoProcessMigration",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhdmAuthoringHA,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaults(tt.args.cr)
			if got := deployProcessMigration(tt.args.cr); got != tt.want {
				t.Errorf("deployProcessMigration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProcessMigrationTemplate(t *testing.T) {
	type args struct {
		cr            *api.KieApp
		serversConfig []api.ServerTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *api.ProcessMigrationTemplate
		wantErr bool
	}{
		{
			"ProcessMigration_Custom",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{
								Image:    "test-pim-image",
								ImageTag: "test-pim-image-tag",
								Database: api.ProcessMigrationDatabaseObject{
									InternalDatabaseObject: api.InternalDatabaseObject{
										Type:             api.DatabaseMySQL,
										StorageClassName: "gold",
										Size:             "32Gi",
									},
								},
							},
						},
						CommonConfig: api.CommonConfig{
							AdminUser:     "testuser",
							AdminPassword: "testpassword",
						},
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
					{KieName: "kieserver2"},
				},
			},
			&api.ProcessMigrationTemplate{
				Image:    "test-pim-image",
				ImageTag: "test-pim-image-tag",
				ImageURL: "test-pim-image:test-pim-image-tag",
				KieServerClients: []api.KieServerClient{
					{
						Host:     "http://kieserver1:8080/services/rest/server",
						Username: "testuser",
						Password: "testpassword",
					},
					{
						Host:     "http://kieserver2:8080/services/rest/server",
						Username: "testuser",
						Password: "testpassword",
					},
				},
				Database: api.ProcessMigrationDatabaseObject{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type:             api.DatabaseMySQL,
						StorageClassName: "gold",
						Size:             "32Gi",
					},
				},
			},
			false,
		},
		{
			"ProcessMigration_Default",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
				},
			},
			&api.ProcessMigrationTemplate{
				Image:    constants.RhpamPrefix + "-process-migration" + constants.RhelVersion,
				ImageTag: constants.CurrentVersion,
				ImageURL: constants.RhpamPrefix + "-process-migration" + constants.RhelVersion + ":" + constants.CurrentVersion,
				KieServerClients: []api.KieServerClient{
					{
						Host:     "http://kieserver1:8080/services/rest/server",
						Username: constants.DefaultAdminUser,
						Password: constants.DefaultPassword,
					},
				},
				Database: api.ProcessMigrationDatabaseObject{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type: api.DatabaseH2,
					},
				},
			},
			false,
		},
		{
			"ProcessMigration_UnsupportedVersion",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{},
						},
						Version: "7.7.1",
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
				},
			},
			nil,
			false,
		},
		{
			"ProcessMigration_Empty",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						CommonConfig: api.CommonConfig{
							AdminUser:     "testuser",
							AdminPassword: "testpassword",
						},
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
				},
			},
			nil,
			false,
		},
		{
			"ProcessMigration_ExternalDB_Error",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{
								Image:    "test-pim-image",
								ImageTag: "test-pim-image-tag",
								Database: api.ProcessMigrationDatabaseObject{
									InternalDatabaseObject: api.InternalDatabaseObject{
										Type: api.DatabaseExternal,
									},
								},
							},
						},
						CommonConfig: api.CommonConfig{
							AdminUser:     "testuser",
							AdminPassword: "testpassword",
						},
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
					{KieName: "kieserver2"},
				},
			},
			nil,
			true,
		},
		{
			"ProcessMigration_ExternalDB",
			args{
				&api.KieApp{
					Spec: api.KieAppSpec{
						Environment: api.RhpamTrial,
						Objects: api.KieAppObjects{
							ProcessMigration: &api.ProcessMigrationObject{
								Image:    "test-pim-image",
								ImageTag: "test-pim-image-tag",
								Database: api.ProcessMigrationDatabaseObject{
									InternalDatabaseObject: api.InternalDatabaseObject{
										Type: api.DatabaseExternal,
									},
									ExternalConfig: &api.CommonExternalDatabaseObject{
										Driver:                     "mariadb",
										JdbcURL:                    "jdbc:mariadb://hello-mariadb:3306/pimdb",
										Username:                   "pim",
										Password:                   "pim",
										MinPoolSize:                "10",
										MaxPoolSize:                "10",
										ConnectionChecker:          "org.jboss.jca.adapters.jdbc.extensions.mysql.MySQLValidConnectionChecker",
										ExceptionSorter:            "org.jboss.jca.adapters.jdbc.extensions.mysql.MySQLExceptionSorter",
										BackgroundValidation:       "true",
										BackgroundValidationMillis: "150000",
									},
								},
							},
						},
						CommonConfig: api.CommonConfig{
							AdminUser:     "testuser",
							AdminPassword: "testpassword",
						},
					},
				},
				[]api.ServerTemplate{
					{KieName: "kieserver1"},
					{KieName: "kieserver2"},
				},
			},
			&api.ProcessMigrationTemplate{
				Image:    "test-pim-image",
				ImageTag: "test-pim-image-tag",
				ImageURL: "test-pim-image:test-pim-image-tag",
				KieServerClients: []api.KieServerClient{
					{
						Host:     "http://kieserver1:8080/services/rest/server",
						Username: "testuser",
						Password: "testpassword",
					},
					{
						Host:     "http://kieserver2:8080/services/rest/server",
						Username: "testuser",
						Password: "testpassword",
					},
				},
				Database: api.ProcessMigrationDatabaseObject{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type: api.DatabaseExternal,
					},
					ExternalConfig: &api.CommonExternalDatabaseObject{
						Driver:                     "mariadb",
						JdbcURL:                    "jdbc:mariadb://hello-mariadb:3306/pimdb",
						Username:                   "pim",
						Password:                   "pim",
						MinPoolSize:                "10",
						MaxPoolSize:                "10",
						ConnectionChecker:          "org.jboss.jca.adapters.jdbc.extensions.mysql.MySQLValidConnectionChecker",
						ExceptionSorter:            "org.jboss.jca.adapters.jdbc.extensions.mysql.MySQLExceptionSorter",
						BackgroundValidation:       "true",
						BackgroundValidationMillis: "150000",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaults(tt.args.cr)
			got, err := getProcessMigrationTemplate(tt.args.cr, tt.args.serversConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("getProcessMigrationTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProcessMigrationTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeProcessMigrationDB(t *testing.T) {
	type args struct {
		service     kubernetes.PlatformService
		cr          *api.KieApp
		env         api.Environment
		envTemplate api.EnvTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    api.Environment
		wantErr bool
	}{
		{
			"ProcessMigration_ExternalDB",
			args{
				test.MockService(),
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
					Spec: api.KieAppSpec{
						Version: constants.CurrentVersion,
					},
				},
				api.Environment{},
				api.EnvTemplate{
					CommonConfig: &api.CommonConfig{
						ApplicationName: "kietest",
					},
					ProcessMigration: api.ProcessMigrationTemplate{
						Database: api.ProcessMigrationDatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
							ExternalConfig: &api.CommonExternalDatabaseObject{
								Driver:   "mariadb",
								JdbcURL:  "jdbc:mariadb://test-process-migration-mysql:3306/pimdb",
								Username: "pim",
								Password: "pim",
							},
						},
					},
				},
			},
			api.Environment{
				ProcessMigration: api.CustomObject{
					ConfigMaps: []corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "kietest-process-migration",
							},
						},
					},
				},
			},
			false,
		},
		{
			"ProcessMigration_H2",
			args{
				test.MockService(),
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
					Spec: api.KieAppSpec{
						Version: constants.CurrentVersion,
					},
				},
				api.Environment{},
				api.EnvTemplate{
					CommonConfig: &api.CommonConfig{
						ApplicationName: "kietest",
					},
					ProcessMigration: api.ProcessMigrationTemplate{
						Database: api.ProcessMigrationDatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseH2,
							},
						},
					},
				},
			},
			api.Environment{},
			false,
		},
		{
			"ProcessMigration_MySQL",
			args{
				test.MockService(),
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
					Spec: api.KieAppSpec{
						Version: constants.CurrentVersion,
					},
				},
				api.Environment{},
				api.EnvTemplate{
					CommonConfig: &api.CommonConfig{
						ApplicationName: "kietest",
					},
					ProcessMigration: api.ProcessMigrationTemplate{
						Database: api.ProcessMigrationDatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseMySQL,
							},
						},
					},
				},
			},
			api.Environment{
				ProcessMigration: api.CustomObject{
					DeploymentConfigs: []appsv1.DeploymentConfig{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "kietest-process-migration",
							},
						},
					},
					ConfigMaps: []corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "kietest-process-migration",
							},
						},
					},
				},
			},
			false,
		},
		{
			"ProcessMigration_POSTGRESSQL",
			args{
				test.MockService(),
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
					Spec: api.KieAppSpec{
						Version: constants.CurrentVersion,
					},
				},
				api.Environment{},
				api.EnvTemplate{
					CommonConfig: &api.CommonConfig{
						ApplicationName: "kietest",
					},
					ProcessMigration: api.ProcessMigrationTemplate{
						Database: api.ProcessMigrationDatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabasePostgreSQL,
							},
						},
					},
				},
			},
			api.Environment{
				ProcessMigration: api.CustomObject{
					DeploymentConfigs: []appsv1.DeploymentConfig{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "kietest-process-migration",
							},
						},
					},
					ConfigMaps: []corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "kietest-process-migration",
							},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaults(tt.args.cr)
			got, err := mergeProcessMigrationDB(tt.args.service, tt.args.cr, tt.args.env, tt.args.envTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeProcessMigrationDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.ProcessMigration, tt.want.ProcessMigration) {
				if len(got.ProcessMigration.DeploymentConfigs) != len(tt.want.ProcessMigration.DeploymentConfigs) ||
					(len(tt.want.ProcessMigration.DeploymentConfigs) == 1 &&
						got.ProcessMigration.DeploymentConfigs[0].ObjectMeta.Name != tt.want.ProcessMigration.DeploymentConfigs[0].ObjectMeta.Name) {
					t.Errorf("mergeProcessMigrationDB() got = %v, want %v", got, tt.want)
					return
				}

				if len(got.ProcessMigration.ConfigMaps) != len(tt.want.ProcessMigration.ConfigMaps) ||
					(len(tt.want.ProcessMigration.ConfigMaps) == 1 &&
						got.ProcessMigration.ConfigMaps[0].ObjectMeta.Name != tt.want.ProcessMigration.ConfigMaps[0].ObjectMeta.Name) {
					t.Errorf("mergeProcessMigrationDB() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetDatabaseDeploymentTemplate(t *testing.T) {
	type args struct {
		cr                       *api.KieApp
		serversConfig            []api.ServerTemplate
		processMigrationTemplate *api.ProcessMigrationTemplate
	}
	tests := []struct {
		name string
		args args
		want []api.DatabaseTemplate
	}{
		{
			"KieServerDeployment",
			args{
				&api.KieApp{},
				[]api.ServerTemplate{
					{
						KieName: "mysql",
						Database: api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type:             api.DatabaseMySQL,
								Size:             "30Gi",
								StorageClassName: "gold",
							},
						},
					},
					{
						KieName: "postgresql",
						Database: api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type:             api.DatabasePostgreSQL,
								Size:             "20Gi",
								StorageClassName: "gold1",
							},
						},
					},
					{
						KieName: "h2",
						Database: api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseH2,
							},
						},
					},
					{
						KieName: "external",
						Database: api.DatabaseObject{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type: api.DatabaseExternal,
							},
							ExternalConfig: &api.ExternalDatabaseObject{},
						},
					},
				},
				nil,
			},
			[]api.DatabaseTemplate{
				{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type:             api.DatabaseMySQL,
						Size:             "30Gi",
						StorageClassName: "gold",
					},
					ServerName:   "mysql",
					DatabaseName: constants.DefaultKieServerDatabaseName,
					Username:     constants.DefaultKieServerDatabaseUsername,
				},
				{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type:             api.DatabasePostgreSQL,
						Size:             "20Gi",
						StorageClassName: "gold1",
					},
					ServerName:   "postgresql",
					DatabaseName: constants.DefaultKieServerDatabaseName,
					Username:     constants.DefaultKieServerDatabaseUsername,
				},
			},
		},
		{
			"ProcessMigrationDeployment_Mysql",
			args{
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysql",
					},
				},
				nil,
				&api.ProcessMigrationTemplate{
					Database: api.ProcessMigrationDatabaseObject{
						InternalDatabaseObject: api.InternalDatabaseObject{
							Type:             api.DatabaseMySQL,
							Size:             "30Gi",
							StorageClassName: "gold",
						},
					},
				},
			},
			[]api.DatabaseTemplate{
				{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type:             api.DatabaseMySQL,
						Size:             "30Gi",
						StorageClassName: "gold",
					},
					ServerName:   "mysql-process-migration",
					DatabaseName: constants.DefaultProcessMigrationDatabaseName,
					Username:     constants.DefaultProcessMigrationDatabaseUsername,
				},
			},
		},
		{
			"ProcessMigrationDeployment_Postgresql",
			args{
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "postgresql",
					},
				},
				nil,
				&api.ProcessMigrationTemplate{
					Database: api.ProcessMigrationDatabaseObject{
						InternalDatabaseObject: api.InternalDatabaseObject{
							Type:             api.DatabasePostgreSQL,
							Size:             "30Gi",
							StorageClassName: "gold",
						},
					},
				},
			},
			[]api.DatabaseTemplate{
				{
					InternalDatabaseObject: api.InternalDatabaseObject{
						Type:             api.DatabasePostgreSQL,
						Size:             "30Gi",
						StorageClassName: "gold",
					},
					ServerName:   "postgresql-process-migration",
					DatabaseName: constants.DefaultProcessMigrationDatabaseName,
					Username:     constants.DefaultProcessMigrationDatabaseUsername,
				},
			},
		},
		{
			"ProcessMigrationDeployment_h2",
			args{
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "h2",
					},
				},
				nil,
				&api.ProcessMigrationTemplate{
					Database: api.ProcessMigrationDatabaseObject{
						InternalDatabaseObject: api.InternalDatabaseObject{
							Type: api.DatabaseH2,
						},
					},
				},
			},
			nil,
		},
		{
			"ProcessMigrationDeployment_external",
			args{
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "external",
					},
				},
				nil,
				&api.ProcessMigrationTemplate{
					Database: api.ProcessMigrationDatabaseObject{
						InternalDatabaseObject: api.InternalDatabaseObject{
							Type: api.DatabaseExternal,
						},
						ExternalConfig: &api.CommonExternalDatabaseObject{},
					},
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDatabaseDeploymentTemplate(tt.args.cr, tt.args.serversConfig, tt.args.processMigrationTemplate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDatabaseDeploymentTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeDBDeployment(t *testing.T) {
	type args struct {
		service     kubernetes.PlatformService
		cr          *api.KieApp
		env         api.Environment
		envTemplate api.EnvTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    api.Environment
		wantErr bool
	}{
		{
			"MergeDBDeployment",
			args{
				test.MockService(),
				&api.KieApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
					Spec: api.KieAppSpec{
						Version: constants.CurrentVersion,
					},
				},
				api.Environment{},
				api.EnvTemplate{
					CommonConfig: &api.CommonConfig{
						ApplicationName: "kietest",
					},
					Databases: []api.DatabaseTemplate{
						{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type:             api.DatabaseMySQL,
								Size:             "30Gi",
								StorageClassName: "gold",
							},
							ServerName:   "mysql",
							DatabaseName: constants.DefaultKieServerDatabaseName,
							Username:     constants.DefaultKieServerDatabaseUsername,
						},
						{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type:             api.DatabasePostgreSQL,
								Size:             "20Gi",
								StorageClassName: "gold1",
							},
							ServerName:   "postgresql",
							DatabaseName: constants.DefaultKieServerDatabaseName,
							Username:     constants.DefaultKieServerDatabaseUsername,
						},
						{
							InternalDatabaseObject: api.InternalDatabaseObject{
								Type:             api.DatabaseMySQL,
								Size:             "10Gi",
								StorageClassName: "gold2",
							},
							ServerName:   "mysql-process-migration",
							DatabaseName: constants.DefaultProcessMigrationDatabaseName,
							Username:     constants.DefaultProcessMigrationDatabaseUsername,
						},
					},
				},
			},
			api.Environment{
				Databases: []api.CustomObject{
					{
						DeploymentConfigs: []appsv1.DeploymentConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-mysql",
								},
							},
						},
						PersistentVolumeClaims: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-mysql-claim",
								},
							},
						},
						Services: []corev1.Service{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-mysql",
								},
							},
						},
					},
					{
						DeploymentConfigs: []appsv1.DeploymentConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "postgresql-postgresql",
								},
							},
						},
						PersistentVolumeClaims: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "postgresql-postgresql-claim",
								},
							},
						},
						Services: []corev1.Service{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "postgresql-postgresql",
								},
							},
						},
					},
					{
						DeploymentConfigs: []appsv1.DeploymentConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-process-migration-mysql",
								},
							},
						},
						PersistentVolumeClaims: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-process-migration-mysql-claim",
								},
							},
						},
						Services: []corev1.Service{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "mysql-process-migration-mysql",
								},
							},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaults(tt.args.cr)
			got, err := mergeDBDeployment(tt.args.service, tt.args.cr, tt.args.env, tt.args.envTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeDBDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				if len(got.Databases) != len(tt.want.Databases) {
					t.Errorf("mergeDBDeployment() got = %v, want %v", got, tt.want)
					return
				}

				for i := range got.Databases {
					if len(got.Databases[i].DeploymentConfigs) != len(tt.want.Databases[i].DeploymentConfigs) ||
						(len(tt.want.Databases[i].DeploymentConfigs) == 1 &&
							got.Databases[i].DeploymentConfigs[0].ObjectMeta.Name != tt.want.Databases[i].DeploymentConfigs[0].ObjectMeta.Name) {
						t.Errorf("mergeDBDeployment() got = %v, want %v", got, tt.want)
						return
					}

					if len(got.Databases[i].PersistentVolumeClaims) != len(tt.want.Databases[i].PersistentVolumeClaims) ||
						(len(tt.want.Databases[i].PersistentVolumeClaims) == 1 &&
							got.Databases[i].PersistentVolumeClaims[0].ObjectMeta.Name != tt.want.Databases[i].PersistentVolumeClaims[0].ObjectMeta.Name) {
						t.Errorf("mergeDBDeployment() got = %v, want %v", got, tt.want)
						return
					}

					if len(got.Databases[i].Services) != len(tt.want.Databases[i].Services) ||
						(len(tt.want.Databases[i].Services) == 1 &&
							got.Databases[i].Services[0].ObjectMeta.Name != tt.want.Databases[i].Services[0].ObjectMeta.Name) {
						t.Errorf("mergeDBDeployment() got = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}
