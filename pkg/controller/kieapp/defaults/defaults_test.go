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
	"k8s.io/apimachinery/pkg/util/intstr"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	bcmImage             = constants.ImageRegistry + "/" + constants.RhpamPrefix + "-7/" + constants.RhpamPrefix + "-businesscentral-monitoring" + constants.RhelVersion
	bcImage              = constants.ImageRegistry + "/" + constants.RhpamPrefix + "-7/" + constants.RhpamPrefix + "-businesscentral" + constants.RhelVersion
	dcImage              = constants.ImageRegistry + "/" + constants.RhdmPrefix + "-7/" + constants.RhdmPrefix + "-decisioncentral" + constants.RhelVersion
	dashImage            = constants.ImageRegistry + "/" + constants.RhpamPrefix + "-7/" + constants.RhpamPrefix + "-dashbuilder" + constants.RhelVersion
	rhpamkieServerImage  = constants.ImageRegistry + "/" + constants.RhpamPrefix + "-7/" + constants.RhpamPrefix + "-kieserver" + constants.RhelVersion
	rhdmkieServerImage   = constants.ImageRegistry + "/" + constants.RhdmPrefix + "-7/" + constants.RhdmPrefix + "-kieserver" + constants.RhelVersion
	latestTag            = ":latest"
	helloRules           = "hello-rules" + latestTag
	byeRules             = "bye-rules" + latestTag
	kieServerName        = "test-kieserver"
	rhpamKieserverAndTag = "rhpam-kieserver-rhel8:%s"
	pimImage             = constants.RhpamPrefix + "-process-migration" + constants.RhelVersion
	bcKeySecret          = fmt.Sprintf(constants.KeystoreSecret, "test-businesscentral")
)

const (
	bcKeystoreVolume         = "/etc/businesscentral-secret-volume"
	bcKeyStoreVolumeName     = "test-rhpamcentr-keystore-volume"
	bcHttpsRouteDescription  = "Route for Business Central's https service."
	bcHttpRouteDescription   = "Route for Business Central's http service."
	dashHttpsRouteDescrition = "Route for Dashbuilder's https service."
	dashHttpRouteDescrition  = "Route for Dashbuilder's http service."
	dashKeyStoreVolume       = "/etc/dashbuilder-secret-volume"
	dashKeyStoreVolumeName   = "test-dash-rhpamdash-keystore-volume"
	dashName                 = "test-dash-rhpamdash"
	dcKeyStoreVolumeName     = "test-rhdmcentr-keystore-volume"
	ksHttpsRouteDescription  = "Route for KIE server's https service."
	ksHttpRouteDescription   = "Route for KIE server's http service."
	routeBalanceAnnotation   = "haproxy.router.openshift.io/balance"
	smartrouterKeyStore      = "smartrouter-keystore-volume"
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
	mockService.GetFunc = func(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
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
	runTrialEnvironmentTests(t, "rhpamcentr", api.RhpamTrial, bcImage)
}

func TestRHDMTrialEnvironment(t *testing.T) {
	runTrialEnvironmentTests(t, "rhdmcentr", api.RhdmTrial, dcImage)
}

func runTrialEnvironmentTests(t *testing.T, consoleName string, environment api.EnvironmentType, consoleImage string) {
	deployments := 2
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: environment,
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
	mainService := getService(wbServices, "test-"+consoleName)
	assert.NotNil(t, mainService, consoleName+" service not found")
	assert.Len(t, mainService.Spec.Ports, 2, "The "+consoleName+" service should have two ports")
	assert.False(t, hasPort(mainService, 8001), "The "+consoleName+" service should NOT listen on port 8001")

	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, len(env.Servers)), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, "test-"+consoleName, env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, consoleImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, getLivenessReadiness("/rest/ready"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/rest/healthy"), env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)

	routeAnnotations := getRouteAnnotations(bcHttpsRouteDescription)

	assert.Equal(t, 2, len(env.Console.Routes))
	assert.Equal(t, "test-"+consoleName, env.Console.Routes[0].Name)
	assert.NotNil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Console.Routes[0].Spec.Port.TargetPort)

	routeAnnotations["description"] = bcHttpRouteDescription
	assert.Equal(t, "test-"+consoleName+"-http", env.Console.Routes[1].Name)
	assert.Nil(t, env.Console.Routes[1].Spec.TLS)
	delete(routeAnnotations, routeBalanceAnnotation)
	assert.Equal(t, routeAnnotations, env.Console.Routes[1].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Console.Routes[1].Spec.Port.TargetPort)

	assertContainBCAndKSVolumes(t, "test-"+consoleName+"-keystore-volume", env)

	assert.Equal(t, 2, len(env.Servers[0].Routes))
	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.NotNil(t, env.Servers[0].Routes[0].Spec.TLS)
	routeAnnotations["description"] = ksHttpsRouteDescription
	routeAnnotations[routeBalanceAnnotation] = "source"
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)

	assert.Equal(t, kieServerName+"-http", env.Servers[0].Routes[1].Name)
	assert.Nil(t, env.Servers[0].Routes[1].Spec.TLS)
	routeAnnotations["description"] = ksHttpRouteDescription
	delete(routeAnnotations, routeBalanceAnnotation)
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[1].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Servers[0].Routes[1].Spec.Port.TargetPort)

}

func TestRHPAMDashbuilderDefaultEnvironment(t *testing.T) {

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting dashbuilder default environment")

	commonDashbuilderAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(dashHttpsRouteDescrition)

	assert.Equal(t, 1, len(env.Dashbuilder.Routes))
	assert.Equal(t, dashName, env.Dashbuilder.Routes[0].Name)
	assert.NotNil(t, env.Dashbuilder.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Dashbuilder.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Dashbuilder.Routes[0].Spec.Port.TargetPort)

	dashVolumeMountSecret, dashVolume := getDashKeyAndVolume()

	assert.Contains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, dashVolumeMountSecret)
	assert.Contains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Volumes, dashVolume)

	// ssl envs
	assertHTTPSEnvs(t, dashKeyStoreVolume, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0])

}

func TestRHPAMDashbuilderDefaultEnvironmentWithSSLDisabled(t *testing.T) {

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
			CommonConfig: api.CommonConfig{
				DisableSsl: true,
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting dashbuilder default with ssl disabled environment")

	commonDashbuilderAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(dashHttpRouteDescrition)
	delete(routeAnnotations, routeBalanceAnnotation)

	assert.Equal(t, 1, len(env.Dashbuilder.Routes))
	assert.Equal(t, dashName, env.Dashbuilder.Routes[0].Name)
	assert.Nil(t, env.Dashbuilder.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Dashbuilder.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Dashbuilder.Routes[0].Spec.Port.TargetPort)

	dashVolumeMountSecret, dashVolume := getDashKeyAndVolume()

	assert.NotContains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, dashVolumeMountSecret)
	assert.NotContains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Volumes, dashVolume)

	// ssl envs
	assertHTTPEmpty(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0])
}

func getDashKeyAndVolume() (corev1.VolumeMount, corev1.Volume) {
	dashVolumeMountSecret, _ := getVolumeMountSecret(dashKeyStoreVolumeName, dashKeyStoreVolume)
	dashVolume, _ := getVolumes(dashKeyStoreVolumeName, "test-dash-dashbuilder-app-secret")
	return dashVolumeMountSecret, dashVolume
}

func assertHTTPEmpty(t *testing.T, container corev1.Container) {
	assert.Empty(t, getEnvVariable(container, "HTTPS_KEYSTORE_DIR"))
	assert.Empty(t, getEnvVariable(container, "HTTPS_KEYSTORE"))
	assert.Empty(t, getEnvVariable(container, "HTTPS_NAME"))
	assert.Empty(t, getEnvVariable(container, "HTTPS_PASSWORD"))
}

func assertHTTPSEnvs(t *testing.T, keystoreVolumeName string, container corev1.Container) {
	assert.Equal(t, keystoreVolumeName, getEnvVariable(container, "HTTPS_KEYSTORE_DIR"))
	assert.Equal(t, constants.KeystoreName, getEnvVariable(container, "HTTPS_KEYSTORE"))
	assert.Equal(t, "jboss", getEnvVariable(container, "HTTPS_NAME"))
	assert.Empty(t, "", getEnvVariable(container, "HTTPS_PASSWORD"))
}

func commonDashbuilderAssertions(t *testing.T, env api.Environment, cr *api.KieApp) {
	dashServices := env.Dashbuilder.Services
	mainService := getService(dashServices, dashName)

	assert.NotNil(t, mainService, "rhpamdash service not found")
	assert.Len(t, mainService.Spec.Ports, 2, "The rhpamdash service should have two ports")
	assert.False(t, hasPort(mainService, 8001), "The rhpamdash service should NOT listen on port 8001")
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Dashbuilder.Replicas)

	assert.Equal(t, dashName, env.Dashbuilder.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, dashImage+":"+cr.Status.Applied.Version, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, getLivenessReadiness("/rest/ready"), env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/rest/healthy"), env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)

	assert.NotNil(t, cr.Status.Applied.Objects.Dashbuilder.Resources)
	assert.Equal(t, "1", cr.Status.Applied.Objects.Dashbuilder.Resources.Limits.Cpu().String())
	assert.Equal(t, "750m", cr.Status.Applied.Objects.Dashbuilder.Resources.Requests.Cpu().String())
	assert.Equal(t, "2Gi", cr.Status.Applied.Objects.Dashbuilder.Resources.Limits.Memory().String())
	assert.Equal(t, "1536Mi", cr.Status.Applied.Objects.Dashbuilder.Resources.Requests.Memory().String())

	checkClusterLabels(t, cr, env.Dashbuilder)
	checkObjectLabels(t, cr, env.Dashbuilder, "PAM", "rhpam-dashbuilder-rhel8")
}

func TestRHPAMDashbuilderEnvironmentWithCustomProperties(t *testing.T) {
	tr := true
	f := false
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Config: &api.DashbuilderConfig{
						ImportFileLocation:        "/test/testing",
						PersistentConfigs:         &tr,
						AllowExternalFileRegister: &f,
						ConfigMapProps:            "/tmp/configMap.properties",
					},
				},
			},
		},
	}

	shouldNotContainEnvs := []string{"DASHBUILDER_EXTERNAL_COMP_DIR", "DASHBUILDER_COMP_ENABLE",
		"DASHBUILDER_UPLOAD_SIZE", "DASHBUILDER_RUNTIME_MULTIPLE_IMPORT", "DASHBUILDER_MODEL_FILE_REMOVAL",
		"DASHBUILDER_MODEL_UPDATE", "DASHBUILDER_IMPORTS_BASE_DIR", "DASHBUILDER_DATASET_PARTITION",
		"DASHBUILDER_COMPONENT_PARTITION"}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	isInSlice := func(a string, list []string) bool {
		for _, b := range list {
			if b == a {
				return true
			}
		}
		return false
	}
	for _, env := range env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env {
		assert.Falsef(t, isInSlice(env.Name, shouldNotContainEnvs), "env %s should not be present", env.Name)
	}
}

