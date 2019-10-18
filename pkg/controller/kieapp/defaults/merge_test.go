package defaults

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"

	"github.com/ghodss/yaml"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeServices(t *testing.T) {
	baseline, err := getEnvironment("rhpam-trial", "test")
	assert.Nil(t, err)
	overwrite := baseline.DeepCopy()

	service1 := baseline.Console.Services[0]
	service1.Labels["source"] = "baseline"
	service1.Labels["baseline"] = "true"
	service2 := service1.DeepCopy()
	service2.Name = service1.Name + "-2"
	service4 := service1.DeepCopy()
	service4.Name = service1.Name + "-4"
	baseline.Console.Services = append(baseline.Console.Services, *service2)
	baseline.Console.Services = append(baseline.Console.Services, *service4)

	service1b := overwrite.Console.Services[0]
	service1b.Labels["source"] = "overwrite"
	service1b.Labels["overwrite"] = "true"
	service3 := service1b.DeepCopy()
	service3.Name = service1b.Name + "-3"
	service5 := service1b.DeepCopy()
	service5.Name = service1b.Name + "-4"
	annotations := service5.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
		service5.Annotations = annotations
	}
	service5.Annotations["delete"] = "true"
	overwrite.Console.Services = append(overwrite.Console.Services, *service3)
	overwrite.Console.Services = append(overwrite.Console.Services, *service5)

	mergedEnv, _ := merge(baseline, *overwrite)
	assert.Equal(t, 3, len(mergedEnv.Console.Services), "Expected 3 services")
	finalService1 := mergedEnv.Console.Services[0]
	finalService2 := mergedEnv.Console.Services[1]
	finalService3 := mergedEnv.Console.Services[2]
	assert.Equal(t, "true", finalService1.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalService1.Labels["overwrite"], "Expected the overwrite label to also be set as part of the merge")
	assert.Equal(t, "overwrite", finalService1.Labels["source"], "Expected the source label to have been overwritten by merge")
	assert.Equal(t, "true", finalService2.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "baseline", finalService2.Labels["source"], "Expected the source label to be baseline")
	assert.Equal(t, "true", finalService3.Labels["overwrite"], "Expected the overwrite label to be set")
	assert.Equal(t, "overwrite", finalService3.Labels["source"], "Expected the source label to be overwrite")
	assert.Equal(t, "test-rhpamcentr-2", finalService2.Name, "Second service name should end with -2")
	assert.Equal(t, "test-rhpamcentr-3", finalService3.Name, "Second service name should end with -3")
}

func TestMergeRoutes(t *testing.T) {
	baseline, err := getEnvironment("rhdm-trial", "test")
	assert.Nil(t, err)
	overwrite := baseline.DeepCopy()

	route1 := baseline.Console.Routes[0]
	route1.Labels["source"] = "baseline"
	route1.Labels["baseline"] = "true"
	route2 := route1.DeepCopy()
	route2.Name = route1.Name + "-2"
	route4 := route1.DeepCopy()
	route4.Name = route1.Name + "-4"
	baseline.Console.Routes = append(baseline.Console.Routes, *route2)
	baseline.Console.Routes = append(baseline.Console.Routes, *route4)

	route1b := overwrite.Console.Routes[0]
	route1b.Labels["source"] = "overwrite"
	route1b.Labels["overwrite"] = "true"
	route3 := route1b.DeepCopy()
	route3.Name = route1b.Name + "-3"
	route5 := route1b.DeepCopy()
	route5.Name = route1b.Name + "-4"
	annotations := route5.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
		route5.Annotations = annotations
	}
	route5.Annotations["delete"] = "true"
	overwrite.Console.Routes = append(overwrite.Console.Routes, *route3)
	overwrite.Console.Routes = append(overwrite.Console.Routes, *route5)

	mergedEnv, err := merge(baseline, *overwrite)
	assert.Nil(t, err, "Error while merging environments")
	assert.Equal(t, 4, len(mergedEnv.Console.Routes), "Expected 4 routes.")
	finalRoute1 := mergedEnv.Console.Routes[0]
	finalRoute3 := mergedEnv.Console.Routes[2]
	finalRoute4 := mergedEnv.Console.Routes[3]
	assert.Equal(t, "true", finalRoute1.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalRoute1.Labels["overwrite"], "Expected the overwrite label to also be set as part of the merge")
	assert.Equal(t, "overwrite", finalRoute1.Labels["source"], "Expected the source label to have been overwritten by merge")
	assert.Equal(t, "true", finalRoute3.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "baseline", finalRoute3.Labels["source"], "Expected the source label to be baseline")
	assert.Equal(t, "true", finalRoute4.Labels["overwrite"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalRoute4.Labels["overwrite"], "Expected the overwrite label to be set")
	assert.Equal(t, "overwrite", finalRoute4.Labels["source"], "Expected the source label to be overwrite")
	assert.Equal(t, "test-rhdmcentr-2", finalRoute3.Name, "Second route name should end with -2")
	assert.Equal(t, "test-rhdmcentr-3", finalRoute4.Name, "Second route name should end with -3")
}

