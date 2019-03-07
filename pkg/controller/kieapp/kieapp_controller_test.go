package kieapp

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	env = consolidateObjects(env, cr)
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

	env, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env = consolidateObjects(env, cr)

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
					{Name: duplicate, Deployments: defaults.Pint(deployments)},
					{Name: duplicate},
				},
			},
		},
	}
	_, err := defaults.GetEnvironment(cr, test.MockService())
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("duplicate kieserver name %s", duplicate))
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
						Deployments: defaults.Pint(deployments),
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
	env, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env = append(env.Servers[deployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, commonAddition)
	env = consolidateObjects(env, cr)

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
						Deployments: defaults.Pint(1),
					},
					{
						Name:        "server-b",
						Deployments: defaults.Pint(deployments),
					},
				},
			},
		},
	}

	env, err := defaults.GetEnvironment(cr, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	env = consolidateObjects(env, cr)

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
	_, err := defaults.GetEnvironment(cr, test.MockService())
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
	_, err = defaults.GetEnvironment(cr2, test.MockService())
	if !assert.Nil(t, err, "error should be nil") {
		log.Error("Error getting environment. ", err)
	}
	assert.Equal(t, registry2, cr2.Spec.ImageRegistry.Registry)
	assert.Equal(t, false, cr2.Spec.ImageRegistry.Insecure)
}

func TestGenerateSecret(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Servers: []v1.KieServerSet{
					{Deployments: defaults.Pint(3)},
					{Name: "testing", Deployments: defaults.Pint(4)},
					{Deployments: defaults.Pint(2)},
				},
			},
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
	reconciler := &Reconciler{mockService}
	env, _, err = reconciler.newEnv(cr)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Len(t, env.Console.Secrets, 1, "One secret should be generated for the trial workbench")
	for _, server := range env.Servers {
		assert.Len(t, server.Secrets, 1, "One secret should be generated for each trial kieserver")
		secretName := fmt.Sprintf(constants.KeystoreSecret, server.DeploymentConfigs[0].Name)
		assert.Equal(t, secretName, server.Secrets[0].Name)
		for _, volume := range server.DeploymentConfigs[0].Spec.Template.Spec.Volumes {
			if volume.Secret != nil {
				assert.Equal(t, secretName, volume.Secret.SecretName)
			}
		}
	}
}

func TestSpecifySecret(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Console: v1.SecuredKieAppObject{
					KieAppObject: v1.KieAppObject{
						KeystoreSecret: "console-ks-secret",
					},
				},
				Servers: []v1.KieServerSet{
					{
						SecuredKieAppObject: v1.SecuredKieAppObject{
							KieAppObject: v1.KieAppObject{
								KeystoreSecret: "server-ks-secret",
							},
						},
					},
				},
				SmartRouter: v1.KieAppObject{
					KeystoreSecret: "smartrouter-ks-secret",
				},
			},
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
	reconciler := &Reconciler{mockService}
	env, _, err = reconciler.newEnv(cr)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Len(t, env.Console.Secrets, 0, "Zero secrets should be generated for the trial workbench")
	assert.Len(t, env.Servers[0].Secrets, 0, "Zero secrets should be generated for the trial kieserver")
	assert.Len(t, env.SmartRouter.Secrets, 0, "Zero secrets should be generated for the smartrouter")
	for _, volume := range env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Spec.Objects.Console.KeystoreSecret, volume.Secret.SecretName)
		}
	}
	for _, volume := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Spec.Objects.Servers[0].KeystoreSecret, volume.Secret.SecretName)
		}
	}
	for _, volume := range env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Spec.Objects.SmartRouter.KeystoreSecret, volume.Secret.SecretName)
		}
	}
}

func TestConsoleHost(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhdm-trial",
		},
	}

	scheme, err := v1.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := &Reconciler{mockService}
	_, _, err = reconciler.newEnv(cr)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Equal(t, fmt.Sprintf("http://%s", cr.Spec.CommonConfig.ApplicationName), cr.Status.ConsoleHost, "spec.commonConfig.consoleHost should be URL from the resulting workbench route host")
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
	env, err := defaults.GetEnvironment(cr, test.MockService())
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
	assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
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

func TestCreateRhpamImageStreams(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	reconciler.createLocalImageTag(fmt.Sprintf("rhpam%s-businesscentral-openshift:1.0", cr.Spec.CommonConfig.Version), cr)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhpam%s-businesscentral-openshift:1.0", cr.Spec.CommonConfig.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhpam-7/rhpam%s-businesscentral-openshift:1.0", cr.Spec.CommonConfig.Version), isTag.Tag.From.Name)
}

func TestCreateRhdmImageStreams(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhdm-trial",
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	reconciler.createLocalImageTag(fmt.Sprintf("rhdm%s-decisioncentral-openshift:1.0", cr.Spec.CommonConfig.Version), cr)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhdm%s-decisioncentral-openshift:1.0", cr.Spec.CommonConfig.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhdm-7/rhdm%s-decisioncentral-openshift:1.0", cr.Spec.CommonConfig.Version), isTag.Tag.From.Name)
}

func TestCreateRhdmTechPreviewImageStreams(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhdm-trial",
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	reconciler.createLocalImageTag(fmt.Sprintf("rhdm%s-decisioncentral-indexing-openshift:1.0", cr.Spec.CommonConfig.Version), cr)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhdm%s-decisioncentral-indexing-openshift:1.0", cr.Spec.CommonConfig.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhdm-7-tech-preview/rhdm%s-decisioncentral-indexing-openshift:1.0", cr.Spec.CommonConfig.Version), isTag.Tag.From.Name)
}

func TestCreateImageStreamsLatest(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhdm-trial",
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	reconciler.createLocalImageTag(fmt.Sprintf("rhdm%s-decisioncentral-indexing-openshift", cr.Spec.CommonConfig.Version), cr)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhdm%s-decisioncentral-indexing-openshift:latest", cr.Spec.CommonConfig.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhdm-7-tech-preview/rhdm%s-decisioncentral-indexing-openshift:latest", cr.Spec.CommonConfig.Version), isTag.Tag.From.Name)
}