func TestRhpamDashbuilderDatasetsAndTemplates(t *testing.T) {

	datasets := []api.KieServerDataSetOrTemplate{
		{
			Name:         "dataset_1",
			Location:     "http://dataset-1.com/rest",
			Token:        "my-dataset-1-token",
			ReplaceQuery: "true",
		},
		{
			Name:     "dataset_2",
			Location: "https://dataset-2.com/rest",
			User:     "user-2",
			Password: "passwd-2",
		},
	}

	templates := []api.KieServerDataSetOrTemplate{
		{
			Name:         "template_1",
			Location:     "http://template-1.com/rest",
			User:         "user-1",
			Password:     "passwd-1",
			ReplaceQuery: "false",
		},
		{
			Name:     "template_2",
			Location: "https://template-2.com/rest",
			Token:    "my-template-2-token",
		},
	}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Config: &api.DashbuilderConfig{
						KieServerDataSets:  datasets,
						KieServerTemplates: templates,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "http://dataset-1.com/rest", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_1_LOCATION"))
	assert.Equal(t, "my-dataset-1-token", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_1_TOKEN"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_1_USER"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_1_PASSWORD"))
	assert.Equal(t, "true", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_1_REPLACE_QUERY"))
	assert.Equal(t, "https://dataset-2.com/rest", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_2_LOCATION"))
	assert.Equal(t, "user-2", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_2_USER"))
	assert.Equal(t, "passwd-2", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_2_PASSWORD"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "dataset_2_TOKEN"))
	assert.Equal(t, "dataset_1,dataset_2", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIESERVER_DATASETS"))
	assert.Equal(t, "http://template-1.com/rest", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_1_LOCATION"))
	assert.Equal(t, "user-1", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_1_USER"))
	assert.Equal(t, "passwd-1", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_1_PASSWORD"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_1_TOKEN"))
	assert.Equal(t, "false", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_1_REPLACE_QUERY"))
	assert.Equal(t, "https://template-2.com/rest", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_2_LOCATION"))
	assert.Equal(t, "my-template-2-token", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_2_TOKEN"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_2_USER"))
	assert.Equal(t, "", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "template_2_PASSWORD"))
	assert.Equal(t, "template_1,template_2", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIESERVER_SERVER_TEMPLATES"))
}

func TestRHPAMDashbuilderIntegrationWithKieServer(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Config: &api.DashbuilderConfig{
						EnableKieServer: true,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "test-dash-kieserver", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIESERVER_SERVER_TEMPLATES"))
	assert.Equal(t, "http://test-dash-kieserver:8080/services/rest/server", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "test_dash_kieserver_LOCATION"))
	assert.Equal(t, "adminUser", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "test_dash_kieserver_USER"))
	assert.Equal(t, "RedHat", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "test_dash_kieserver_PASSWORD"))
}

func TestRHPAMDashbuilderIntegrationWithBC(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Config: &api.DashbuilderConfig{
						EnableBusinessCentral: true,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "http://rhpamdash:8080", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_DASHBUILDER_RUNTIME_LOCATION"))
	assert.Equal(t, "true", getEnvVariable(env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "DASHBUILDER_RUNTIME_MULTIPLE_IMPORT"))
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
	assert.Equal(t, "64Mi", env.Console.PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())
	assert.Equal(t, adminPassword, cr.Status.Applied.CommonConfig.AdminPassword)
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, bcmImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

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
	checkAuthoringHAEnv(t, cr, env, constants.RhdmPrefix)
	assert.Equal(t, "1Gi", env.Console.PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
	assert.Equal(t, constants.ImageRegistry+"/"+constants.RhdmPrefix+"-7/"+constants.RhdmPrefix+"-decisioncentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
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
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					PvSize: "3Gi",
				},
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
	assert.Nil(t, err, "Error getting prod environment")
	checkAuthoringHAEnv(t, cr, env, constants.RhpamPrefix)
	assert.Equal(t, "3Gi", env.Console.PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())
	assert.Equal(t, constants.ImageRegistry+"/"+constants.RhpamPrefix+"-7/"+constants.RhpamPrefix+"-businesscentral-rhel8"+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	amqClusterPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_JMS_BROKER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	amqPassword := getEnvVariable(env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0], "AMQ_PASSWORD")
	assert.Equal(t, "amq", amqPassword, "Expected provided password to take effect, but found %v", amqPassword)
	adminPassword := getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, "admin", adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	amqClusterPassword = getEnvVariable(env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0], "AMQ_CLUSTER_PASSWORD")
	assert.Equal(t, "cluster", amqClusterPassword, "Expected provided password to take effect, but found %v", amqClusterPassword)
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")

}

func checkAuthoringHAEnv(t *testing.T, cr *api.KieApp, env api.Environment, productPrefix string) {
	var partitionValue int32
	partitionValue = 0
	assert.Equal(t, "test-"+productPrefix+"centr", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)
	assert.Equal(t, "test-datagrid", env.Others[0].StatefulSets[0].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[0].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[0].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, "test-amq", env.Others[0].StatefulSets[1].ObjectMeta.Name)
	assert.Equal(t, "RollingUpdate", string(env.Others[0].StatefulSets[1].Spec.UpdateStrategy.Type))
	assert.Equal(t, &partitionValue, env.Others[0].StatefulSets[1].Spec.UpdateStrategy.RollingUpdate.Partition)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/datagrid/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/amq7/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].BrokerImage, constants.VersionConstants[cr.Status.Applied.Version].BrokerImageTag), env.Others[0].StatefulSets[1].Spec.Template.Spec.Containers[0].Image)
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
	assert.Equal(t, fmt.Sprintf(rhdmkieServerImage+":"+cr.Status.Applied.Version), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Decision Central should be omitted")
	assert.Nil(t, env.Console.PersistentVolumeClaims)
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PORT"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func TestRhdmProdImmutableEnvironmentWithReposPersistedWithoutStorageClass(t *testing.T) {

	cr := api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						PersistRepos: true,
					},
				},
			},
		},
	}

	cr.Spec.Environment = api.RhdmProductionImmutable

	env, err := GetEnvironment(&cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	runCommonAssertsForKieServerPersistentStorageVolumeMounts(t, cr, env)

	assert.Equal(t, fmt.Sprintf(rhdmkieServerImage+":"+cr.Status.Applied.Version), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	runCommonAssertsForKieServerPersistentStorageTests(t, cr, env)
}

func TestRhdmProdImmutableEnvironmentWithReposPersistedWithStorageClassAndCustomPVSize(t *testing.T) {

	cr := api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						PersistRepos:     true,
						ServersM2PvSize:  "10Gi",
						ServersKiePvSize: "150Mi",
					},
				},
			},
		},
	}

	env, err := GetEnvironment(&cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	runCommonAssertsForKieServerPersistentStorageVolumeMounts(t, cr, env)

	assert.Equal(t, fmt.Sprintf(rhdmkieServerImage+":"+cr.Status.Applied.Version), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	runCommonAssertsForKieServerPersistentStorageTests(t, cr, env)
	runCommonAssertsForKieServerPersistentStoragePVCTests(t, cr, env, "10Gi", "150Mi")
}

func TestRhpamProdImmutableEnvironmentWithReposPersistedWithStorageClassAndDefaultPVSize(t *testing.T) {

	cr := api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						PersistRepos: true,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(&cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	runCommonAssertsForKieServerPersistentStorageVolumeMounts(t, cr, env)

	assert.Equal(t, fmt.Sprintf(rhpamkieServerImage+":"+cr.Status.Applied.Version), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	runCommonAssertsForKieServerPersistentStorageTests(t, cr, env)
	runCommonAssertsForKieServerPersistentStoragePVCTests(t, cr, env, "1Gi", "10Mi")
}

func runCommonAssertsForKieServerPersistentStoragePVCTests(t *testing.T, cr api.KieApp, env api.Environment, m2Size string, kieSize string) {
	assert.Nil(t, env.Servers[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	cr.Spec.Objects.Servers[0].StorageClassName = "silver"

	env, err := GetEnvironment(&cr, test.MockService())
	assert.Nil(t, err)

	assert.Equal(t, "test-m2-repository-claim", env.Servers[0].PersistentVolumeClaims[0].Name)
	assert.Equal(t, "silver", *env.Servers[0].PersistentVolumeClaims[0].Spec.StorageClassName)
	assert.Equal(t, m2Size, env.Servers[0].PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())

	assert.Equal(t, "test-kie-repository-claim", env.Servers[0].PersistentVolumeClaims[1].Name)
	assert.Equal(t, "silver", *env.Servers[0].PersistentVolumeClaims[1].Spec.StorageClassName)
	assert.Equal(t, kieSize, env.Servers[0].PersistentVolumeClaims[1].Spec.Resources.Requests.Storage().String())
}

func TestRhpamTrialWithReposPersistedWithStorageClass(t *testing.T) {

	cr := api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						PersistRepos:     true,
						ServersM2PvSize:  "2Gi",
						ServersKiePvSize: "150Mi",
					},
				},
			},
		},
	}

	env, err := GetEnvironment(&cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	m2RepoVM, kieRepoVM, m2Vol, kieVol := kieServerPersistentStorageCommonConfig(&cr)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, m2RepoVM)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, kieRepoVM)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, m2Vol)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, kieVol)

	// there shouldn't be any pvc on trial env
	assert.Len(t, env.Servers[0].PersistentVolumeClaims, 0)
	assert.Equal(t, false, cr.Status.Applied.Objects.Servers[0].PersistRepos)
}

func runCommonAssertsForKieServerPersistentStorageVolumeMounts(t *testing.T, cr api.KieApp, env api.Environment) {
	m2RepoVM, kieRepoVM, m2Vol, kieVol := kieServerPersistentStorageCommonConfig(&cr)

	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, m2RepoVM)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, kieRepoVM)

	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, m2Vol)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, kieVol)
}

func runCommonAssertsForKieServerPersistentStorageTests(t *testing.T, cr api.KieApp, env api.Environment) {
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Business Central should be omitted")
	assert.Nil(t, env.Console.PersistentVolumeClaims)
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PORT"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func kieServerPersistentStorageCommonConfig(cr *api.KieApp) (corev1.VolumeMount, corev1.VolumeMount, corev1.Volume, corev1.Volume) {

	m2RepoVM := corev1.VolumeMount{
		Name:      cr.Status.Applied.CommonConfig.ApplicationName + "-m2-repository",
		MountPath: "/home/jboss/.m2/repository",
	}
	kieRepoVM := corev1.VolumeMount{
		Name:      cr.Status.Applied.CommonConfig.ApplicationName + "-kie-repository",
		MountPath: "/home/jboss/.kie/repository",
	}

	m2Vol := corev1.Volume{
		Name: cr.Status.Applied.CommonConfig.ApplicationName + "-m2-repository",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: cr.Status.Applied.CommonConfig.ApplicationName + "-m2-repository-claim",
			},
		},
	}
	kieVol := corev1.Volume{
		Name: cr.Status.Applied.CommonConfig.ApplicationName + "-kie-repository",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: cr.Status.Applied.CommonConfig.ApplicationName + "-kie-repository-claim",
			},
		},
	}
	return m2RepoVM, kieRepoVM, m2Vol, kieVol
}

func TestRhpamProdImmutableEnvironmentDisableKCVerification(t *testing.T) {
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
							KieServerContainerDeployment: "",
							DisableKCVerification:        true,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_VERIFICATION"), "Variable should exist and be true")
	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_PULL_DEPS"), "Variable should exist and be false")
}

func TestRhdmProdImmutableEnvironmentDisableKCVerificationAndPull(t *testing.T) {
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
							KieServerContainerDeployment: "",
							DisableKCVerification:        true,
							DisablePullDeps:              true,
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_VERIFICATION"), "Variable should exist and be true")
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_PULL_DEPS"), "Variable should exist and be true")
}

func TestRhpamProdImmutableEnvironmentEnableKCVerification(t *testing.T) {
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
							KieServerContainerDeployment: "",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_VERIFICATION"), "Variable should exist and be false")
	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_DISABLE_KC_PULL_DEPS"), "Variable should exist and be false")
}