func getEnvironment(environment api.EnvironmentType, name string) (api.Environment, error) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: environment,
		},
	}

	env, err := GetEnvironment(cr, test.MockService())
	if err != nil {
		return api.Environment{}, err
	}
	return env, nil
}

func TestMergeServerDeploymentConfigs(t *testing.T) {
	var dbEnv api.Environment
	err := getParsedTemplate("dbs/postgresql.yaml", "prod", &dbEnv)
	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, appsv1.DeploymentStrategyTypeRolling, dbEnv.Servers[0].DeploymentConfigs[1].Spec.Strategy.Type)
	assert.Equal(t, &intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "100%"}, dbEnv.Servers[0].DeploymentConfigs[1].Spec.Strategy.RollingParams.MaxSurge)

	var prodEnv api.Environment
	err = getParsedTemplate("envs/rhpam-production.yaml", "prod", &prodEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplate("common.yaml", "prod", &common)
	assert.Nil(t, err, "Error: %v", err)

	baseEnvCount := len(common.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	prodEnvCount := len(prodEnv.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)

	mergedDCs := mergeDeploymentConfigs(common.Servers[0].DeploymentConfigs, prodEnv.Servers[0].DeploymentConfigs)
	mergedDCs = mergeDeploymentConfigs(mergedDCs, dbEnv.Servers[0].DeploymentConfigs)

	assert.NotNil(t, mergedDCs, "Must have encountered an error, merged DCs should not be null")
	assert.Len(t, mergedDCs, 2, "Expect 2 deployment descriptors but got %v", len(mergedDCs))

	mergedEnvCount := len(mergedDCs[0].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, mergedEnvCount > baseEnvCount, "Merged DC should have a higher number of environment variables than the base server")
	assert.True(t, mergedEnvCount > prodEnvCount, "Merged DC should have a higher number of environment variables than the server")

	assert.Len(t, mergedDCs[0].Spec.Template.Spec.Containers[0].Ports, 4, "Expecting 4 ports")

}

func TestMergeServerDeploymentConfigsWithJms(t *testing.T) {
	var dbEnv api.Environment
	err := getParsedTemplate("dbs/h2.yaml", "immutable-prod", &dbEnv)
	assert.Nil(t, err, "Error: %v", err)

	var jmsEnv api.Environment
	err = getParsedTemplate("jms/activemq-jms-config.yaml", "immutable-prod", &jmsEnv)
	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, jmsEnv.Servers[0].DeploymentConfigs[1].Name, "immutable-prod-kieserver-amq")
	assert.Equal(t, appsv1.DeploymentStrategyTypeRolling, jmsEnv.Servers[0].DeploymentConfigs[1].Spec.Strategy.Type)
	assert.Equal(t, &intstr.IntOrString{Type: 1, IntVal: 0, StrVal: "100%"}, jmsEnv.Servers[0].DeploymentConfigs[1].Spec.Strategy.RollingParams.MaxSurge)

	var prodEnv api.Environment
	err = getParsedTemplate("envs/rhpam-production-immutable.yaml", "immutable-prod", &prodEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplate("common.yaml", "immutable-prod", &common)
	assert.Nil(t, err, "Error: %v", err)

	baseEnvCount := len(common.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	prodEnvCount := len(prodEnv.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)

	mergedDCs := mergeDeploymentConfigs(common.Servers[0].DeploymentConfigs, prodEnv.Servers[0].DeploymentConfigs)
	mergedDCs = mergeDeploymentConfigs(mergedDCs, dbEnv.Servers[0].DeploymentConfigs)
	mergedDCs = mergeDeploymentConfigs(mergedDCs, jmsEnv.Servers[0].DeploymentConfigs)

	assert.NotNil(t, mergedDCs, "Must have encountered an error, merged DCs should not be null")
	assert.Len(t, mergedDCs, 2, "Expect 2 deployment descriptors but got %v", len(mergedDCs))

	mergedEnvCount := len(mergedDCs[0].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, mergedEnvCount > baseEnvCount, "Merged DC should have a higher number of environment variables than the base server")
	assert.True(t, mergedEnvCount > prodEnvCount, "Merged DC should have a higher number of environment variables than the server")

	assert.Len(t, mergedDCs[0].Spec.Template.Spec.Containers[0].Ports, 4, "Expecting 4 ports")
}

func TestMergeConfigsWithoutOverrides(t *testing.T) {
	var authEnv api.Environment
	err := getParsedTemplate("envs/rhdm-authoring.yaml", "authoring", &authEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplate("common.yaml", "authoring", &common)
	assert.Nil(t, err, "Error: %v", err)

	assert.Equal(t, 1, len(common.Servers))
	assert.Equal(t, 0, len(authEnv.Servers))

	merged, err := merge(common, authEnv)
	assert.Nil(t, err, "Error: %v", err)

	assert.Equal(t, merged.Servers, common.Servers)
}

func TestMergeConfigsWithoutBaseline(t *testing.T) {
	var authEnv api.Environment
	err := getParsedTemplate("envs/rhdm-authoring.yaml", "authoring", &authEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplate("common.yaml", "authoring", &common)
	assert.Nil(t, err, "Error: %v", err)

	assert.Equal(t, 1, len(common.Servers))
	assert.Equal(t, 0, len(authEnv.Servers))

	//Use authEnv as baseline and common as overrides
	merged, err := merge(authEnv, common)
	assert.Nil(t, err, "Error: %v", err)

	assert.Equal(t, merged.Servers, common.Servers)
}

func TestMergeConsoleOmitted(t *testing.T) {
	var trialEnv api.Environment

	err := getParsedTemplate("envs/rhpam-trial.yaml", "test", &trialEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplate("common.yaml", "test", &common)
	assert.Nil(t, err, "Error: %v", err)

	mergedEnv, err := merge(common, trialEnv)
	assert.Nil(t, err, "Error: %v", err)
	assert.False(t, mergedEnv.Console.Omit, "Console deployment must not be omitted")
}

func TestMergeBuildConfigandIStreams(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProductionImmutable,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{
						Build: &api.KieAppBuildObject{
							KieServerContainerDeployment: "test",
						},
					},
				},
			},
		},
	}
	var prodImmutableEnv api.Environment
	err := getParsedTemplateFromCR(cr, "envs/rhpam-production-immutable.yaml", &prodImmutableEnv)
	assert.Nil(t, err, "Error: %v", err)

	var common api.Environment
	err = getParsedTemplateFromCR(cr, "common.yaml", &common)
	assert.Nil(t, err, "Error: %v", err)

	mergedEnv, err := merge(common, prodImmutableEnv)
	assert.Nil(t, err, "Error: %v", err)
	server := mergedEnv.Servers[0]
	assert.Len(t, server.ImageStreams, 1)
	assert.Equal(t, "test-kieserver", server.ImageStreams[0].ObjectMeta.Name)
	assert.Equal(t, "test-kieserver", server.BuildConfigs[0].ObjectMeta.Name)
	assert.Empty(t, server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "test-kieserver:latest", server.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, "test-kieserver", server.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Image)
}

func TestMergeDeploymentconfigs(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}

	overwrite := []appsv1.DeploymentConfig{
		*buildDC("overwrite-dc2"),
		*buildDC("dc1"),
	}
	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, overwrite[0], results[1])
	assert.Equal(t, overwrite[1], results[0])
}

