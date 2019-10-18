package defaults

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/kiegroup/kie-cloud-operator/version"
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
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Version, cr.Spec.Environment, cr.Name), err.Error())
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
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Version, cr.Spec.Environment, cr.Name), err.Error())
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, deployments), env.Servers[deployments-1].DeploymentConfigs[0].Name)
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, deployments), cr.Spec.Objects.Servers[deployments-1].Name)
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Spec.CommonConfig.ApplicationName, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhdm-decisioncentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

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

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

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
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/jboss-datagrid-7/%s:%s", constants.VersionConstants[cr.Spec.Version].DatagridImage, constants.VersionConstants[cr.Spec.Version].DatagridImageTag), env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/amq7/%s:%s", constants.VersionConstants[cr.Spec.Version].BrokerImage, constants.VersionConstants[cr.Spec.Version].BrokerImageTag), env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "rhdm-decisioncentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
				AMQPassword:        "amq",
				AMQClusterPassword: "cluster",
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
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/jboss-datagrid-7/%s:%s", constants.VersionConstants[cr.Spec.Version].DatagridImage, constants.VersionConstants[cr.Spec.Version].DatagridImageTag), env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/amq7/%s:%s", constants.VersionConstants[cr.Spec.Version].BrokerImage, constants.VersionConstants[cr.Spec.Version].BrokerImageTag), env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "rhpam-businesscentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	amqClusterPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_JMS_BROKER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	amqPassword := getEnvVariable(env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0], "AMQ_PASSWORD")
	assert.Equal(t, "amq", amqPassword, "Expected provided password to take effect, but found %v", amqPassword)
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
	assert.Equal(t, "rhdm-decisioncentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhdmProdImmutableJMSEnvironment(t *testing.T) {
	f := false
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
							AuditTransacted:    &f,
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
	assert.Equal(t, "rhdm-decisioncentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestRhpamProdImmutableJMSEnvironment(t *testing.T) {
	f := false
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
							AuditTransacted:    &f,
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
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[2].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[2].Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

}

func TestRhpamProdImmutableJMSEnvironmentWithSSL(t *testing.T) {
	f := false
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
							AuditTransacted:       &f,
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
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[2].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.False(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	assert.Equal(t, "amq-tcp-ssl", env.Servers[0].Routes[2].Name)
	assert.False(t, env.Servers[0].Routes[2].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[2].Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, true, cr.Spec.Objects.Servers[0].Jms.AMQEnableSSL)
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

}

func TestRhpamProdImmutableJMSEnvironmentExecutorDisabled(t *testing.T) {
	f := false
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
							Executor:           &f,
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

	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[2].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_EXECUTOR_JMS"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_EXECUTOR_JMS_TRANSACTED"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_ENABLE_AUDIT"), "Variable should exist")
	assert.Equal(t, "queue/CUSTOM.KIE.SERVER.AUDIT", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_QUEUE_AUDIT"), "Variable should exist")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_ENABLE_SIGNAL"), "Variable should exist")
	assert.Equal(t, "queue/CUSTOM.KIE.SERVER.SIGNAL", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_JMS_QUEUE_SIGNAL"), "Variable should exist")
	assert.Equal(t, "queue/KIE.SERVER.REQUEST, queue/KIE.SERVER.RESPONSE, queue/CUSTOM.KIE.SERVER.SIGNAL, queue/CUSTOM.KIE.SERVER.AUDIT", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "AMQ_QUEUES"), "Variable should exist")

	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

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
							Webhooks: []api.WebhookSecret{
								{
									Type:   api.GitHubWebhook,
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

	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Spec.CommonConfig.ApplicationName), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
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
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Spec.CommonConfig.ApplicationName), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.NotEqual(t, api.Environment{}, env, "Environment should not be empty")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
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
	assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
	assert.Equal(t, fmt.Sprintf("rhpam-smartrouter-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.SmartRouter.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
		assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
			assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
		assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
	assert.Equal(t, fmt.Sprintf("%s/envs/%s.yaml does not exist, '%s' KieApp not deployed", cr.Spec.Version, cr.Spec.Environment, cr.Name), err.Error())

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
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						SecuredKieAppObject: api.SecuredKieAppObject{
							KieAppObject: api.KieAppObject{
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
	assert.Equal(t, fmt.Sprintf("rhpam-businesscentral-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
						SecuredKieAppObject: api.SecuredKieAppObject{
							KieAppObject: api.KieAppObject{
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
	for index := 0; index < 1; index++ {
		s := env.Servers[index]
		assert.Equal(t, fmt.Sprintf("rhpam-kieserver-rhel8:%s", cr.Spec.CommonConfig.ImageTag), s.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
				Console: api.SecuredKieAppObject{
					KieAppObject: api.KieAppObject{
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
	assert.Equal(t, fmt.Sprintf("rhdm-decisioncentral-rhel8:%s", cr.Spec.CommonConfig.ImageTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	adminUser := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_USER")
	assert.Equal(t, constants.DefaultAdminUser, adminUser, "AdminUser default not being set correctly")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
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

	assert.NotContains(t, cr.Spec.Objects.Console.Env, corev1.EnvVar{
		Name: "empty",
	})
	assert.False(t, cr.Spec.Upgrades.Enabled, "Spec.Upgrades.Enabled should be false by default")
	assert.False(t, cr.Spec.Upgrades.Minor, "Spec.Upgrades.Minor should be false by default")
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
	defer os.Unsetenv("REGISTRY")
	os.Setenv("INSECURE", "true")
	defer os.Unsetenv("INSECURE")
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
}

func buildKieApp(name string, deployments int) *api.KieApp {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: api.SecuredKieAppObject{
					KieAppObject: api.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources,
					},
				},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						SecuredKieAppObject: api.SecuredKieAppObject{
							KieAppObject: api.KieAppObject{
								Env:       sampleEnv,
								Resources: sampleResources,
							},
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
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting authoring environment")
	adminUser := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_USER")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, cr.Spec.CommonConfig.AdminUser, adminUser, "Expected provided user to take effect, but found %v", adminUser)
	assert.Equal(t, cr.Spec.CommonConfig.AdminPassword, adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Len(t, mavenPassword, 8, "Expected a password with length of 8 to be generated, but got %v", mavenPassword)

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
	assert.Equal(t, "RedHat", mavenPassword, "Expected default password of RedHat, but found %v", mavenPassword)

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
	assert.Equal(t, constants.DefaultKieDeployments, *cr.Spec.Objects.Servers[0].Deployments, "Default number of kieserver deployments not being set in CR")
	assert.Len(t, cr.Spec.Objects.Servers, 1, "There should be 1 custom kieserver being set by default")
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
	assert.Equal(t, env.Servers[0].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], cr.Spec.Objects.Servers[0].Name)
	assert.Equal(t, env.Servers[1].DeploymentConfigs[0].Labels["services.server.kie.org/kie-server-id"], strings.Join([]string{cr.Spec.Objects.Servers[0].Name, "2"}, "-"))
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
	assert.Equal(t, fmt.Sprintf("rhdm-kieserver-rhel8:%v", cr.Spec.CommonConfig.ImageTag), env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
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
							Type: api.DatabaseExternal,
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
							Type: api.DatabaseExternal,
							ExternalConfig: &api.ExternalDatabaseObject{
								Dialect:              "org.hibernate.dialect.Oracle10gDialect",
								Driver:               "oracle",
								ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker",
								ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter",
								BackgroundValidation: "false",
								Username:             "oracleUser",
								Password:             "oraclePwd",
								JdbcURL:              "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7",
							},
						},
						SecuredKieAppObject: api.SecuredKieAppObject{
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
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	env = ConsolidateObjects(env, cr)

	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabaseH2,
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, constants.CurrentVersion, cr.Spec.Version)
	assert.Equal(t, constants.VersionConstants[constants.CurrentVersion].ImageTag, cr.Spec.CommonConfig.ImageTag)
	assert.True(t, checkVersion(cr.Spec.Version))
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, fmt.Sprintf("Product version %s is not allowed in operator version %s. The following versions are allowed - %s", cr.Spec.Version, version.Version, constants.SupportedVersions), err.Error())
	assert.Equal(t, "6.3.1", cr.Spec.Version)
	major, minor, micro := MajorMinorMicro(cr.Spec.Version)
	assert.Equal(t, "6", major)
	assert.Equal(t, "3", minor)
	assert.Equal(t, "1", micro)
	assert.False(t, checkVersion(cr.Spec.Version))
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
	filename := strings.Join([]string{cr.Spec.Version, filepath}, "/")
	cmNameT, fileT := convertToConfigMapName(filename)

	fileslice := strings.Split(filepath, "/")
	file := fileslice[len(fileslice)-1]
	assert.Equal(t, file, fileT)

	cmName := strings.Join([]string{constants.ConfigMapPrefix, cr.Spec.Version, fileslice[0]}, "-")
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
							Type: api.DatabaseH2,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabaseMySQL,
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabaseMySQL,
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabaseMySQL,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabasePostgreSQL,
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabasePostgreSQL,
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
	assert.Equal(t, "rhpam-businesscentral-monitoring-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
							Type: api.DatabasePostgreSQL,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting trial environment")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "rhpam-businesscentral-rhel8", env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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

func TestCustomImageTag(t *testing.T) {
	image := "testing-images"
	imageTag := "1.5"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				Console: api.SecuredKieAppObject{},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(2),
					},
				},
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	cr.Spec.Objects.Console.Image = image
	cr.Spec.Objects.Console.ImageTag = imageTag
	cr.Spec.Objects.Servers[0].Image = image
	cr.Spec.Objects.Servers[0].ImageTag = imageTag
	cr.Spec.Objects.SmartRouter.Image = image
	cr.Spec.Objects.SmartRouter.ImageTag = imageTag
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 2, "Expect two KIE Servers to be created based on provided build configs")

	imageName := strings.Join([]string{image, imageTag}, ":")
	assert.Equal(t, getImageChangeName(env.Console.DeploymentConfigs[0]), imageName)
	assert.Equal(t, getImageChangeName(env.Servers[0].DeploymentConfigs[0]), imageName)
	assert.Equal(t, getImageChangeName(env.Servers[1].DeploymentConfigs[0]), imageName)
	assert.Equal(t, getImageChangeName(env.SmartRouter.DeploymentConfigs[0]), imageName)
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