func TestRhpamProdWithSmartRouterWithSSLDisabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			CommonConfig: api.CommonConfig{
				DisableSsl: true,
			},
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.False(t, env.SmartRouter.Omit, "SmarterRouter should not be omitted")
	assert.Equal(t, "64Mi", env.Console.PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())
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

	routeAnnotations := make(map[string]string)
	routeAnnotations["description"] = "Route for Smart Router's http service."

	assert.Equal(t, 1, len(env.SmartRouter.Routes))
	assert.Equal(t, "test-smartrouter", env.SmartRouter.Routes[0].Name)
	assert.Nil(t, env.SmartRouter.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.SmartRouter.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.SmartRouter.Routes[0].Spec.Port.TargetPort)

	smVolumeMountSecret, _ := getVolumeMountSecret(smartrouterKeyStore, "/etc/smartrouter-secret-volume")
	smVolume, _ := getVolumes(smartrouterKeyStore, "test-smartrouter-app-secret")

	assert.NotContains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, smVolumeMountSecret)
	assert.NotContains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes, smVolume)

	// ssl envs
	assert.Empty(t, getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE"))
	assert.Empty(t, getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE_KEYALIAS"))
	assert.Empty(t, getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE_PASSWORD"))
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

	routeAnnotations := make(map[string]string)
	routeAnnotations["description"] = "Route for Smart Router's https service."
	routeAnnotations[routeBalanceAnnotation] = "source"

	assert.Equal(t, 1, len(env.SmartRouter.Routes))
	assert.Equal(t, "test-smartrouter", env.SmartRouter.Routes[0].Name)
	assert.NotNil(t, env.SmartRouter.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.SmartRouter.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.SmartRouter.Routes[0].Spec.Port.TargetPort)

	smVolumeMountSecret, _ := getVolumeMountSecret(smartrouterKeyStore, "/etc/smartrouter-secret-volume")
	smVolume, _ := getVolumes(smartrouterKeyStore, "test-smartrouter-app-secret")

	assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, smVolumeMountSecret)
	assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes, smVolume)

	// ssl envs
	assert.Equal(t, "/etc/smartrouter-secret-volume/keystore.jks", getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE"))
	assert.Equal(t, "jboss", getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE_KEYALIAS"))
	assert.Empty(t, "", getEnvVariable(env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_TLS_KEYSTORE_PASSWORD"))
}

func TestRhdmProdImmutableJMSEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				// set console anyways to make sure rhdmcentrl is not created
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Replicas: Pint32(1),
					},
				},
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
	assert.True(t, env.Console.Omit, "Decision Central should be omitted")
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.Nil(t, env.Console.DeploymentConfigs)
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
	assert.Equal(t, fmt.Sprintf(rhpamkieServerImage+":"+cr.Status.Applied.Version), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Business Central Monitoring should be omitted by default on immutable env.")
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func TestRhpamProdImmutableEnvironmentWithConsole(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},

		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Replicas: Pint32(2),
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Equal(t, "64Mi", env.Console.PersistentVolumeClaims[0].Spec.Resources.Requests.Storage().String())
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "test-rhpamcentrmon", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhpamcentrmon", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.False(t, env.Console.Omit, "Business Central Monitoring should not be omitted on immutable env if Console is set.")
	assert.NotNil(t, cr.Status.Applied.Objects.Console, "Console should not be nil")
	assert.Equal(t, "test-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, int32(2), env.Console.DeploymentConfigs[0].Spec.Replicas)
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
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Business Central Monitoring should be omitted by default on immutable env.")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func TestRhpamProdImmutableJMSEnvironmentWithConsole(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jms",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Replicas: Pint32(1),
					},
				},
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
	assert.False(t, env.Console.Omit, "Business Central Monitoring should not be omitted.")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_ROUTER_PROTOCOL"), "Variable should not exist")
	assert.Equal(t, "test-jms-rhpamcentrmon", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "OpenShiftStartupStrategy", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"), "Variable should exist")
	assert.NotNil(t, cr.Status.Applied.Objects.Console, "Console should not be nil")
	assert.Equal(t, "test-jms-rhpamcentrmon", env.Console.DeploymentConfigs[0].ObjectMeta.Name)
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.True(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
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
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Business Central Monitoring should be omitted by default on immutable env.")
	assert.Equal(t, "test-jms-kieserver", env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-postgresql", env.Databases[0].DeploymentConfigs[0].Name)
	assert.Equal(t, "test-jms-kieserver-amq", env.Servers[0].DeploymentConfigs[1].Name)
	assert.Equal(t, "amq-jolokia-console", env.Servers[0].Routes[1].Name)
	assert.False(t, env.Servers[0].Routes[1].Spec.TLS == nil)
	assert.Equal(t, "amq-tcp-ssl", env.Servers[0].Routes[2].Name)
	assert.False(t, env.Servers[0].Routes[2].Spec.TLS == nil)
	testAMQEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, env.Servers[0].DeploymentConfigs[1].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, cr.Status.Applied.Objects.Servers[0].Jms.AMQEnableSSL)
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
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
	user1 := cr.Status.Applied.Objects.Servers[0].Jms.Username
	password1 := cr.Status.Applied.Objects.Servers[0].Jms.Password
	assert.Empty(t, cr.Spec.Objects.Servers[1].Jms.Username)
	assert.Empty(t, cr.Spec.Objects.Servers[1].Jms.Password)
	assert.NotEmpty(t, cr.Status.Applied.Objects.Servers[1].Jms.Username)
	assert.NotEmpty(t, cr.Status.Applied.Objects.Servers[1].Jms.Password)
	user2 := cr.Status.Applied.Objects.Servers[1].Jms.Username
	password2 := cr.Status.Applied.Objects.Servers[1].Jms.Password

	_, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.True(t, env.Console.Omit, "Business Central Monitoring should be omitted by default on immutable env.")
	assert.Equal(t, user1, cr.Status.Applied.Objects.Servers[0].Jms.Username)
	assert.Equal(t, password1, cr.Status.Applied.Objects.Servers[0].Jms.Password)
	assert.Equal(t, user2, cr.Status.Applied.Objects.Servers[1].Jms.Username)
	assert.Equal(t, password2, cr.Status.Applied.Objects.Servers[1].Jms.Password)

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

	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERV"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should not exist")
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should not exist")
	assert.Nil(t, env.Console.DeploymentConfigs)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
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
		JavaMaxMemRatio:            Pint32(80),
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
			assert.Equal(t, "80", env.Value)

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
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT",
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
								CommonExtDBObjectURL: api.CommonExtDBObjectURL{
									JdbcURL: "jdbc:sqlserver://192.168.1.129:1433;DatabaseName=rhpam",
									CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
										Driver:               "mssql",
										ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLValidConnectionChecker",
										ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLExceptionSorter",
										BackgroundValidation: "true",
										MinPoolSize:          "10",
										MaxPoolSize:          "10",
										Username:             "sqlserverUser",
										Password:             "sqlserverPwd",
									},
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
	assert.Equal(t, kieServerName+latestTag, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
								CommonExtDBObjectURL: api.CommonExtDBObjectURL{
									JdbcURL: "jdbc:sqlserver://192.168.1.129:1433;DatabaseName=rhpam",
									CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
										Driver:               "mssql",
										ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLValidConnectionChecker",
										ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.mssql.MSSQLExceptionSorter",
										BackgroundValidation: "true",
										MinPoolSize:          "10",
										MaxPoolSize:          "10",
										Username:             "sqlserverUser",
										Password:             "sqlserverPwd",
									},
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
	assert.Equal(t, kieServerName+latestTag, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestKieAppContainerDeploymentWithoutS2iAndNotUseImageTags_BuildConfigNotSet(t *testing.T) {
	serverName := "testing-name"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name: serverName,
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	// ConsolidateObjects
	ConsolidateObjects(env, cr)

	// Since there is not Build section with GitSource
	assert.Len(t, env.Servers[0].BuildConfigs, 0)
	assert.Equal(t, "registry.redhat.io/rhpam-7/rhpam-kieserver-rhel8:"+constants.CurrentVersion, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestKieAppContainerDeploymentWithoutS2iAndWithImageTags_BuildConfigNotSet(t *testing.T) {
	serverName := "testing-name"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			UseImageTags: true,
			Environment:  api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name: serverName,
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT",
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	// ConsolidateObjects
	ConsolidateObjects(env, cr)

	// Since there is not Build section with GitSource
	assert.Len(t, env.Servers[0].BuildConfigs, 0)
	assert.Equal(t, "openshift", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "rhpam-kieserver-rhel8:"+constants.CurrentVersion, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "ImageStreamTag", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
}

func TestBuildConfiguration(t *testing.T) {
	serverName := "testing-name"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name: serverName,
						Build: &api.KieAppBuildObject{
							Env: []corev1.EnvVar{
								{
									Name:  "JAVA_OPTS_APPEND",
									Value: "-Dmyprop=test",
								},
								{
									Name:  "OTHER_ENV",
									Value: "other",
								},
							},
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
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "example1",
							},
						},
					},
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.7.0-SNAPSHOT",
							MavenMirrorURL:               "https://maven.mirror.com/",
							ArtifactDir:                  "dir",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "somebranch",
								ContextDir: "example2",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ANOTHER_ENV",
									Value: "AnotherEnv",
								},
							},
						},
					},
					{
						From: &api.ImageObjRef{
							Kind: "ImageStreamTag",
							ObjectReference: api.ObjectReference{
								Name:      "test",
								Namespace: "other-ns",
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, cr.Spec.Objects.Servers[1].Build.Webhooks, 0)
	assert.Len(t, cr.Spec.Objects.Servers[2].Build.Webhooks, 0)
	var secret1, secret2 string
	for _, s := range cr.Status.Applied.Objects.Servers[1].Build.Webhooks {
		if s.Type == api.GitHubWebhook {
			secret1 = s.Secret
		}
	}
	for _, s := range cr.Status.Applied.Objects.Servers[2].Build.Webhooks {
		if s.Type == api.GitHubWebhook {
			secret2 = s.Secret
		}
	}
	checkWebhooks(t, secret1, secret2, cr, env)

	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, cr.Spec.Objects.Servers[1].Build.Webhooks, 0)
	assert.Len(t, cr.Spec.Objects.Servers[2].Build.Webhooks, 0)
	assert.Equal(t, "example1", env.Servers[1].BuildConfigs[0].Spec.Source.ContextDir)
	assert.Equal(t, "example2", env.Servers[2].BuildConfigs[0].Spec.Source.ContextDir)
	checkWebhooks(t, secret1, secret2, cr, env)

	secret1 = "s3cr3t1"
	secret2 = "s3cr3t2"
	cr.Spec.Objects.Servers[1].Build.Webhooks = []api.WebhookSecret{{Type: api.GitHubWebhook, Secret: secret1}}
	cr.Spec.Objects.Servers[2].Build.Webhooks = []api.WebhookSecret{{Type: api.GitHubWebhook, Secret: secret2}}
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	// ConsolidateObjects
	ConsolidateObjects(env, cr)
	assert.Len(t, cr.Spec.Objects.Servers[1].Build.Webhooks, 1)
	assert.Len(t, cr.Spec.Objects.Servers[2].Build.Webhooks, 1)
	assert.Equal(t, secret1, cr.Spec.Objects.Servers[1].Build.Webhooks[0].Secret)
	assert.Equal(t, secret2, cr.Spec.Objects.Servers[2].Build.Webhooks[0].Secret)
	assert.Equal(t, 4, len(env.Servers))
	assert.Equal(t, "example1", env.Servers[1].BuildConfigs[0].Spec.Source.ContextDir)
	assert.Equal(t, "example2", env.Servers[2].BuildConfigs[0].Spec.Source.ContextDir)
	checkWebhooks(t, secret1, secret2, cr, env)

	// Server Test
	crServer := cr.Status.Applied.Objects.Servers[0]
	server := env.Servers[0]

	assert.Equal(t, serverName, crServer.Name)
	assert.Equal(t, crServer.Name+latestTag, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
	assert.Equal(t, "-Dmyprop=test", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[3].Value)
	assert.Equal(t, "other", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[4].Value)

	// Server #1
	crServer = cr.Status.Applied.Objects.Servers[1]
	server = env.Servers[1]
	assert.Equal(t, kieServerName, crServer.Name)
	assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
	assert.Equal(t, "http://git.example.com", server.BuildConfigs[0].Spec.Source.Git.URI)
	assert.Equal(t, "somebranch", server.BuildConfigs[0].Spec.Source.Git.Ref)
	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
	assert.Equal(t, "https://maven.mirror.com/", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[1].Value)
	assert.Equal(t, "dir", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[2].Value)
	// default envs size is 3
	assert.Equal(t, 3, len(server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env))
	for _, s := range server.BuildConfigs[0].Spec.Triggers {
		if s.GitHubWebHook != nil {
			assert.NotEmpty(t, s.GitHubWebHook.Secret)
			assert.Equal(t, secret1, s.GitHubWebHook.Secret)
		}
	}
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, crServer.Name+latestTag, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)

	// Server #2
	crServer = cr.Status.Applied.Objects.Servers[2]
	server = env.Servers[2]
	assert.Equal(t, "test-kieserver2", crServer.Name)
	assert.Equal(t, buildv1.BuildSourceGit, server.BuildConfigs[0].Spec.Source.Type)
	assert.Equal(t, "http://git.example.com", server.BuildConfigs[0].Spec.Source.Git.URI)
	assert.Equal(t, "somebranch", server.BuildConfigs[0].Spec.Source.Git.Ref)
	assert.Equal(t, "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.7.0-SNAPSHOT", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[0].Value)
	assert.Equal(t, "https://maven.mirror.com/", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[1].Value)
	assert.Equal(t, "dir", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[2].Value)
	assert.Equal(t, "AnotherEnv", server.BuildConfigs[0].Spec.Strategy.SourceStrategy.Env[3].Value)
	for _, s := range server.BuildConfigs[0].Spec.Triggers {
		if s.GitHubWebHook != nil {
			assert.NotEmpty(t, s.GitHubWebHook.Secret)
			assert.Equal(t, secret2, s.GitHubWebHook.Secret)
		}
	}
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, crServer.Name+latestTag, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)

	// Server #3
	crServer = cr.Status.Applied.Objects.Servers[3]
	server = env.Servers[3]
	assert.Equal(t, "test-kieserver3", crServer.Name)
	assert.Empty(t, server.ImageStreams)
	assert.Empty(t, server.BuildConfigs)
	assert.Equal(t, "ImageStreamTag", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Kind)
	assert.Equal(t, "test", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "other-ns", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func checkWebhooks(t *testing.T, secret1, secret2 string, cr *api.KieApp, env api.Environment) {
	checkWebhook(t, secret1, cr.Status.Applied.Objects.Servers[1], env.Servers[1])
	checkWebhook(t, secret2, cr.Status.Applied.Objects.Servers[2], env.Servers[2])
}

func checkWebhook(t *testing.T, secret string, crServer api.KieServerSet, server api.CustomObject) {
	assert.Len(t, crServer.Build.Webhooks, 2)
	for _, webhook := range crServer.Build.Webhooks {
		if webhook.Type == api.GitHubWebhook {
			assert.NotEmpty(t, webhook.Secret)
			assert.Equal(t, secret, webhook.Secret, "secret not correct")
		}
		if webhook.Type == api.GenericWebhook {
			assert.NotEmpty(t, webhook.Secret)
		}
	}
	assert.Len(t, server.BuildConfigs[0].Spec.Triggers, 4)
	for _, trigger := range server.BuildConfigs[0].Spec.Triggers {
		if trigger.GitHubWebHook != nil {
			assert.NotEmpty(t, trigger.GitHubWebHook.Secret)
			assert.Equal(t, secret, trigger.GitHubWebHook.Secret, "secret in trigger not correct")
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

func TestRhpamAuthoringEnvironment(t *testing.T) {
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
	assert.Nil(t, err, "Error rhpam getting authoring environment")

	commonRhpamAuthoringAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(bcHttpsRouteDescription)

	assert.Equal(t, 1, len(env.Console.Routes))
	assert.Equal(t, "test-rhpamcentr", env.Console.Routes[0].Name)
	assert.NotNil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Console.Routes[0].Spec.Port.TargetPort)

	assertContainBCAndKSVolumes(t, bcKeyStoreVolumeName, env)

	assert.Equal(t, 1, len(env.Servers[0].Routes))
	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.NotNil(t, env.Servers[0].Routes[0].Spec.TLS)
	routeAnnotations["description"] = ksHttpsRouteDescription
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)

	// bc ssl envs
	assertHTTPSEnvs(t, bcKeystoreVolume, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0])

	// ks ssl envs
	assertHTTPSEnvs(t, "/etc/kieserver-secret-volume", env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0])
}

func TestRhpamAuthoringEnvironmentWithSSLDisabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			CommonConfig: api.CommonConfig{
				DBPassword: "Database",
				DisableSsl: true,
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting rhpam authoring environment with SSL disabled.")

	commonRhpamAuthoringAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(bcHttpRouteDescription)
	delete(routeAnnotations, routeBalanceAnnotation)

	assert.Equal(t, 1, len(env.Console.Routes))
	assert.Equal(t, "test-rhpamcentr", env.Console.Routes[0].Name)
	assert.Nil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Console.Routes[0].Spec.Port.TargetPort)

	assertNotContainBCAndKSVolumes(t, env)

	assert.Equal(t, 1, len(env.Servers[0].Routes))
	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.Nil(t, env.Servers[0].Routes[0].Spec.TLS)
	routeAnnotations["description"] = ksHttpRouteDescription
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)

	// bc ssl envs
	assertHTTPEmpty(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0])

	// ks ssl envs
	assertHTTPEmpty(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0])
}

func commonRhpamAuthoringAssertions(t *testing.T, env api.Environment, cr *api.KieApp) {
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	dbPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHPAM_PASSWORD")
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
	assert.Equal(t, "Database", dbPassword, "Expected provided password to take effect, but found %v", dbPassword)
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Name), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, string(appsv1.DeploymentStrategyTypeRolling), string(env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Strategy.Type), "The DC should use a Rolling strategy when using the H2 DB")
	assert.NotEqual(t, api.Environment{}, env, "Rhpam Authoring Environment should not be empty.")

	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

	// test kieserver probes
	assert.Equal(t, getLivenessReadiness("/services/rest/server/readycheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/services/rest/server/healthcheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)

}

func TestRhdmAuthoringEnvironment(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmAuthoring,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting rhdm authoring environment")

	commonRhdmAuthoringAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(bcHttpsRouteDescription)

	assert.Equal(t, 1, len(env.Console.Routes))
	assert.Equal(t, "test-rhdmcentr", env.Console.Routes[0].Name)
	assert.NotNil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Console.Routes[0].Spec.Port.TargetPort)

	assertContainBCAndKSVolumes(t, dcKeyStoreVolumeName, env)

	assert.Equal(t, 1, len(env.Servers[0].Routes))
	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.NotNil(t, env.Servers[0].Routes[0].Spec.TLS)
	routeAnnotations["description"] = ksHttpsRouteDescription
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)
}

func TestRhdmAuthoringEnvironmentWithSSLDisabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmAuthoring,
			CommonConfig: api.CommonConfig{
				DisableSsl: true,
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting authoring environment")

	commonRhdmAuthoringAssertions(t, env, cr)

	routeAnnotations := getRouteAnnotations(bcHttpRouteDescription)
	delete(routeAnnotations, routeBalanceAnnotation)

	assert.Equal(t, 1, len(env.Console.Routes))
	assert.Equal(t, "test-rhdmcentr", env.Console.Routes[0].Name)
	assert.Nil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Console.Routes[0].Spec.Port.TargetPort)

	assertNotContainBCAndKSVolumes(t, env)

	assert.Equal(t, 1, len(env.Servers[0].Routes))
	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.Nil(t, env.Servers[0].Routes[0].Spec.TLS)
	routeAnnotations["description"] = ksHttpRouteDescription
	assert.Equal(t, routeAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)
}

func assertContainBCAndKSVolumes(t *testing.T, keyStoreVolumeName string, env api.Environment) {
	bcVolumeMountSecret, ksVolumeMountSecret := getVolumeMountSecret(keyStoreVolumeName, bcKeystoreVolume)
	bcVolume, ksVolume := getVolumes(keyStoreVolumeName, bcKeySecret)
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, bcVolumeMountSecret)
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, bcVolume)

	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, ksVolumeMountSecret)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, ksVolume)
}

func assertNotContainBCAndKSVolumes(t *testing.T, env api.Environment) {
	bcVolumeMountSecret, ksVolumeMountSecret := getVolumeMountSecret(dcKeyStoreVolumeName, bcKeystoreVolume)
	bcVolume, ksVolume := getVolumes(dcKeyStoreVolumeName, bcKeySecret)
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, bcVolumeMountSecret)
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, bcVolume)

	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, ksVolumeMountSecret)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, ksVolume)
}

func commonRhdmAuthoringAssertions(t *testing.T, env api.Environment, cr *api.KieApp) {
	assert.True(t, env.SmartRouter.Omit, "SmarterRouter should be omitted")
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
	assert.Equal(t, fmt.Sprintf("%s-kieserver", cr.Name), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Name, "the container name should have incremented")
	assert.Equal(t, string(appsv1.DeploymentStrategyTypeRolling), string(env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Strategy.Type), "The DC should use a Rolling strategy when using the H2 DB")
	assert.NotEqual(t, api.Environment{}, env, "Rhdm Authoring Environment should not be empty.")

	assert.Equal(t, "test-rhdmcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

	// test kieserver probes
	assert.Equal(t, getLivenessReadiness("/services/rest/server/readycheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Equal(t, getLivenessReadiness("/services/rest/server/healthcheck"), env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
}

func getVolumeMountSecret(consoleVolumeMountName string, mountPath string) (bc corev1.VolumeMount, ks corev1.VolumeMount) {
	return corev1.VolumeMount{
			Name:      consoleVolumeMountName,
			ReadOnly:  true,
			MountPath: mountPath,
		},
		corev1.VolumeMount{
			Name:      "kieserver-keystore-volume",
			ReadOnly:  true,
			MountPath: "/etc/kieserver-secret-volume",
		}
}

func getVolumes(consoleVolumeName string, consoleSecretName string) (console corev1.Volume, ks corev1.Volume) {
	return corev1.Volume{
			Name: consoleVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: consoleSecretName,
				},
			},
		},
		corev1.Volume{
			Name: "kieserver-keystore-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "test-kieserver-app-secret",
				},
			},
		}
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
	assert.NotEqual(t, api.Environment{}, env, "Authoring HA Environment should not be empty")
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhpamcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
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
	assert.Nil(t, cr.Status.Applied.Objects.Console.Cors)
	assert.Nil(t, cr.Status.Applied.Objects.Console.DataGridAuth)

	cr.Spec.Objects = api.KieAppObjects{Console: &api.ConsoleObject{}}
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