func TestMergeDeploymentconfigs_Metadata(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		{
			ObjectMeta: *buildObjectMeta("dc1-dc"),
		},
	}
	baseline[0].ObjectMeta.Labels["foo"] = "replace me"
	baseline[0].ObjectMeta.Labels["john"] = "doe"
	overwrite[0].ObjectMeta.Labels["foo"] = "replaced"
	overwrite[0].ObjectMeta.Labels["ping"] = "pong"

	baseline[0].ObjectMeta.Annotations["foo"] = "replace me"
	baseline[0].ObjectMeta.Annotations["john"] = "doe"
	overwrite[0].ObjectMeta.Annotations["foo"] = "replaced"
	overwrite[0].ObjectMeta.Annotations["ping"] = "pong"

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, "replaced", results[0].ObjectMeta.Labels["foo"])
	assert.Equal(t, "doe", results[0].ObjectMeta.Labels["john"])
	assert.Equal(t, "pong", results[0].ObjectMeta.Labels["ping"])

	assert.Equal(t, "replaced", results[0].ObjectMeta.Annotations["foo"])
	assert.Equal(t, "doe", results[0].ObjectMeta.Annotations["john"])
	assert.Equal(t, "pong", results[0].ObjectMeta.Annotations["ping"])
}

func TestMergeDeploymentconfigs_TemplateMetadata(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	baseline[0].Spec.Template.ObjectMeta.Labels["foo"] = "replace me"
	baseline[0].Spec.Template.ObjectMeta.Labels["john"] = "doe"
	overwrite[0].Spec.Template.ObjectMeta.Labels["foo"] = "replaced"
	overwrite[0].Spec.Template.ObjectMeta.Labels["ping"] = "pong"

	baseline[0].Spec.Template.ObjectMeta.Annotations["foo"] = "replace me"
	baseline[0].Spec.Template.ObjectMeta.Annotations["john"] = "doe"
	overwrite[0].Spec.Template.ObjectMeta.Annotations["foo"] = "replaced"
	overwrite[0].Spec.Template.ObjectMeta.Annotations["ping"] = "pong"

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, "replaced", results[0].Spec.Template.ObjectMeta.Labels["foo"])
	assert.Equal(t, "doe", results[0].Spec.Template.ObjectMeta.Labels["john"])
	assert.Equal(t, "pong", results[0].Spec.Template.ObjectMeta.Labels["ping"])

	assert.Equal(t, "replaced", results[0].Spec.Template.ObjectMeta.Annotations["foo"])
	assert.Equal(t, "doe", results[0].Spec.Template.ObjectMeta.Annotations["john"])
	assert.Equal(t, "pong", results[0].Spec.Template.ObjectMeta.Annotations["ping"])
}