func TestConstructDashbuilderObject(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Dashbuilder.Replicas)
	assert.Nil(t, cr.Status.Applied.Objects.Dashbuilder.Cors)

	cr.Spec.Objects = api.KieAppObjects{Dashbuilder: &api.DashbuilderObject{}}
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Nil(t, cr.Spec.Objects.Servers)
	assert.Nil(t, cr.Spec.Objects.Dashbuilder.Replicas)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Dashbuilder.Replicas)

	cr.Spec.Objects.Dashbuilder.Replicas = Pint32(1)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, Pint32(1), cr.Status.Applied.Objects.Dashbuilder.Replicas)
	assert.Equal(t, Pint32(1), cr.Spec.Objects.Dashbuilder.Replicas)

	assert.Equal(t, fmt.Sprintf("%s-rhpamdash", name), env.Dashbuilder.DeploymentConfigs[0].Name)
	assert.Equal(t, int32(1), env.Dashbuilder.DeploymentConfigs[0].Spec.Replicas)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhpam-7/rhpam-dashbuilder-rhel8:%s", cr.Status.Applied.Version), env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

	cr.Spec.UseImageTags = true
	env, err = GetEnvironment(cr, test.MockService())
	assert.Equal(t, fmt.Sprintf("rhpam-dashbuilder-rhel8:%s", cr.Status.Applied.Version), env.Dashbuilder.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)

	cr.Spec.Objects.Dashbuilder.Replicas = Pint32(3)
	cr.Spec.Objects.Dashbuilder.Image = "test"
	cr.Spec.UseImageTags = false
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.Equal(t, "test", cr.Status.Applied.Objects.Dashbuilder.Image)
	assert.Equal(t, Pint32(3), cr.Status.Applied.Objects.Dashbuilder.Replicas)
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
		assert.Nil(t, cr.Status.Applied.Objects.Servers[0].Cors)

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
			assert.Equal(t, fmt.Sprintf(rhpamKieserverAndTag, cr.Status.Applied.Version), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
		assert.Equal(t, fmt.Sprintf(rhpamKieserverAndTag, cr.Status.Applied.Version), env.Servers[i].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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

func sampleResources() *corev1.ResourceRequirements {
	cpuL, _ := resource.ParseQuantity("1")
	cpuR, _ := resource.ParseQuantity("750m")
	memL, _ := resource.ParseQuantity("2Gi")
	memR, _ := resource.ParseQuantity("1536Mi")
	return &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuL,
			corev1.ResourceMemory: memL,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuR,
			corev1.ResourceMemory: memR,
		},
	}
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
			Environment:  api.RhpamTrial,
			UseImageTags: true,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
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
			Environment:  api.RhpamTrial,
			UseImageTags: true,
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
		assert.Equal(t, fmt.Sprintf(rhpamKieserverAndTag, cr.Status.Applied.Version), s.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
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
			Environment:  api.RhdmTrial,
			UseImageTags: true,
			CommonConfig: api.CommonConfig{
				ApplicationName: "trial",
			},
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
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
	assert.Nil(t, cr.Spec.Objects.Console)
	assert.Nil(t, cr.Spec.Objects.Servers)
	assert.NotEmpty(t, cr.Status.Applied.CommonConfig.ApplicationName)
	assert.NotNil(t, cr.Status.Applied.Objects.Console.Replicas)
	assert.Len(t, cr.Status.Applied.Objects.Servers, 1)
}

func TestOpenshiftCA(t *testing.T) {
	smartOptsAppend := "testing"
	dashOptsAppend := "-Djavax.net.ssl.trustStoreType=jks"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Jvm: &api.JvmObject{
						JavaOptsAppend: dashOptsAppend,
					},
				},
				SmartRouter: &api.SmartRouterObject{
					Jvm: &api.JvmObject{
						JavaOptsAppend: smartOptsAppend,
					},
				},
				ProcessMigration: &api.ProcessMigrationObject{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)

	assert.Nil(t, cr.Status.Applied.Truststore)
	assert.False(t, IsOcpCA(cr))
	assert.Empty(t, env.Others[0].ConfigMaps)
	trustVolMnt := corev1.VolumeMount{
		Name:      cr.Status.Applied.CommonConfig.ApplicationName + constants.TruststoreSecret,
		MountPath: constants.TruststorePath,
		ReadOnly:  true,
	}
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.NotContains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.NotContains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	trustVol := corev1.Volume{
		Name: cr.Status.Applied.CommonConfig.ApplicationName + constants.TruststoreSecret,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: cr.Status.Applied.CommonConfig.ApplicationName + constants.TruststoreSecret,
			},
		},
	}
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.NotContains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.NotContains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.NotContains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)

	cr.Spec.Truststore = &api.KieAppTruststore{
		OpenshiftCaBundle: true,
	}
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.True(t, cr.Status.Applied.Truststore.OpenshiftCaBundle)
	assert.True(t, IsOcpCA(cr))
	assert.Len(t, env.Others[0].ConfigMaps, 1)
	// Truststore volumes are mounted
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.Contains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)
	assert.NotContains(t, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, trustVolMnt)

	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.Contains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)
	assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes, trustVol)

	assert.NotNil(t, cr.Status.Applied.Objects.Console.Jvm)
	assert.NotNil(t, cr.Status.Applied.Objects.Dashbuilder.Jvm)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter.Jvm)
	assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].Jvm)

	for _, caOption := range caOptsAppend {
		assert.Contains(t, cr.Status.Applied.Objects.Console.Jvm.JavaOptsAppend, caOption)
		assert.Contains(t, cr.Status.Applied.Objects.Dashbuilder.Jvm.JavaOptsAppend, caOption)
		assert.Contains(t, cr.Status.Applied.Objects.SmartRouter.Jvm.JavaOptsAppend, caOption)
		assert.Contains(t, cr.Status.Applied.Objects.Servers[0].Jvm.JavaOptsAppend, caOption)
	}
	assert.NotEqual(t, smartOptsAppend, cr.Status.Applied.Objects.SmartRouter.Jvm.JavaOptsAppend)
	assert.Contains(t, cr.Status.Applied.Objects.SmartRouter.Jvm.JavaOptsAppend, smartOptsAppend)
	assert.NotEqual(t, dashOptsAppend, cr.Status.Applied.Objects.Dashbuilder.Jvm.JavaOptsAppend)
	assert.Contains(t, cr.Status.Applied.Objects.Dashbuilder.Jvm.JavaOptsAppend, dashOptsAppend)
	envVar := corev1.EnvVar{
		Name:  "JAVA_OPTS_APPEND",
		Value: strings.Join(caOptsAppend, " "),
	}
	smartRouterVar := corev1.EnvVar{
		Name:  "JAVA_OPTS_APPEND",
		Value: strings.Join(append([]string{smartOptsAppend}, caOptsAppend...), " "),
	}
	assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envVar)
	assert.Contains(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envVar)
	assert.Contains(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, smartRouterVar)
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envVar)
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

	bcHttpsAnnotations := getRouteAnnotations(bcHttpsRouteDescription)

	assert.Equal(t, "test-rhpamcentr", env.Console.Routes[0].Name)
	assert.NotNil(t, env.Console.Routes[0].Spec.TLS)
	assert.Equal(t, bcHttpsAnnotations, env.Console.Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Console.Routes[0].Spec.Port.TargetPort)
	assert.Equal(t, "test-rhpamcentr-http", env.Console.Routes[1].Name)
	assert.Nil(t, env.Console.Routes[1].Spec.TLS)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Console.Routes[1].Spec.Port.TargetPort)

	delete(bcHttpsAnnotations, routeBalanceAnnotation)
	bcHttpsAnnotations["description"] = "Route for Business Central's http service."
	assert.Equal(t, bcHttpsAnnotations, env.Console.Routes[1].Annotations)

	serverHttpsAnnotations := getRouteAnnotations(ksHttpsRouteDescription)

	assert.Equal(t, kieServerName, env.Servers[0].Routes[0].Name)
	assert.NotNil(t, env.Servers[0].Routes[0].Spec.TLS)
	assert.Equal(t, serverHttpsAnnotations, env.Servers[0].Routes[0].Annotations)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "https"}, env.Servers[0].Routes[0].Spec.Port.TargetPort)
	assert.Equal(t, "test-kieserver-http", env.Servers[0].Routes[1].Name)
	assert.Nil(t, env.Servers[0].Routes[1].Spec.TLS)
	assert.Equal(t, intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "http"}, env.Servers[0].Routes[1].Spec.Port.TargetPort)

	delete(serverHttpsAnnotations, routeBalanceAnnotation)

	serverHttpsAnnotations["description"] = ksHttpRouteDescription
	assert.Equal(t, serverHttpsAnnotations, env.Servers[0].Routes[1].Annotations)

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
	assert.Equal(t, kieServerName, env.Servers[0].DeploymentConfigs[0].Name)
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
	assert.Equal(t, kieServerName, env.Servers[0].DeploymentConfigs[0].Name)
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
	assert.Equal(t, kieServerName, env.Servers[0].DeploymentConfigs[0].Name)
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

func TestSetProductLabels(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: Pint(1)},
					{Deployments: Pint(1)},
				},
				SmartRouter:      &api.SmartRouterObject{},
				ProcessMigration: &api.ProcessMigrationObject{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, constants.CurrentVersion, cr.Status.Applied.Version)
	testObjectLabels(t, cr, env)

	cr.Spec.Version = constants.PriorVersion
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, constants.PriorVersion, cr.Status.Applied.Version)
	testObjectLabels(t, cr, env)
}

func testObjectLabels(t *testing.T, cr *api.KieApp, env api.Environment) {
	assert.NotNil(t, cr.Status.Applied.Objects.Console)
	assert.Len(t, cr.Spec.Objects.Servers, 2)
	assert.NotNil(t, cr.Spec.Objects.SmartRouter)
	assert.NotNil(t, cr.Spec.Objects.ProcessMigration)
	component := "PAM"
	checkObjectLabels(t, cr, env.Console, component, "rhpam-businesscentral-rhel8")
	checkObjectLabels(t, cr, env.Dashbuilder, component, "rhpam-dashbuilder-rhel8")
	for _, server := range env.Servers {
		checkObjectLabelsForServer(t, cr, server, component)
	}
	checkObjectLabels(t, cr, env.SmartRouter, component, "rhpam-smartrouter-rhel8")
	checkObjectLabels(t, cr, env.ProcessMigration, component, "rhpam-process-migration-rhel8")
}

func checkObjectLabels(t *testing.T, cr *api.KieApp, object api.CustomObject, component string, subcomponent string) {
	for _, dc := range object.DeploymentConfigs {
		checkLabels(t, dc.Spec.Template.Labels, component, cr.Status.Applied.Version, subcomponent)
	}
	for _, ss := range object.StatefulSets {
		checkLabels(t, ss.Spec.Template.Labels, component, cr.Status.Applied.Version, subcomponent)
	}
}

func checkObjectLabelsForServer(t *testing.T, cr *api.KieApp, object api.CustomObject, component string) {
	for _, dc := range object.DeploymentConfigs {
		checkLabels(t, dc.Spec.Template.Labels, component, cr.Status.Applied.Version, getFormattedComponentName(cr, "kieserver"))
	}
	for _, ss := range object.StatefulSets {
		checkLabels(t, ss.Spec.Template.Labels, component, cr.Status.Applied.Version, getFormattedComponentName(cr, "kieserver"))
	}
}

func checkLabels(t *testing.T, labels map[string]string, component, version string, subcomponent string) {
	assert.NotNil(t, labels)
	assert.Equal(t, constants.ProductName, labels[constants.LabelRHproductName])
	assert.Equal(t, version, labels[constants.LabelRHproductVersion])
	assert.Equal(t, component, labels[constants.LabelRHcomponentName])
	assert.Equal(t, version, labels[constants.LabelRHcomponentVersion])
	assert.Equal(t, subcomponent, labels[constants.LabelRHsubcomponentName])
	assert.Equal(t, "application", labels[constants.LabelRHsubcomponentType])
	assert.Equal(t, "Red_Hat", labels[constants.LabelRHcompany])
}

func checkClusterLabels(t *testing.T, cr *api.KieApp, object api.CustomObject) {
	if cr.Spec.Version == constants.CurrentVersion {

		for _, dc := range object.DeploymentConfigs {
			assert.NotNil(t, dc.Spec.Template.Labels[constants.ClusterLabel])
			assert.True(t, strings.HasPrefix(dc.Spec.Template.Labels[constants.ClusterLabel], constants.ClusterLabelPrefix))
		}
		for _, ss := range object.StatefulSets {
			assert.NotNil(t, ss.Spec.Template.Labels[constants.ClusterLabel])
			assert.True(t, strings.HasPrefix(ss.Spec.Template.Labels[constants.ClusterLabel], constants.ClusterLabelPrefix))
		}
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
			Environment:  api.RhpamTrial,
			UseImageTags: true,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources(),
					},
				},
				Servers: []api.KieServerSet{
					{
						Deployments: Pint(deployments),
						KieAppObject: api.KieAppObject{
							Env:       sampleEnv,
							Resources: sampleResources(),
						},
					},
				},
				SmartRouter: &api.SmartRouterObject{
					KieAppObject: api.KieAppObject{
						Env:       sampleEnv,
						Resources: sampleResources(),
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

	assert.Nil(t, err, "Error getting partial trial environment")
	adminUser := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_USER")
	adminPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_ADMIN_PWD")
	assert.Equal(t, cr.Spec.CommonConfig.AdminUser, adminUser, "Expected provided user to take effect, but found %v", adminUser)
	assert.Equal(t, cr.Spec.CommonConfig.AdminPassword, adminPassword, "Expected provided password to take effect, but found %v", adminPassword)
	assert.Equal(t, cr.Spec.CommonConfig.AdminPassword, cr.Status.Applied.CommonConfig.AdminPassword)
	mavenPassword := getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "RHDMCENTR_MAVEN_REPO_PASSWORD")
	assert.Equal(t, "MyPassword", mavenPassword, "Expected default password of RedHat, but found %v", mavenPassword)
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "WORKBENCH_SERVICE_NAME"), "Variable should exist")
	assert.Equal(t, "ws", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_PROTOCOL"), "Variable should exist")
	assert.Equal(t, "test-rhdmcentr", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_SERVICE"), "Variable should exist")
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
						From: &api.ImageObjRef{
							Kind: "ImageStreamTag",
							ObjectReference: api.ObjectReference{
								Name: helloRules,
							},
						},
					},
					{
						From: &api.ImageObjRef{
							Kind: "ImageStreamTag",
							ObjectReference: api.ObjectReference{
								Name: byeRules,
							},
						},
					},
					{
						From: &api.ImageObjRef{
							Kind: "DockerImage",
							ObjectReference: api.ObjectReference{
								Name: "quay.io/custom/image:1.0",
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.Equal(t, helloRules, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, byeRules, env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)

	assert.Equal(t, (*appsv1.DeploymentTriggerImageChangeParams)(nil), env.Servers[2].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams)
	assert.Equal(t, "quay.io/custom/image:1.0", env.Servers[2].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

}

func TestSetKieServerFromBuild(t *testing.T) {
	cr := getCRforTestKieServerFromBuild(false)

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.False(t, cr.Spec.UseImageTags)

	assert.Equal(t, helloRules, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, cr.Status.Applied.Objects.Servers[1].Name+latestTag, env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func TestSetKieServerFromBuildAndWithImageTags(t *testing.T) {
	cr := getCRforTestKieServerFromBuild(true)

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
	assert.True(t, cr.Spec.UseImageTags)

	assert.Equal(t, helloRules, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, cr.Status.Applied.Objects.Servers[1].Name+latestTag, env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "", env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
}

func getCRforTestKieServerFromBuild(useImageTags bool) *api.KieApp {

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						From: &api.ImageObjRef{
							Kind: "ImageStreamTag",
							ObjectReference: api.ObjectReference{
								Name: helloRules,
							},
						},
					},
					{
						From: &api.ImageObjRef{
							Kind: "ImageStreamTag",
							ObjectReference: api.ObjectReference{
								Name: byeRules,
							},
						},
						Build: &api.KieAppBuildObject{
							GitSource: api.GitSource{
								URI: "https://test",
							},
						},
					},
				},
			},
		},
	}

	if useImageTags {
		cr.Spec.UseImageTags = true
	}

	return cr
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
							From: &api.ImageObjRef{
								Kind: "ImageStreamTag",
								ObjectReference: api.ObjectReference{
									Name:      "custom-kieserver",
									Namespace: "",
								},
							},
						},
					},
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.6.0-SNAPSHOT",
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
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.7.0-SNAPSHOT",
							GitSource: api.GitSource{
								URI:        "http://git.example.com",
								Reference:  "anotherbranch",
								ContextDir: "test",
							},
							From: &api.ImageObjRef{
								Kind: "DockerImage",
								ObjectReference: api.ObjectReference{
									Name:      "quay.io/test/custom:1.0",
									Namespace: "",
								},
							},
						},
					},
				},
			},
		},
	}
	// set disconnected env var w/ sha
	imageURL := "image-registry.openshift-image-registry.svc:5000/openshift/testing@sha256:e1168e1a1c6e4f248"
	os.Setenv(constants.DmKieImageVar+constants.CurrentVersion, imageURL)
	env, err := GetEnvironment(cr, test.MockService())

	assert.Nil(t, err, "Error getting prod environment")
	assert.Len(t, env.Servers, 3, "Expect two KIE Servers to be created based on provided build configs")
	assert.Equal(t, "somebranch", env.Servers[0].BuildConfigs[0].Spec.Source.Git.Ref)
	assert.Equal(t, "anotherbranch", env.Servers[1].BuildConfigs[0].Spec.Source.Git.Ref)

	assert.Equal(t, "ImageStreamTag", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Kind)
	assert.Equal(t, "custom-kieserver", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "", env.Servers[0].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Namespace)
	assert.Equal(t, cr.Status.Applied.Objects.Servers[0].Name+latestTag, env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)

	assert.Equal(t, "ImageStreamTag", env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Kind)
	assert.Equal(t, fmt.Sprintf("rhdm-kieserver-rhel8:%v", cr.Status.Applied.Version), env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "openshift", env.Servers[1].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Namespace)
	assert.Len(t, env.Servers[1].ImageStreams, 1)
	assert.Equal(t, cr.Status.Applied.Objects.Servers[1].Name+latestTag, env.Servers[1].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)

	assert.Equal(t, "DockerImage", env.Servers[2].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Kind)
	assert.Equal(t, "quay.io/test/custom:1.0", env.Servers[2].BuildConfigs[0].Spec.Strategy.SourceStrategy.From.Name)

	os.Clearenv()
}