func TestMergeDeploymentconfigs_Spec_Triggers(t *testing.T) {
	emptyImageChangeParams := &appsv1.DeploymentTriggerImageChangeParams{}
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	// Let's assume it has a build strategy=sourceStrategy
	baseline[0].Spec.Triggers = append(baseline[0].Spec.Triggers, appsv1.DeploymentTriggerPolicy{
		Type:              "ImageChange",
		ImageChangeParams: emptyImageChangeParams,
	})
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}

	overwrite[0].Spec.Triggers[0] = appsv1.DeploymentTriggerPolicy{
		Type: "ImageChange",
		ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
			From: corev1.ObjectReference{
				Name: "other-image:future",
			},
		},
	}
	overwrite[0].Spec.Triggers = append(overwrite[0].Spec.Triggers, appsv1.DeploymentTriggerPolicy{
		Type: "ConfigChange",
	})

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, 3, len(results[0].Spec.Triggers))
	assert.Equal(t, appsv1.DeploymentTriggerType("ImageChange"), results[0].Spec.Triggers[0].Type)
	assert.Empty(t, results[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "other-image:future", results[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, appsv1.DeploymentTriggerType("ImageChange"), results[0].Spec.Triggers[1].Type)
	assert.Equal(t, emptyImageChangeParams, results[0].Spec.Triggers[1].ImageChangeParams)
	assert.Equal(t, appsv1.DeploymentTriggerType("ConfigChange"), results[0].Spec.Triggers[2].Type)
}

func TestMergeDeploymentconfigs_Spec_Other(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	baseline[0].Spec.Selector["foo"] = "replace me"
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite[0].Spec.Strategy.Type = "Other Strategy"
	overwrite[0].Spec.Selector["foo"] = "replaced"
	overwrite[0].Spec.Selector["other"] = "bar"
	overwrite[0].Spec.Paused = true
	overwrite[0].Spec.Test = true
	overwrite[0].Spec.Replicas = 2

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, appsv1.DeploymentStrategyType("Other Strategy"), results[0].Spec.Strategy.Type)
	assert.Equal(t, 3, len(results[0].Spec.Selector))
	assert.Equal(t, "dc1", results[0].Spec.Selector["deploymentConfig"])
	assert.Equal(t, "replaced", results[0].Spec.Selector["foo"])
	assert.Equal(t, "bar", results[0].Spec.Selector["other"])
	assert.True(t, results[0].Spec.Paused)
	assert.True(t, results[0].Spec.Test)
	assert.Equal(t, int32(2), results[0].Spec.Replicas)
}
func TestMergeDeploymentconfigs_PodSpec_Volumes(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite[0].Spec.Template.Spec.Volumes[0] = corev1.Volume{
		Name: "dc1-some-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	overwrite[0].Spec.Template.Spec.Volumes[1] = corev1.Volume{
		Name: "dc1-other-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	overwrite[0].Spec.Template.Spec.Volumes = append(overwrite[0].Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "dc1-secret-volume",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "other-secret",
			},
		},
	})

	overwrite[0].Spec.Template.Spec.Containers[0].VolumeMounts[0] = corev1.VolumeMount{
		Name:      "dc1-volume-mount-1",
		MountPath: "/etc/kieserver/dc1/pathX",
		ReadOnly:  true,
	}
	overwrite[0].Spec.Template.Spec.Containers[0].VolumeMounts = append(overwrite[0].Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      "dc1-volume-mount-2",
		MountPath: "/etc/kieserver/dc1/path2",
		ReadOnly:  true,
	})

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, 3, len(results[0].Spec.Template.Spec.Volumes))
	assert.Equal(t, "dc1-some-volume", results[0].Spec.Template.Spec.Volumes[0].Name)
	assert.Nil(t, results[0].Spec.Template.Spec.Volumes[0].PersistentVolumeClaim)
	assert.Equal(t, &corev1.EmptyDirVolumeSource{}, results[0].Spec.Template.Spec.Volumes[0].EmptyDir)
	assert.Equal(t, "other-secret", results[0].Spec.Template.Spec.Volumes[1].VolumeSource.Secret.SecretName)
	assert.Equal(t, "dc1-other-volume", results[0].Spec.Template.Spec.Volumes[2].Name)
	assert.Equal(t, 2, len(results[0].Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, "/etc/kieserver/dc1/pathX", results[0].Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "/etc/kieserver/dc1/path2", results[0].Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath)
}

func getParsedTemplate(filename string, name string, object interface{}) error {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
	}
	return getParsedTemplateFromCR(cr, filename, object)
}

func getParsedTemplateFromCR(cr *api.KieApp, filename string, object interface{}) error {
	envTemplate, err := getEnvTemplate(cr)
	if err != nil {
		log.Error("Error getting environment template", err)
	}

	yamlBytes, err := loadYaml(test.MockService(), filename, cr.Spec.Version, cr.Namespace, envTemplate)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlBytes, object)
	if err != nil {
		log.Error("Error unmarshalling yaml. ", err)
	}
	return nil
}

func buildDC(name string) *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: *buildObjectMeta(name + "-dc"),
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							name + "-container-1",
						},
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Namespace: "openshift",
							Name:      "rhpam70-kieserver:latest",
						},
					},
				},
			},
			Replicas: 3,
			Selector: map[string]string{
				"deploymentConfig": name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: *buildObjectMeta(name + "-tplt"),
				Spec: corev1.PodSpec{
					ServiceAccountName: name + "test-sa",
					Containers: []corev1.Container{
						{
							Name:  name + "container",
							Image: "image-" + name,
							Env: []corev1.EnvVar{
								{
									Name:  name + "-env-1",
									Value: name + "-val-1",
								},
								{
									Name:  name + "-env-2",
									Value: name + "-val-2",
								},
								{
									Name: name + "-env-3",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: name + "-configmap",
											},
											Key: name + "-configmap-key",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"memory": *resource.NewQuantity(1, "Mi"),
									"cpu":    *resource.NewQuantity(1, ""),
								},
								Requests: corev1.ResourceList{
									"memory": *resource.NewQuantity(2, "Mi"),
									"cpu":    *resource.NewQuantity(2, ""),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      name + "-volume-mount-1",
									MountPath: "/etc/kieserver/" + name + "/path1",
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          name + "-port1",
									Protocol:      "TCP",
									ContainerPort: 9090,
								},
								{
									Name:          name + "-port2",
									Protocol:      "TCP",
									ContainerPort: 8443,
								},
							},
							LivenessProbe:  buildProbe(name+"-liveness", 30, 2),
							ReadinessProbe: buildProbe(name+"-readiness", 60, 4),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: name + "-some-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-claim",
								},
							},
						},
						{
							Name: name + "-secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: name + "-secret",
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildObjectMeta(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: name + "-ns",
		Labels: map[string]string{
			name + ".label1": name + "-labelValue1",
			name + ".label2": name + "-labelValue2",
		},
		Annotations: map[string]string{
			name + ".annotation1": name + "-annValue1",
			name + ".annotation2": name + "-annValue2",
		},
	}
}

func buildProbe(name string, delay, timeout int32) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/" + name,
					"-c",
					name,
				},
			},
		},
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
	}
}