func TestExampleServerCommonConfig(t *testing.T) {
	kieApp := LoadKieApp(t, "crs/v2/snippets/", "server_config.yaml")
	kieApp.Spec.Environment = api.RhpamTrial
	env, err := GetEnvironment(&kieApp, test.MockService())
	assert.NoError(t, err, "Error getting environment for %v", kieApp.Spec.Environment)
	assert.Equal(t, 6, len(env.Servers), "Expect six servers")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].DeploymentConfigs[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].Services[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2", env.Servers[len(env.Servers)-2].Routes[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-http", env.Servers[len(env.Servers)-2].Routes[1].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2", env.Servers[len(env.Servers)-1].DeploymentConfigs[0].Name, "Unexpected name for object")
	assert.Equal(t, "server-config-kieserver2-2", env.Servers[len(env.Servers)-1].Services[0].Name, "Unexpected name for object")
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
								CommonExtDBObjectURL: api.CommonExtDBObjectURL{
									JdbcURL: "jdbc:oracle:thin:@myoracle.example.com:1521:rhpam7",
									CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
										Driver:               "oracle",
										ConnectionChecker:    "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker",
										ExceptionSorter:      "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter",
										BackgroundValidation: "false",
										Username:             "oracleUser",
										Password:             "oraclePwd",
									},
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	major, minor, micro := GetMajorMinorMicro(cr.Status.Applied.Version)
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
	assert.Equal(t, bcImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Equal(t, bcImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "test-rhpamcentr", env.Console.DeploymentConfigs[0].Name)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRecreate, env.Console.DeploymentConfigs[0].Spec.Strategy.Type)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Nil(t, env.Console.DeploymentConfigs)

	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
	assert.Equal(t, bcImage+":"+cr.Status.Applied.Version, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
	for i := 0; i < deployments; i++ {
		idx := ""
		if i > 0 {
			idx = fmt.Sprintf("-%d", i+1)
		}
		assert.Equal(t, 1, len(env.Servers[i].Services))
		assert.Equal(t, 1, len(env.Databases[i].Services))
		assert.Equal(t, fmt.Sprintf("test-kieserver%s", idx), env.Servers[i].Services[0].ObjectMeta.Name)
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
				Console: &api.ConsoleObject{},
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
	assert.Nil(t, err)
	assert.Equal(t, constants.ImageRegistry+"/"+constants.RhdmPrefix+"-7/"+constants.RhdmPrefix+"-kieserver"+constants.RhelVersion+":"+cr.Status.Applied.Version, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)

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
	cr.Spec.Objects.Console = &api.ConsoleObject{
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
	var defaultMode int32 = 0770
	var knownHosts int32 = 0660
	var idRsa int32 = 0440
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
			From: &api.ObjRef{
				Kind: "ConfigMap",
				ObjectReference: api.ObjectReference{
					Name: "test-cm",
				},
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
			From: &api.ObjRef{
				Kind: "Secret",
				ObjectReference: api.ObjectReference{
					Name: "test-secret",
				},
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
					SecretName:  "test-secret",
					DefaultMode: &defaultMode,
				},
			},
		},
		expectedPath: "/some/path",
	}, {
		name: "PersistentVolumeClaim GitHooks are configured",
		gitHooks: &api.GitHooksVolume{
			MountPath: "/some/path",
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test-pvc",
				},
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
	}, {
		name: "SSH Secret for GitHooks are configured",
		gitHooks: &api.GitHooksVolume{
			MountPath: "/some/path",
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test-pvc",
				},
			},
			SSHSecret: "test-ssh-secret",
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.GitHooksSSHSecret,
			MountPath: "/home/jboss/.ssh",
		},
		expectedVolume: &corev1.Volume{
			Name: constants.GitHooksSSHSecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "test-ssh-secret",
					Items: []corev1.KeyToPath{
						{
							Key:  "id_rsa",
							Path: "id_rsa",
							Mode: &idRsa,
						},
						{
							Key:  "known_hosts",
							Path: "known_hosts",
							Mode: &knownHosts,
						},
					},
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
				Console: &api.ConsoleObject{},
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
	assert.Greater(t, len(box.List()), 0)
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
								Username:       "testpim-user",
								Password:       "test-pim-pwd",
								ExtraClassPath: "/tmp/test.jar",
								KieAppObject: api.KieAppObject{
									Replicas:     Pint32(5),
									Image:        "test-pim-image",
									ImageContext: "test-context",
									ImageTag:     "test-pim-image-tag",
								},
								Database: api.ProcessMigrationDatabaseObject{
									InternalDatabaseObject: api.InternalDatabaseObject{
										Type:             api.DatabaseMySQL,
										StorageClassName: "gold",
										Size:             "32Gi",
									},
								},
								Jvm: &api.JvmObject{
									JavaOptsAppend:      "-Dmy-property=value",
									JavaMaxMemRatio:     Pint32(20),
									JavaInitialMemRatio: Pint32(25),
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
				Username:       "testpim-user",
				Password:       "2491032541ee362db900f11af2f8fe0a",
				ExtraClassPath: "/tmp/test.jar",
				KieAppObject: api.KieAppObject{
					Replicas:     Pint32(5),
					Image:        "test-pim-image",
					ImageContext: "test-context",
					ImageTag:     "test-pim-image-tag",
				},
				ImageURL: "test-context/test-pim-image:test-pim-image-tag",
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
				Jvm: api.JvmObject{
					JavaOptsAppend:      "-Dmy-property=value",
					JavaMaxMemRatio:     Pint32(20),
					JavaInitialMemRatio: Pint32(25),
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
				// empty credentials provided, in this case the common.AdminUser and password will be used
				// and the password will be hashed using md5.
				Username: "adminUser",
				Password: "a2d11c9699448828d6fc052bddc37fe6",
				KieAppObject: api.KieAppObject{
					Replicas:     Pint32(1),
					Image:        pimImage,
					ImageTag:     constants.CurrentVersion,
					ImageContext: constants.RhpamPrefix + "-7",
				},
				ImageURL: constants.ProcessMigrationDefaultImageURL + ":" + constants.CurrentVersion,
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
				Jvm: api.JvmObject{
					JavaMaxMemRatio:     Pint32(80),
					JavaInitialMemRatio: Pint32(25),
				},
			},
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
								KieAppObject: api.KieAppObject{
									Image:        "test-pim-image",
									ImageContext: "test-context",
									ImageTag:     "test-pim-image-tag",
								},
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
								KieAppObject: api.KieAppObject{
									Image:        "test-pim-image",
									ImageContext: "test-context",
									ImageTag:     "test-pim-image-tag",
								},
								Database: api.ProcessMigrationDatabaseObject{
									InternalDatabaseObject: api.InternalDatabaseObject{
										Type: api.DatabaseExternal,
									},
									ExternalConfig: &api.CommonExtDBObjectRequiredURL{
										JdbcURL: "jdbc:mariadb://hello-mariadb:3306/pimdb",
										CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
											Driver:                     "mariadb",
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
				// empty credentials provided, in this case the common.AdminUser and password will be used
				// and the password will be hashed using md5.
				Username: "testuser",
				Password: "288252a54f57c3d846d613868f8165f3",
				KieAppObject: api.KieAppObject{
					Replicas:     Pint32(1),
					Image:        "test-pim-image",
					ImageTag:     "test-pim-image-tag",
					ImageContext: "test-context",
				},
				ImageURL: "test-context/test-pim-image:test-pim-image-tag",
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
					ExternalConfig: &api.CommonExtDBObjectRequiredURL{
						JdbcURL: "jdbc:mariadb://hello-mariadb:3306/pimdb",
						CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
							Driver:                     "mariadb",
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
				Jvm: api.JvmObject{
					JavaMaxMemRatio:     Pint32(80),
					JavaInitialMemRatio: Pint32(25),
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

func TestProcessMigrationRouteCustomConfig(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			Objects: api.KieAppObjects{
				ProcessMigration: &api.ProcessMigrationObject{
					Username:       "testpim",
					Password:       "testpimpwd",
					Jvm:            createJvmTestObjectWithoutJavaMaxMemRatio(),
					ExtraClassPath: "/tmp/test.jar",
					KieAppObject: api.KieAppObject{
						Replicas: Pint32(3),

						Env: []corev1.EnvVar{
							{
								Name:  "SCRIPT_DEBUG",
								Value: "true",
							},
						},
					},
				},
			},
		},
	}

	routeAnnotations := make(map[string]string)
	routeAnnotations["description"] = "Route for Process Migration https service."

	env, _ := GetEnvironment(cr, test.MockService())
	env = ConsolidateObjects(env, cr)

	assert.Equal(t, 1, len(env.ProcessMigration.Routes))
	assert.Equal(t, "test-process-migration", env.ProcessMigration.Routes[0].ObjectMeta.Name)
	assert.Nil(t, env.ProcessMigration.Routes[0].Spec.TLS)
	assert.Equal(t, routeAnnotations, env.ProcessMigration.Routes[0].Annotations)
	assert.Equal(t, *Pint32(3), env.ProcessMigration.DeploymentConfigs[0].Spec.Replicas)

	assert.Equal(t, "testpim", cr.Status.Applied.Objects.ProcessMigration.Username)
	assert.Equal(t, "c6b08e2600dd7bb5ae5c8755b25ef45d", cr.Status.Applied.Objects.ProcessMigration.Password)
	assert.Equal(t, "c6b08e2600dd7bb5ae5c8755b25ef45d", cr.Spec.Objects.ProcessMigration.Password)

	assert.Equal(t, 3, len(env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts))

	assert.Equal(t, "/opt/rhpam-process-migration/quarkus-app/config/application.yaml", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "application.yaml", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath)

	assert.Equal(t, "/opt/rhpam-process-migration/quarkus-app/config/application-users.properties", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath)
	assert.Equal(t, "application-users.properties", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].SubPath)

	assert.Equal(t, "/opt/rhpam-process-migration/quarkus-app/config/application-roles.properties", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[2].MountPath)
	assert.Equal(t, "application-roles.properties", env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts[2].SubPath)

	assert.Equal(t, "true", getEnvVariable(env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "SCRIPT_DEBUG"))
	assert.Equal(t, "/tmp/test.jar", getEnvVariable(env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "JBOSS_KIE_EXTRA_CLASSPATH"))
	testJvmObjectWithoutJavaMaxMemRatio(t, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)

	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				ProcessMigration: &api.ProcessMigrationObject{},
			},
		},
	}
	env, _ = GetEnvironment(cr, test.MockService())

	routeAnnotations["description"] = "Route for Process Migration http service."
	assert.Equal(t, 1, len(env.ProcessMigration.Routes))
	assert.Equal(t, "test-process-migration", env.ProcessMigration.Routes[0].ObjectMeta.Name)
	assert.Nil(t, env.ProcessMigration.Routes[0].Spec.TLS)
	assert.Equal(t, *Pint32(1), env.ProcessMigration.DeploymentConfigs[0].Spec.Replicas)
	// check default jvm settings
	testDefaultJvm(t, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)

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
							ExternalConfig: &api.CommonExtDBObjectRequiredURL{
								JdbcURL: "jdbc:mariadb://test-process-migration-mysql:3306/pimdb",
								CommonExternalDatabaseObject: api.CommonExternalDatabaseObject{
									Driver:   "mariadb",
									Username: "pim",
									Password: "pim",
								},
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
						ExternalConfig: &api.CommonExtDBObjectRequiredURL{},
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

func TestJvmDefaultConsole(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					Jvm: createJvmTestObjectWithoutJavaMaxMemRatio(),
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testJvmObjectWithoutJavaMaxMemRatio(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestJvmEmptyConsole(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testDefaultJvm(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestJvmDefaultSmartRouter(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{
					Jvm: createJvmTestObjectWithoutJavaMaxMemRatio(),
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testJvmObjectWithoutJavaMaxMemRatio(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestJvmEmptySmartRouter(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testDefaultJvm(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestJvmDefaultServers(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Jvm: createJvmTestObjectWithoutJavaMaxMemRatio(),
					},
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testJvmObjectWithoutJavaMaxMemRatio(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func TestJvmEmptyServer(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testDefaultJvm(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
}

func testDefaultJvm(t *testing.T, envs []corev1.EnvVar) {
	ratioPresent := false
	initialRatio := false
	for _, env := range envs {
		switch e := env.Name; e {
		case "JAVA_MAX_MEM_RATIO":
			ratioPresent = true
			assert.Equal(t, "80", env.Value)
		case "JAVA_INITIAL_MEM_RATIO":
			initialRatio = true
			assert.Equal(t, "25", env.Value)
		}
	}
	assert.True(t, ratioPresent)
	assert.True(t, initialRatio)
}

func createJvmTestObjectWithoutJavaMaxMemRatio() *api.JvmObject {
	jvmObject := api.JvmObject{
		JavaOptsAppend:             "-Dsome.property=foo",
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

func testJvmObjectWithoutJavaMaxMemRatio(t *testing.T, envs []corev1.EnvVar) {

	assert.Equal(t, "-Dsome.property=foo", getSpecEnv(envs, "JAVA_OPTS_APPEND"))
	assert.Equal(t, "4096", getSpecEnv(envs, "JAVA_MAX_INITIAL_MEM"))
	assert.Equal(t, "true", getSpecEnv(envs, "JAVA_DIAGNOSTICS"))
	assert.Equal(t, "true", getSpecEnv(envs, "JAVA_DEBUG"))
	assert.Equal(t, "8787", getSpecEnv(envs, "JAVA_DEBUG_PORT"))
	assert.Equal(t, "20", getSpecEnv(envs, "GC_MIN_HEAP_FREE_RATIO"))
	assert.Equal(t, "40", getSpecEnv(envs, "GC_MAX_HEAP_FREE_RATIO"))
	assert.Equal(t, "4", getSpecEnv(envs, "GC_TIME_RATIO"))
	assert.Equal(t, "90", getSpecEnv(envs, "GC_ADAPTIVE_SIZE_POLICY_WEIGHT"))
	assert.Equal(t, "100", getSpecEnv(envs, "GC_MAX_METASPACE_SIZE"))
	assert.Equal(t, "-XX:+UseG1GC", getSpecEnv(envs, "GC_CONTAINER_OPTIONS"))
}

func TestSimplifiedMonitoringSwitch(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Env: []corev1.EnvVar{
							{
								Name:  "ORG_APPFORMER_SIMPLIFIED_MONITORING_ENABLED",
								Value: "true",
							},
						},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")

	env = ConsolidateObjects(env, cr)
	spec := env.Console.DeploymentConfigs[0].Spec.Template.Spec
	assert.Equal(t, "true", getEnvVariable(spec.Containers[0], "ORG_APPFORMER_SIMPLIFIED_MONITORING_ENABLED"), "Simplified monitoring should be enabled!")

	for _, volumeMounts := range spec.Containers[0].VolumeMounts {
		if volumeMounts.MountPath == "/opt/kie/data" {
			assert.FailNow(t, "Should not have volume mount for '/opt/kie/data'!")
		}
	}

	for _, volume := range spec.Volumes {
		if strings.Contains(volume.Name, "-pvol") {
			assert.FailNow(t, "Should not have volume configuration for PVC!")
		}
	}

	assert.Nil(t, env.Console.PersistentVolumeClaims, "Should not have PVC!")
}

func TestResourcesDefault(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{
					KieAppObject: api.KieAppObject{},
				},

				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{},
				},
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{},
					},
				},
				ProcessMigration: &api.ProcessMigrationObject{
					KieAppObject: api.KieAppObject{},
				},
			},
		},
	}
	GetEnvironment(cr, test.MockService())
	testCPUReqAndLimit(t, cr, constants.ServersCPULimit, constants.ServersCPURequests,
		constants.ConsoleProdCPULimit, constants.ConsoleProdCPURequests,
		constants.SmartRouterLimits["CPU"], constants.SmartRouterRequests["CPU"],
		constants.ProcessMigrationLimits["CPU"], constants.ProcessMigrationRequests["CPU"])
	testMemoryReqAndLimit(t, cr, constants.ServersMemLimit, constants.ServersMemRequests,
		constants.ConsoleProdMemLimit, constants.ConsoleProdMemRequests,
		constants.SmartRouterLimits["MEM"], constants.SmartRouterRequests["MEM"],
		constants.ProcessMigrationLimits["MEM"], constants.ProcessMigrationRequests["MEM"])
}

func TestResourcesOverrideServers(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{
					KieAppObject: api.KieAppObject{
						Resources: sampleLimitAndRequestsResources,
					},
				},

				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Resources: sampleLimitAndRequestsResources,
					},
				},
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							Resources: sampleLimitAndRequestsResources,
						},
					},
				},
				ProcessMigration: &api.ProcessMigrationObject{
					KieAppObject: api.KieAppObject{
						Resources: sampleLimitAndRequestsResources,
					},
				},
			},
		},
	}
	GetEnvironment(cr, test.MockService())
	testCPUReqAndLimit(t, cr, sampleLimitAndRequestsResources.Limits.Cpu().String(), sampleLimitAndRequestsResources.Requests.Cpu().String(),
		sampleLimitAndRequestsResources.Limits.Cpu().String(), sampleLimitAndRequestsResources.Requests.Cpu().String(),
		sampleLimitAndRequestsResources.Limits.Cpu().String(), sampleLimitAndRequestsResources.Requests.Cpu().String(),
		sampleLimitAndRequestsResources.Limits.Cpu().String(), sampleLimitAndRequestsResources.Requests.Cpu().String())
	testMemoryReqAndLimit(t, cr, sampleLimitAndRequestsResources.Limits.Memory().String(), sampleLimitAndRequestsResources.Requests.Memory().String(),
		sampleLimitAndRequestsResources.Limits.Memory().String(), sampleLimitAndRequestsResources.Requests.Memory().String(),
		sampleLimitAndRequestsResources.Limits.Memory().String(), sampleLimitAndRequestsResources.Requests.Memory().String(),
		sampleLimitAndRequestsResources.Limits.Memory().String(), sampleLimitAndRequestsResources.Requests.Memory().String())
}

func testCPUReqAndLimit(t *testing.T, cr *api.KieApp, lCPUServer string, rCPUServer string, lCPUConsole string, rCPUConsole string, lCPUSmartRouter string, rCPUSmartRouter string, lCPUProcessMigration string, rCPUProcessMigration string) {

	assert.NotNil(t, cr.Status.Applied)
	assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.Console.Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter.Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.ProcessMigration.Resources)

	limitCPUServer := cr.Status.Applied.Objects.Servers[0].Resources.Limits[corev1.ResourceCPU]
	assert.True(t, limitCPUServer.String() == lCPUServer)

	requestsCPUServer := cr.Status.Applied.Objects.Servers[0].Resources.Requests[corev1.ResourceCPU]
	assert.True(t, requestsCPUServer.String() == rCPUServer)

	limitCPUConsole := cr.Status.Applied.Objects.Console.KieAppObject.Resources.Limits[corev1.ResourceCPU]
	assert.True(t, limitCPUConsole.String() == lCPUConsole)

	requestsCPUConsole := cr.Status.Applied.Objects.Console.Resources.Requests[corev1.ResourceCPU]
	assert.True(t, requestsCPUConsole.String() == rCPUConsole)

	limitCPUSmartRouter := cr.Status.Applied.Objects.SmartRouter.KieAppObject.Resources.Limits[corev1.ResourceCPU]
	assert.True(t, limitCPUSmartRouter.String() == lCPUSmartRouter)

	requestsCPUSmartRouter := cr.Status.Applied.Objects.SmartRouter.Resources.Requests[corev1.ResourceCPU]
	assert.True(t, requestsCPUSmartRouter.String() == rCPUSmartRouter)

	limitCPUProcessMigration := cr.Status.Applied.Objects.ProcessMigration.KieAppObject.Resources.Limits[corev1.ResourceCPU]
	assert.True(t, limitCPUProcessMigration.String() == lCPUProcessMigration)

	requestsCPUProcessMigration := cr.Status.Applied.Objects.ProcessMigration.Resources.Requests[corev1.ResourceCPU]
	assert.True(t, requestsCPUProcessMigration.String() == rCPUProcessMigration)
}

func testMemoryReqAndLimit(t *testing.T, cr *api.KieApp, lMEMServers string, rMEMServers string, lMEMConsole string, rMEMConsole string, lMEMSmartRouter string, rMEMSmartRouter string, lMEMProcessMigration string, rMEMProcessMigration string) {
	assert.NotNil(t, cr.Status.Applied)
	assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.Console.Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.SmartRouter.Resources)
	assert.NotNil(t, cr.Status.Applied.Objects.ProcessMigration.Resources)

	limitMEMServer := cr.Status.Applied.Objects.Servers[0].Resources.Limits[corev1.ResourceMemory]
	assert.True(t, limitMEMServer.String() == lMEMServers)

	requestsMEMServer := cr.Status.Applied.Objects.Servers[0].Resources.Requests[corev1.ResourceMemory]
	assert.True(t, requestsMEMServer.String() == rMEMServers)

	limitMEMConsole := cr.Status.Applied.Objects.Console.KieAppObject.Resources.Limits[corev1.ResourceMemory]
	assert.True(t, limitMEMConsole.String() == lMEMConsole)

	requestsMEMConsole := cr.Status.Applied.Objects.Console.Resources.Requests[corev1.ResourceMemory]
	assert.True(t, requestsMEMConsole.String() == rMEMConsole)

	limitMEMSmartRouter := cr.Status.Applied.Objects.SmartRouter.KieAppObject.Resources.Limits[corev1.ResourceMemory]
	assert.True(t, limitMEMSmartRouter.String() == lMEMSmartRouter)

	requestsMEMSmartRouter := cr.Status.Applied.Objects.SmartRouter.Resources.Requests[corev1.ResourceMemory]
	assert.True(t, requestsMEMSmartRouter.String() == rMEMSmartRouter)

	limitMEMProcessMigration := cr.Status.Applied.Objects.ProcessMigration.KieAppObject.Resources.Limits[corev1.ResourceMemory]
	assert.True(t, limitMEMProcessMigration.String() == lMEMProcessMigration)

	requestsMEMProcessMigration := cr.Status.Applied.Objects.ProcessMigration.Resources.Requests[corev1.ResourceMemory]
	assert.True(t, requestsMEMProcessMigration.String() == rMEMProcessMigration)
}

var sampleLimitAndRequestsResources = &corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(200, "m"),
		corev1.ResourceMemory: *resource.NewQuantity(256, "Mi"),
	},
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(100, "m"),
		corev1.ResourceMemory: *resource.NewQuantity(102, "Mi"),
	},
}

func TestSmartRouterDefaultConf(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				SmartRouter: &api.SmartRouterObject{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.SmartRouter.ImageContext, "smartrouter")
}

func TestSmartRouterWithImageContext(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				SmartRouter: createSmartRouter(),
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.SmartRouter.ImageContext, "smartrouter")
}

func createSmartRouter() *api.SmartRouterObject {
	smartRouter := api.SmartRouterObject{
		KieAppObject: api.KieAppObject{
			ImageContext: "rhpam-42",
		},
		Protocol:         "",
		UseExternalRoute: false,
	}
	return &smartRouter
}

func TestConsoleDefaultImage(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.Console.ImageContext, "businesscentral")
}

func TestConsoleWithImageContext(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						ImageContext: "rhpam-41",
					},
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.Console.ImageContext, "businesscentral")
}

func TestServersDefaultImage(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.Servers[0].ImageContext, "kieserver")
}

func TestServersWithImageContext(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							ImageContext: "rhpam-42",
						},
					},
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.Servers[0].ImageContext, "kieserver")
}

func TestProcessMigrationDefaultImage(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				ProcessMigration: &api.ProcessMigrationObject{},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.ProcessMigration.ImageContext, "process-migration")
}

func TestProcessMigrationWithImageContext(t *testing.T) {
	name := "test"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
			Objects: api.KieAppObjects{
				ProcessMigration: &api.ProcessMigrationObject{
					KieAppObject: api.KieAppObject{
						ImageContext: "rhpam-43",
					},
				},
			},
		},
	}
	env, _ := GetEnvironment(cr, test.MockService())
	testContext(t, env.ProcessMigration.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image, cr.Status.Applied.Version, cr.Status.Applied.Objects.ProcessMigration.ImageContext, "process-migration")
}

func testContext(t *testing.T, image, version, context, label string) {
	if context != "" {
		assert.Equal(t, context+"/"+constants.RhpamPrefix+"-"+label+constants.RhelVersion+":"+version, image)
	} else {
		assert.Equal(t, constants.ImageRegistry+constants.PamContext+label+constants.RhelVersion+":"+version, image)
	}
}

func TestClusterLabelsDefaultEnvironment(t *testing.T) {
	consoleLabel := "jgrp.k8s.test-clusterlabel.rhpamcentr"
	serverLabel := "jgrp.k8s.test-clusterlabel-kieserver"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-clusterlabel",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoringHA,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting environment")
	consoleClusterLabel := env.Console.DeploymentConfigs[0].Spec.Template.Labels[constants.ClusterLabel]
	assert.Equal(t, consoleClusterLabel, consoleLabel)
	serverClusterLabel := env.Servers[0].DeploymentConfigs[0].Spec.Template.Labels[constants.ClusterLabel]
	assert.Equal(t, serverClusterLabel, serverLabel)

	consoleKubeLabelNSPresent, consoleKubeLabelPresent := checkKubePingEnvs(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], consoleLabel)
	assert.True(t, consoleKubeLabelNSPresent)
	assert.True(t, consoleKubeLabelPresent)

	serverKubeLabelNSPresent, serverKubeLabelPresent := checkKubePingEnvs(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], serverLabel)
	assert.True(t, serverKubeLabelNSPresent)
	assert.True(t, serverKubeLabelPresent)

}

func checkKubePingEnvs(t *testing.T, container corev1.Container, kubeLabel string) (kubeLabelNSPresent bool, kubeLabelPresent bool) {
	envs := container.Env
	kubeLabelNSPresent = false
	kubeLabelPresent = false
	for _, env := range envs {
		if env.Name == constants.KubeNS {
			kubeLabelNSPresent = true
			assert.True(t, env.ValueFrom.FieldRef.FieldPath == "metadata.namespace")
		}

		if env.Name == constants.KubeLabels {
			kubeLabelPresent = true
			assert.True(t, env.Value == "cluster="+kubeLabel)
		}
	}
	return kubeLabelNSPresent, kubeLabelPresent
}

func TestClusterLabelsRHPAMDashbuilderDefaultEnvironment(t *testing.T) {
	dashLabel := "jgrp.k8s.test-labels-dash.rhpamdash"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-labels-dash",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting dashbuilder rhpam default environment environment")
	checkObjectLabels(t, cr, env.Dashbuilder, "PAM", "rhpam-dashbuilder-rhel8")
	checkClusterLabels(t, cr, env.Dashbuilder)
	dashClusterLabel := env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Labels[constants.ClusterLabel]
	assert.Equal(t, dashClusterLabel, dashLabel)

	dashKubeLabelNSPresent, dashKubeLabelPresent := checkKubePingEnvs(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], dashLabel)
	assert.True(t, dashKubeLabelNSPresent)
	assert.True(t, dashKubeLabelPresent)
}

func TestRhdmProdImmutableEnvironmentWithJbpmClusterEnabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						JbpmCluster: true,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.True(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(2), env.Servers[0].DeploymentConfigs[0].Spec.Replicas)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")

	cr.Spec.Objects.Servers[0].Replicas = Pint32(0)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.True(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(0), env.Servers[0].DeploymentConfigs[0].Spec.Replicas, "a replica setting of zero in spec should not be overriden")

	cr.Spec.Objects.Servers[0].Replicas = Pint32(1)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.True(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(2), env.Servers[0].DeploymentConfigs[0].Spec.Replicas, "a user's setting in spec should only be overridden if set to 1")

	cr.Spec.Objects.Servers[0].Replicas = Pint32(3)
	env, err = GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.True(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(3), env.Servers[0].DeploymentConfigs[0].Spec.Replicas, "a user's setting in spec should only be overridden if set to 1")
}

func TestRhdmProdImmutableEnvironmentWithJbpmClusterDisabled(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						JbpmCluster: false,
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.False(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(1), env.Servers[0].DeploymentConfigs[0].Spec.Replicas)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func TestRhdmProdImmutableEnvironmentWithoutJbpmCluster(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.False(t, cr.Status.Applied.Objects.Servers[0].JbpmCluster)
	assert.Equal(t, int32(1), env.Servers[0].DeploymentConfigs[0].Spec.Replicas)
	assert.Nil(t, cr.Status.Applied.Objects.Console, "Console should be nil")
}

func TestRhdmEnvironmentWithKafkaExt(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Kafka: createKafkaExtObject(),
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)
	assert.Len(t, cr.Spec.Objects.Servers[0].Kafka.Topics, 2)
	envs := env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env

	for _, env := range envs {

		switch e := env.Name; e {
		case "KIE_SERVER_KAFKA_EXT_ENABLED":
			assert.Equal(t, env.Value, "true")

		case "KIE_SERVER_KAFKA_EXT_GROUP_ID":
			assert.Equal(t, env.Value, "my-kafka-group")

		case "KIE_SERVER_KAFKA_EXT_ACKS":
			assert.Equal(t, env.Value, "2")

		case "KIE_SERVER_KAFKA_EXT_AUTOCREATE_TOPICS":
			assert.Equal(t, env.Value, "true")

		case "KIE_SERVER_KAFKA_EXT_MAX_BLOCK_MS":
			assert.Equal(t, env.Value, "2100")

		case "KIE_SERVER_KAFKA_EXT_CLIENT_ID":
			assert.Equal(t, env.Value, "C1234567")

		case "KIE_SERVER_KAFKA_EXT_BOOTSTRAP_SERVERS":
			assert.Equal(t, env.Value, "localhost:9092")

		case "KIE_SERVER_KAFKA_EXT_TOPICS":
			assert.Equal(t, env.Value, "events=my-topics,errors=my-errs")
		}
	}
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_ENABLED"))
	assert.Equal(t, "my-kafka-group", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_GROUP_ID"))
	assert.Equal(t, "2", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_ACKS"))
	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_AUTOCREATE_TOPICS"))
	assert.Equal(t, "2100", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_MAX_BLOCK_MS"))
	assert.Equal(t, "C1234567", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_CLIENT_ID"))
	assert.Equal(t, "localhost:9092", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_BOOTSTRAP_SERVERS"))
	assert.Equal(t, "events=my-topics,errors=my-errs", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_TOPICS"))
}

func createKafkaExtObject() *api.KafkaExtObject {
	kafkaExtObject := api.KafkaExtObject{
		MaxBlockMs:       Pint32(2100),
		AutocreateTopics: Pbool(true),
		BootstrapServers: "localhost:9092",
		GroupID:          "my-kafka-group",
		Acks:             Pint(2),
		Topics:           []string{"events=my-topics", "errors=my-errs"},
		ClientID:         "C1234567",
	}
	return &kafkaExtObject
}

func TestRhdmEnvironmentWithKafkaExtDefault(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Name:  "Server-0",
						Kafka: &api.KafkaExtObject{},
					},
					{
						Name: "Server-1",
					},
					{
						Name:  "Server-2",
						Kafka: createKafkaExtObject(),
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)
	envs := env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env

	extEnabled := false
	for _, env := range envs {
		if strings.HasPrefix(env.Value, "KIE_SERVER_KAFKA") {
			extEnabled = true
		}
	}
	assert.False(t, extEnabled)

	assert.Equal(t, cr.Spec.Objects.Servers[0].Name, "Server-0")
	kafkaSpec := cr.Spec.Objects.Servers[0].Kafka
	assert.NotNil(t, kafkaSpec)
	assert.Empty(t, kafkaSpec.ClientID)
	assert.Nil(t, kafkaSpec.AutocreateTopics)
	assert.Nil(t, kafkaSpec.Topics)
	assert.Empty(t, kafkaSpec.BootstrapServers)
	assert.Empty(t, kafkaSpec.GroupID)
	assert.Nil(t, kafkaSpec.Acks)
	assert.Nil(t, kafkaSpec.MaxBlockMs)

	assert.Equal(t, cr.Status.Applied.Objects.Servers[0].Name, "Server-0")
	kafkaStatus := cr.Status.Applied.Objects.Servers[0].Kafka
	assert.NotNil(t, kafkaStatus)
	assert.Empty(t, kafkaStatus.ClientID)
	assert.Equal(t, kafkaStatus.AutocreateTopics, Pbool(true))
	assert.Nil(t, kafkaStatus.Topics)
	assert.Equal(t, kafkaStatus.BootstrapServers, "localhost:9092")
	assert.Equal(t, kafkaStatus.GroupID, "jbpm-consumer")
	assert.Equal(t, kafkaStatus.Acks, Pint(1))
	assert.Equal(t, kafkaStatus.MaxBlockMs, Pint32(2000))

	assert.Equal(t, cr.Spec.Objects.Servers[1].Name, "Server-1")
	kafkaSpecOne := cr.Spec.Objects.Servers[1].Kafka
	assert.Nil(t, kafkaSpecOne)
	assert.Equal(t, cr.Status.Applied.Objects.Servers[1].Name, "Server-1")
	kafkaStatusOne := cr.Status.Applied.Objects.Servers[1].Kafka
	assert.Nil(t, kafkaStatusOne)

	assert.Equal(t, cr.Spec.Objects.Servers[2].Name, "Server-2")
	kafkaSpecTwo := cr.Spec.Objects.Servers[2].Kafka
	assert.NotNil(t, kafkaSpecTwo)
	assert.Equal(t, kafkaSpecTwo.ClientID, "C1234567")
	assert.True(t, *kafkaSpecTwo.AutocreateTopics)
	assert.Len(t, kafkaSpecTwo.Topics, 2)
	assert.Equal(t, kafkaSpecTwo.Topics[0], "events=my-topics")
	assert.Equal(t, kafkaSpecTwo.Topics[1], "errors=my-errs")
	assert.Equal(t, kafkaSpecTwo.BootstrapServers, "localhost:9092")
	assert.Equal(t, kafkaSpecTwo.GroupID, "my-kafka-group")
	assert.Equal(t, kafkaSpecTwo.Acks, Pint(2))
	assert.Equal(t, kafkaSpecTwo.MaxBlockMs, Pint32(2100))

	assert.Equal(t, cr.Status.Applied.Objects.Servers[2].Name, "Server-2")
	kafkaStatusTwo := cr.Status.Applied.Objects.Servers[2].Kafka
	assert.Equal(t, kafkaStatusTwo.ClientID, "C1234567")
	assert.True(t, *kafkaStatusTwo.AutocreateTopics)
	assert.Len(t, kafkaStatusTwo.Topics, 2)
	assert.Equal(t, kafkaStatusTwo.BootstrapServers, "localhost:9092")
	assert.Equal(t, kafkaStatusTwo.GroupID, "my-kafka-group")
	assert.Equal(t, kafkaStatusTwo.Acks, Pint(2))
	assert.Equal(t, kafkaStatusTwo.MaxBlockMs, Pint32(2100))
}

func TestRhdmEnvironmentWithoutKafkaExt(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting environment")
	assert.NotNil(t, env)
	envs := env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env
	extEnabled := false
	for _, env := range envs {
		if strings.HasPrefix(env.Name, "KIE_SERVER_KAFKA") {
			extEnabled, _ = strconv.ParseBool(env.Value)
		}
	}
	assert.False(t, extEnabled)
	assert.True(t, len(cr.Spec.Objects.Servers) == 0)
	assert.Nil(t, cr.Status.Applied.Objects.Servers[0].Kafka)
}

func TestCRServerCPULimitAndRequestUsingMilicores(t *testing.T) {
	cpuL, _ := resource.ParseQuantity("1500m")
	cpuR, _ := resource.ParseQuantity("1000m")
	cpu := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: cpuL,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: cpuR,
		},
	}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kieapp-cpu-test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							Resources: cpu,
							Replicas:  Pint32(1),
						},
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting environment")
	assert.NotNil(t, env)

	env = ConsolidateObjects(env, cr)
	values := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1500m"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1"),
		},
	}
	assert.Equal(t, values.Requests.Cpu(), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Resources.Requests.Cpu())
	assert.Equal(t, values.Limits.Cpu(), env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Resources.Limits.Cpu())
}

func TestRhpamEnvironmentWithKafkaJBPM(t *testing.T) {
	const dateFormat = "dd-MM-yyyy'T'HH:mm:ss.SSSZ"
	const tasksTopics = "my-tasks-topic"
	const casesTopics = "my-cases-topic"
	const processesTopics = "my-processes-topic"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{Name: "testJbpmEmitter"},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Kafka: createKafkaExtObject(), KafkaJbpmEventEmitters: createKafkaJbpmObject(dateFormat, tasksTopics, casesTopics, processesTopics)},
				},
			},
		},
	}

	testEnvironmentWithKafkaJBPM(t, cr, dateFormat, tasksTopics, casesTopics, processesTopics)
}

func testEnvironmentWithKafkaJBPM(t *testing.T, cr *api.KieApp, dateFormat string, tasksTopic string, casesTopic string, processesTopic string) {
	env, err := GetEnvironment(cr, test.MockService())
	assert.NotNil(t, env)
	assert.Nil(t, err, "Error getting environment")

	for _, env := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env {

		checkJbpmKafkaEnvs(t, env)
	}

	assert.Equal(t, "true", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_EXT_ENABLED"))
	assert.Equal(t, "3", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_ACKS"))
	assert.Equal(t, "localhost:9092", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_BOOTSTRAP_SERVERS"))
	assert.Equal(t, "D12345678", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_CLIENT_ID"))
	assert.Equal(t, "2000", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_MAX_BLOCK_MS"))
	assert.Equal(t, dateFormat, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_DATE_FORMAT"))
	assert.Equal(t, tasksTopic, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_TASKS_TOPIC_NAME"))
	assert.Equal(t, casesTopic, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_CASES_TOPIC_NAME"))
	assert.Equal(t, processesTopic, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_PROCESSES_TOPIC_NAME"))
}

func checkJbpmKafkaEnvs(t *testing.T, env corev1.EnvVar) {
	switch e := env.Name; e {

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_ACKS":
		assert.Equal(t, env.Value, "3")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_BOOTSTRAP_SERVERS":
		assert.Equal(t, env.Value, "localhost:9092")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_CLIENT_ID":
		assert.Equal(t, env.Value, "D12345678")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_MAX_BLOCK_MS":
		assert.Equal(t, env.Value, "2000")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_DATE_FORMAT":
		assert.Equal(t, env.Value, "dd-MM-yyyy'T'HH:mm:ss.SSSZ")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_TASKS_TOPIC_NAME":
		assert.Equal(t, env.Value, "my-tasks-topic")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_CASES_TOPIC_NAME":
		assert.Equal(t, env.Value, "my-cases-topic")

	case "KIE_SERVER_KAFKA_JBPM_EVENT_EMITTER_PROCESSES_TOPIC_NAME":
		assert.Equal(t, env.Value, "my-processes-topic")
	}
}

func createKafkaJbpmObject(dateFormat string, tasksTopics string, casesTopics string, processesTopics string) *api.KafkaJBPMEventEmittersObject {
	kafkaJBPMEventEmittersObject := api.KafkaJBPMEventEmittersObject{
		Acks:               Pint(3),
		BootstrapServers:   "localhost:9092",
		ClientID:           "D12345678",
		MaxBlockMs:         Pint32(2000),
		DateFormat:         dateFormat,
		CasesTopicName:     casesTopics,
		ProcessesTopicName: processesTopics,
		TasksTopicName:     tasksTopics,
	}
	return &kafkaJBPMEventEmittersObject
}

func getRouteAnnotations(routeDescription string) map[string]string {
	routeAnnotation := make(map[string]string)
	routeAnnotation["description"] = routeDescription
	routeAnnotation[routeBalanceAnnotation] = "source"
	routeAnnotation["haproxy.router.openshift.io/timeout"] = "60s"
	return routeAnnotation
}

func getPartialCors() *api.CORSFiltersObject {
	return &api.CORSFiltersObject{
		Filters:          "AC_ALLOW_ORIGIN",
		AllowOriginName:  "Access-Control-Allow-Origin-custom-test",
		AllowOriginValue: "custom-test-value",
	}
}

func checkCors(t *testing.T, cors *api.CORSFiltersObject) {
	assert.Equal(t, cors.Filters, constants.ACFilters)
	assert.Equal(t, cors.AllowOriginName, "Access-Control-Allow-Origin")
	assert.Equal(t, cors.AllowOriginValue, "*")
	assert.Equal(t, cors.AllowMethodsName, "Access-Control-Allow-Methods")
	assert.Equal(t, cors.AllowMethodsValue, "GET, POST, OPTIONS, PUT")
	assert.Equal(t, cors.AllowHeadersName, "Access-Control-Allow-Headers")
	assert.Equal(t, cors.AllowHeadersValue, "Accept, Authorization, Content-Type, X-Requested-With")
	assert.Equal(t, cors.AllowCredentialsName, "Access-Control-Allow-Credentials")
	assert.True(t, *cors.AllowCredentialsValue)
	assert.Equal(t, cors.MaxAgeName, "Access-Control-Max-Age")
	assert.Equal(t, cors.MaxAgeValue, Pint32(1))
}

func checkCustomCors(t *testing.T, cors *api.CORSFiltersObject) {
	assert.Equal(t, cors.Filters, "AC_ALLOW_ORIGIN")
	checkCustomAcAllowOrigin(t, cors)
	assert.Empty(t, cors.AllowMethodsName)
	assert.Empty(t, cors.AllowMethodsValue)

}

func checkCustomAcAllowOrigin(t *testing.T, cors *api.CORSFiltersObject) {
	assert.Equal(t, cors.AllowOriginName, "Access-Control-Allow-Origin-custom-test")
	assert.Equal(t, cors.AllowOriginValue, "custom-test-value")
}

func checkConsoleCORSAssertions(t *testing.T, cr *api.KieApp, env api.Environment) {
	corsEnabled := isCORSEnabled(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, corsEnabled)
	assert.NotNil(t, cr.Spec.Objects.Console)
	cors := cr.Status.Applied.Objects.Console.Cors
	assert.NotNil(t, cors)
	checkCors(t, cors)
}

func checkConsoleCustomCORSAssertions(t *testing.T, cr *api.KieApp, env api.Environment) {
	corsEnabled := isCORSEnabled(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, corsEnabled)
	assert.NotNil(t, cr.Spec.Objects.Console)
	cors := cr.Status.Applied.Objects.Console.Cors
	assert.NotNil(t, cors)
	checkCustomCors(t, cors)
}

func checkEnvCORSAssertions(t *testing.T, env []corev1.EnvVar) {
	corsEnabled := isCORSEnabled(env)
	assert.True(t, corsEnabled)
}

func isCORSEnabled(envs []corev1.EnvVar) bool {
	corsEnabled := false
	for _, env := range envs {
		if strings.HasPrefix(env.Name, "AC_ALLOW") || strings.HasPrefix(env.Name, "FILTERS") {
			if env.Name == "FILTERS" && len(env.Value) > 9 { //AC_ALLOW_
				corsEnabled = true
			} else if env.Name != "FILTERS" && env.Value != "<nil>" && len(env.Value) > 1 {
				corsEnabled = true
			}
		}
	}
	return corsEnabled
}

func TestEnvironmentWithCORS(t *testing.T) {
	const name = "test-cors"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					Cors: &api.CORSFiltersObject{Default: true},
				},
				Servers: []api.KieServerSet{
					{
						Cors: &api.CORSFiltersObject{Default: true},
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting test-cors Test environment")
	assert.NotNil(t, env)
	assert.Len(t, cr.Spec.Objects.Servers, 1)
	checkEnvCORSAssertions(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	cors := cr.Status.Applied.Objects.Servers[0].Cors
	assert.NotNil(t, cors)
	checkCors(t, cors)
	checkConsoleCORSAssertions(t, cr, env)
}

func TestEnvironmentWithPartialCORS(t *testing.T) {
	const name = "test-cors"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					Cors: getPartialCors(),
				},
				Servers: []api.KieServerSet{
					{
						Cors: getPartialCors(),
					},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting test-cors Test environment")
	assert.NotNil(t, env)
	checkEnvCORSAssertions(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	cors := cr.Status.Applied.Objects.Servers[0].Cors
	assert.NotNil(t, cors)
	checkCustomCors(t, cors)
	checkConsoleCustomCORSAssertions(t, cr, env)
}

func TestDashbuilderWithCORS(t *testing.T) {
	const name = "test-cors-dashbuilder"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Cors: &api.CORSFiltersObject{Default: true},
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting test-cors-dashbuilder Test environment")
	assert.NotNil(t, env)
	assert.NotNil(t, cr.Spec.Objects.Dashbuilder)
	checkEnvCORSAssertions(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	cors := cr.Status.Applied.Objects.Dashbuilder.Cors
	assert.NotNil(t, cors)
	checkCors(t, cors)
}

func TestDashbuilderWithPartialCORS(t *testing.T) {
	const name = "test-cors-dashbuilder"
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamStandaloneDashbuilder,
			Objects: api.KieAppObjects{
				Dashbuilder: &api.DashbuilderObject{
					Cors: getPartialCors(),
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting test-cors-dashbuilder Test environment")
	assert.NotNil(t, env)
	assert.NotNil(t, cr.Spec.Objects.Dashbuilder)
	checkEnvCORSAssertions(t, env.Dashbuilder.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	cors := cr.Status.Applied.Objects.Dashbuilder.Cors
	assert.NotNil(t, cors)
	checkCustomCors(t, cors)
}

func TestOpenshitStartupStrategyConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			CommonConfig: api.CommonConfig{
				StartupStrategy: createStartupStrategy(api.OpenshiftStartupStrategy, 4000),
			},
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{},
				Servers: []api.KieServerSet{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)

	assert.Equal(t, api.OpenshiftStartupStrategy, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"))
	assert.Equal(t, "4000", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_TEMPLATE_CACHE_TTL"))

	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_ENABLED"))
	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_GLOBAL_DISCOVERY_ENABLED"))
	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_PREFER_KIESERVER_SERVICE"))
	assert.Equal(t, "4000", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_TEMPLATE_CACHE_TTL"))
}

func TestDefaultStartupStrategyConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
				Console: &api.ConsoleObject{},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)

	assert.Equal(t, api.OpenshiftStartupStrategy, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"))
	assert.Equal(t, "5000", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_TEMPLATE_CACHE_TTL"))

	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_ENABLED"))
	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_GLOBAL_DISCOVERY_ENABLED"))
	assert.Equal(t, "true", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_PREFER_KIESERVER_SERVICE"))
	assert.Equal(t, "5000", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_TEMPLATE_CACHE_TTL"))
}

func TestControllerStartupStrategyConfiguration(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamAuthoring,
			CommonConfig: api.CommonConfig{
				StartupStrategy: createStartupStrategy(api.ControllerStartupStrategy, 5000),
			},
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{},
				Servers: []api.KieServerSet{},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)
	assert.Equal(t, api.ControllerStartupStrategy, getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_STARTUP_STRATEGY"))
	assert.Equal(t, "false", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_ENABLED"))
	assert.Equal(t, "", getEnvVariable(env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_TEMPLATE_CACHE_TTL"))
	checkConsoleControllerStrategyAssertions(t, cr)
}

func createStartupStrategy(strategyName string, ttl int) *api.StartupStrategy {
	strategy := api.StartupStrategy{
		StrategyName:               strategyName,
		ControllerTemplateCacheTTL: Pint(ttl),
	}
	return &strategy
}

func checkConsoleControllerStrategyAssertions(t *testing.T, cr *api.KieApp) {
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting prod environment")
	assert.NotNil(t, env)
	assert.Equal(t, "false", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "KIE_SERVER_CONTROLLER_OPENSHIFT_ENABLED"))
}

func TestRhpamTrialInvalidRouteHostname(t *testing.T) {

	cr := getRhpamTrialRouteHostnameWithCR(
		"server-test random,123.st.com",
		"console-test random,123.st.com",
		"dashbuilder-test random,123.st.com",
		"dashbuilder-test random,123.st.com",
		"process-migration-test random,123.st.com")

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestRhpamTrialInvalidRouteHostname environment")

	assertRouteHostnameEmpty(t, env)
	assert.Empty(t, env.ProcessMigration.Routes[0].Spec.Host)
}

func TestRhpamTrialValidRouteHostname(t *testing.T) {

	cr := getRhpamTrialRouteHostnameWithCR(
		"server-my-custom-route.openshift.com",
		"console-my-custom-route.openshift.com",
		"dashbuilder-my-custom-route.openshift.com",
		"smartrouter-my-custom-route.openshift.com",
		"process-migration-my-custom-route.openshift.com")

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestRhpamTrialValidRouteHostname environment")
	assert.Equal(t, "server-my-custom-route.openshift.com", env.Servers[0].Routes[0].Spec.Host)
	assert.Equal(t, "console-my-custom-route.openshift.com", env.Console.Routes[0].Spec.Host)
	assert.Equal(t, "smartrouter-my-custom-route.openshift.com", env.SmartRouter.Routes[0].Spec.Host)
	assert.Equal(t, "dashbuilder-my-custom-route.openshift.com", env.Dashbuilder.Routes[0].Spec.Host)
	assert.Equal(t, "process-migration-my-custom-route.openshift.com", env.ProcessMigration.Routes[0].Spec.Host)
}

func TestRhpamTrialInvalidRouteHostnameUsingEnvs(t *testing.T) {

	cr := getRhpamTrialRouteHostnameWithEnv(
		"server-test random,123.st.com",
		"console-test random,123.st.com",
		"dashbuilder-test random,123.st.com",
		"smartrouter-test random,123.st.com")

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestRhpamTrialInvalidRouteHostnameUsingEnvs environment")
	assertRouteHostnameEmpty(t, env)
}

func TestRhpamTrialValidRouteHostnameUsingEnvs(t *testing.T) {
	cr := getRhpamTrialRouteHostnameWithEnv(
		"server-env-var.test.com",
		"console-env-var.test.com",
		"dashbuilder-env-var.test.com",
		"smartrouter-env-var.test.com")

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestRhpamTrialValidRouteHostnameUsingEnvs environment")
	assert.Equal(t, "server-env-var.test.com", env.Servers[0].Routes[0].Spec.Host)
	assert.Equal(t, "console-env-var.test.com", env.Console.Routes[0].Spec.Host)
	assert.Equal(t, "smartrouter-env-var.test.com", env.SmartRouter.Routes[0].Spec.Host)
	assert.Equal(t, "dashbuilder-env-var.test.com", env.Dashbuilder.Routes[0].Spec.Host)
}

func getRhpamTrialRouteHostnameWithCR(server string, console string, dash string, smartR string, processM string) *api.KieApp {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							RouteHostname: server,
						},
					},
				},
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						RouteHostname: console,
					},
				},
				SmartRouter: createSmartRouter(),
				Dashbuilder: &api.DashbuilderObject{
					KieAppObject: api.KieAppObject{
						RouteHostname: dash,
					},
				},
				ProcessMigration: &api.ProcessMigrationObject{
					RouteHostname: processM,
				},
			},
		},
	}
	cr.Spec.Objects.SmartRouter.RouteHostname = smartR
	return cr
}

func getRhpamTrialRouteHostnameWithEnv(server string, console string, dashB string, smartR string) *api.KieApp {

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							Env: []corev1.EnvVar{
								{
									Name:  constants.ServersRouteEnv,
									Value: server,
								},
							},
						},
					},
				},
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						Env: []corev1.EnvVar{
							{
								Name:  constants.ConsoleRouteEnv,
								Value: console,
							},
						},
					},
				},
				SmartRouter: createSmartRouter(),
				Dashbuilder: &api.DashbuilderObject{
					KieAppObject: api.KieAppObject{
						Env: []corev1.EnvVar{
							{
								Name:  constants.DashbuilderRouteEnv,
								Value: dashB,
							},
						},
					},
				},
			},
		},
	}

	cr.Spec.Objects.SmartRouter.Env = []corev1.EnvVar{{
		Name:  constants.SmartRouterRouteEnv,
		Value: smartR,
	}}

	return cr
}
func assertRouteHostnameEmpty(t *testing.T, env api.Environment) {
	assert.Empty(t, env.Servers[0].Routes[0].Spec.Host)
	assert.Empty(t, env.Console.Routes[0].Spec.Host)
	assert.Empty(t, env.SmartRouter.Routes[0].Spec.Host)
	assert.Empty(t, env.Dashbuilder.Routes[0].Spec.Host)
}

func TestKieExecutorMDB(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{Name: "testKieExecutorMDB"},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{MDBMaxSession: Pint(40)},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestKieExecutorMDB environment")

	assert.NotNil(t, cr.Status.Applied.Objects.Servers[0].MDBMaxSession)
	mdbMaxSessionPassed := false
	for _, env := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env {
		if strings.HasPrefix(env.Name, "JBOSS_MDB") {
			if env.Name != "JBOSS_MDB_MAX_SESSION" && env.Value == "40" {
				mdbMaxSessionPassed = true
			}
		}
	}
	assert.True(t, mdbMaxSessionPassed)
}

func TestKieExecutorMDBEmpty(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{Name: "testKieExecutorMDB"},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting TestKieExecutorMDBEmpty environment")

	assert.Nil(t, cr.Status.Applied.Objects.Servers[0].MDBMaxSession)
	mdbMaxSessionNotPassed := true
	for _, env := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env {
		if strings.HasPrefix(env.Name, "JBOSS_MDB") {
			if env.Name != "JBOSS_MDB_MAX_SESSION" {
				mdbMaxSessionNotPassed = false
			}
		}
	}
	assert.True(t, mdbMaxSessionNotPassed)
}

func TestDataGridRHPAMAuth(t *testing.T) {
	DataGridAuth(t, api.RhpamAuthoringHA)
}

func TestDataGridRHDMAuth(t *testing.T) {
	DataGridAuth(t, api.RhdmAuthoringHA)
}

func DataGridAuth(t *testing.T, environment api.EnvironmentType) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: environment,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					DataGridAuth: &api.DataGridAuth{
						Username: "InfinispanUser",
						Password: "InfinispanPassword",
					},
				},
			},
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting Test RhDM Authoring HA environment")
	assert.Equal(t, "InfinispanUser", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_INFINISPAN_USERNAME"))
	assert.Equal(t, "InfinispanPassword", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_INFINISPAN_PASSWORD"))
	assert.Equal(t, "auth", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_INFINISPAN_SASL_QOP"))
	assert.Equal(t, "infinispan", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_INFINISPAN_SERVER_NAME"))
	assert.Equal(t, "default", getEnvVariable(env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0], "APPFORMER_INFINISPAN_REALM"))
	assert.Equal(t, "InfinispanUser", getEnvVariable(env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0], "USER"))
	assert.Equal(t, "InfinispanPassword", getEnvVariable(env.Others[0].StatefulSets[0].Spec.Template.Spec.Containers[0], "PASS"))
	assert.NotNil(t, cr.Status.Applied.Objects.Console.DataGridAuth)
	assert.Equal(t, cr.Status.Applied.Objects.Console.DataGridAuth.Username, "InfinispanUser")
	assert.Equal(t, cr.Status.Applied.Objects.Console.DataGridAuth.Password, "InfinispanPassword")
}
