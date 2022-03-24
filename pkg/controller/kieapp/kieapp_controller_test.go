package kieapp

import (
	"context"
	"fmt"
	oimagev1 "github.com/openshift/api/image/v1"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/google/uuid"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var depMessage = "Deployment should be completed"
var caConfigMap = &corev1.ConfigMap{}

func TestGenerateSecret(t *testing.T) {
	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := Reconciler{
		Service: mockService,
	}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: defaults.Pint(3)},
				},
			},
		},
	}
	env, err := defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	assert.Len(t, env.Console.Secrets, 0, "No secret is available when reading the trial workbench from yaml files")

	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	//assert.Len(t, env.Console.Secrets, 1, "One secret should be generated for the trial workbench")
	//assert.Len(t, env.Servers[0].Secrets, 1, "One secret should be generated for each trial kieserver")
	generateSecretCommonAssertions(t, env)

	consoleSecret := env.Console.Secrets[0]
	serverSecret := env.Servers[0].Secrets[0]
	consoleRoute := cr.Status.ConsoleHost
	secretName := fmt.Sprintf(constants.KeystoreSecret, env.Servers[0].DeploymentConfigs[0].Name)
	assert.Equal(t, secretName, serverSecret.Name)
	for _, volume := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, secretName, volume.Secret.SecretName)
		}
	}

	err = reconciler.Service.Create(context.TODO(), &consoleSecret)
	assert.Nil(t, err)
	err = reconciler.Service.Create(context.TODO(), &serverSecret)
	assert.Nil(t, err)
	consoleTestSecret := corev1.Secret{}
	err = reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: consoleSecret.Name, Namespace: consoleSecret.Namespace}, &consoleTestSecret)
	assert.Nil(t, err)

	consoleCN := reconciler.setConsoleHost(cr, env, getRequestedRoutes(env, cr))
	assert.Equal(t, consoleRoute, "http://"+consoleCN)
	ok, err := shared.IsValidKeyStoreSecret(corev1.Secret{}, consoleCN, []byte(cr.Status.Applied.CommonConfig.KeyStorePassword))
	assert.False(t, ok)
	assert.Nil(t, err)
	ok, err = shared.IsValidKeyStoreSecret(consoleTestSecret, "blah", []byte(cr.Status.Applied.CommonConfig.KeyStorePassword))
	assert.False(t, ok)
	assert.Nil(t, err)
	ok, err = shared.IsValidKeyStoreSecret(consoleTestSecret, consoleCN, []byte("wrongPwd"))
	assert.False(t, ok)
	assert.NotNil(t, err)
	ok, err = shared.IsValidKeyStoreSecret(consoleTestSecret, consoleCN, []byte(cr.Status.Applied.CommonConfig.KeyStorePassword))
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, consoleSecret.DeepCopy(), consoleTestSecret.DeepCopy())

	secret, err := reconciler.generateKeystoreSecret(
		fmt.Sprintf(constants.KeystoreSecret, strings.Join([]string{cr.Status.Applied.CommonConfig.ApplicationName, "businesscentral"}, "-")),
		consoleCN,
		cr,
	)
	assert.Nil(t, err)
	assert.Equal(t, consoleSecret, secret)

	// secrets should be identical between reconciles since nothing has changed in CR
	env = api.Environment{}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	assertSecret(t, consoleSecret, serverSecret, env, true)

	// change app name which should change commonname and keystore secret
	currentRoute := cr.Status.ConsoleHost
	cr.Spec.CommonConfig.ApplicationName = "changed"
	env = api.Environment{}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	assert.NotEqual(t, currentRoute, cr.Status.ConsoleHost)
	generateSecretCommonAssertions(t, env)
	assertSecret(t, consoleSecret, serverSecret, env, false)
	err = reconciler.Service.Delete(context.TODO(), &consoleSecret)
	assert.Nil(t, err)
	err = reconciler.Service.Delete(context.TODO(), &serverSecret)
	assert.Nil(t, err)

	consoleSecret = env.Console.Secrets[0]
	serverSecret = env.Servers[0].Secrets[0]
	err = reconciler.Service.Create(context.TODO(), &consoleSecret)
	assert.Nil(t, err)
	err = reconciler.Service.Create(context.TODO(), &serverSecret)
	assert.Nil(t, err)

	// secrets should be identical between reconciles since nothing has changed in CR
	env = api.Environment{}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	assertSecret(t, consoleSecret, serverSecret, env, true)

	// change keystore password which should change commonname and keystore secret
	oldPassword := cr.Status.Applied.CommonConfig.KeyStorePassword
	cr.Spec.CommonConfig.KeyStorePassword = "changed"
	env = api.Environment{}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	assert.NotEqual(t, oldPassword, cr.Status.Applied.CommonConfig.KeyStorePassword)
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	generateSecretCommonAssertions(t, env)

	assertSecret(t, consoleSecret, serverSecret, env, false)
	err = reconciler.Service.Delete(context.TODO(), &consoleSecret)
	assert.Nil(t, err)
	err = reconciler.Service.Delete(context.TODO(), &serverSecret)
	assert.Nil(t, err)

	consoleSecret = env.Console.Secrets[0]
	serverSecret = env.Servers[0].Secrets[0]
	err = reconciler.Service.Create(context.TODO(), &consoleSecret)
	assert.Nil(t, err)
	err = reconciler.Service.Create(context.TODO(), &serverSecret)
	assert.Nil(t, err)

	// secrets should be identical between reconciles since nothing has changed in CR
	env = api.Environment{}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	assertSecret(t, consoleSecret, serverSecret, env, true)
}

func generateSecretCommonAssertions(t *testing.T, env api.Environment) {
	assert.Len(t, env.Console.Secrets, 1, "One secret should be generated for the trial workbench")
	assert.Len(t, env.Servers[0].Secrets, 1, "One secret should be generated for each trial kieserver")
}

func assertSecret(t *testing.T, consoleSecret corev1.Secret, serverSecret corev1.Secret, env api.Environment, equal bool) {
	if equal {
		assert.Equal(t, consoleSecret, env.Console.Secrets[0])
		assert.Equal(t, serverSecret, env.Servers[0].Secrets[0])
	} else {
		assert.NotEqual(t, consoleSecret, env.Console.Secrets[0])
		assert.NotEqual(t, serverSecret, env.Servers[0].Secrets[0])
	}

}

func TestGenerateTruststoreSecret(t *testing.T) {
	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := Reconciler{
		Service: mockService,
	}

	caBundle, err := ioutil.ReadFile("shared/test-" + constants.CaBundleKey)
	assert.Nil(t, err)
	assert.NotEmpty(t, caBundle)

	caConfigMap = &corev1.ConfigMap{
		Data: map[string]string{
			constants.CaBundleKey: string(caBundle),
		},
	}

	cr := &api.KieApp{ObjectMeta: metav1.ObjectMeta{Name: "test"}}
	secret, err := reconciler.generateTruststoreSecret(
		cr.Status.Applied.CommonConfig.ApplicationName+constants.TruststoreSecret,
		cr,
		caConfigMap,
	)
	assert.Nil(t, err)
	assert.Equal(t, cr.Status.Applied.CommonConfig.ApplicationName+constants.TruststoreSecret, secret.Name)
	ok, err := shared.IsValidTruststoreSecret(secret, caBundle)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestGenerateSecrets(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
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

	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error getting a new environment")
	reconciler := Reconciler{
		Service: mockService,
	}
	env, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						KeystoreSecret: "console-ks-secret",
					},
				},
				Servers: []api.KieServerSet{
					{
						KieAppObject: api.KieAppObject{
							KeystoreSecret: "server-ks-secret",
						},
					},
				},
				SmartRouter: &api.SmartRouterObject{
					KieAppObject: api.KieAppObject{
						KeystoreSecret: "smartrouter-ks-secret",
					},
				},
			},
		},
	}
	env, err := defaults.GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting a new environment")
	assert.Len(t, env.Console.Secrets, 0, "No secret is available when reading the trial workbench from yaml files")

	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	env, err = defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error creating a new environment")
	assert.Len(t, env.Console.Secrets, 0, "Zero secrets should be generated for the trial workbench")
	assert.Len(t, env.Servers[0].Secrets, 0, "Zero secrets should be generated for the trial kieserver")
	assert.Len(t, env.SmartRouter.Secrets, 0, "Zero secrets should be generated for the smartrouter")
	for _, volume := range env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Status.Applied.Objects.Console.KeystoreSecret, volume.Secret.SecretName)
		}
	}
	for _, volume := range env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Status.Applied.Objects.Servers[0].KeystoreSecret, volume.Secret.SecretName)
		}
	}
	for _, volume := range env.SmartRouter.DeploymentConfigs[0].Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			assert.Equal(t, cr.Status.Applied.Objects.SmartRouter.KeystoreSecret, volume.Secret.SecretName)
		}
	}
}

func TestConsoleHost(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
		},
	}

	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	env, err := defaults.GetEnvironment(cr, mockService)
	assert.Nil(t, err, "Error creating a new environment")
	reconciler := &Reconciler{Service: mockService}
	_, err = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr), caConfigMap)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("http://%s", cr.Name), cr.Status.ConsoleHost, "status.ConsoleHost should be URL from the resulting workbench route host")
}

func TestVerifyExternalReferencesRoleMapper(t *testing.T) {
	tests := []struct {
		name       string
		roleMapper *api.RoleMapperAuthConfig
		errMsg     string
	}{{
		name:       "Empty reference",
		roleMapper: &api.RoleMapperAuthConfig{},
	}, {
		name: "Unsupported Kind: Service",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "Service",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "unsupported Kind: Service",
	}, {
		name: "Not found ConfigMap",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "ConfigMap",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found Secret",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "Secret",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found PersistentVolumeClaim",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Found ConfigMap",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "ConfigMap",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}, {
		name: "Found Secret",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "Secret",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}, {
		name: "Found PersistentVolumeClaim",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Auth:        &api.KieAppAuthObject{},
		},
	}
	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}

	reconciler := &Reconciler{Service: mockService}
	for _, test := range tests {
		mockService.GetFunc = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
			if test.errMsg == "" {
				return nil
			}
			return fmt.Errorf("Mock: Not found")
		}
		cr.Spec.Auth.RoleMapper = test.roleMapper
		_, err = defaults.GetEnvironment(cr, reconciler.Service)
		assert.NotNil(t, err)
		err = reconciler.verifyExternalReferences(cr)
		if test.errMsg == "" {
			assert.Nil(t, err, "%s: Expected nil found [%s]", test.name, err)
		} else {
			assert.Error(t, err, "%s: Expected error [%s]", test.name, test.errMsg)
			if err != nil {
				assert.EqualError(t, err, test.errMsg, "Test case %s got an Unexpected error", test.name)
			}
		}
	}
}

func TestVerifyExternalReferencesGitHooks(t *testing.T) {
	tests := []struct {
		name     string
		gitHooks *api.GitHooksVolume
		errMsg   string
	}{{
		name:     "Empty reference",
		gitHooks: &api.GitHooksVolume{},
	}, {
		name: "Unsupported type",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "Service",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "unsupported Kind: Service",
	}, {
		name: "Not found ConfigMap",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "ConfigMap",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found Secret",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "Secret",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found PersistentVolumeClaim",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Found ConfigMap",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "ConfigMap",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}, {
		name: "Found Secret",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "Secret",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}, {
		name: "Found PersistentVolumeClaim",
		gitHooks: &api.GitHooksVolume{
			From: &api.ObjRef{
				Kind: "PersistentVolumeClaim",
				ObjectReference: api.ObjectReference{
					Name: "test",
				},
			},
		},
	}}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{},
			},
		},
	}
	scheme, err := api.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")
	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}
	reconciler := &Reconciler{Service: mockService}

	for _, test := range tests {
		mockService.GetFunc = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
			if test.errMsg == "" {
				return nil
			}
			return fmt.Errorf("Mock: Not found")
		}
		cr.Spec.Objects.Console.GitHooks = test.gitHooks
		defaults.SetDefaults(cr)
		err = reconciler.verifyExternalReferences(cr)
		if test.errMsg == "" {
			assert.Nil(t, err, "%s: Expected nil found [%s]", test.name, err)
		} else {
			assert.Error(t, err, "%s: Expected error [%s]", test.name, test.errMsg)
			if err != nil {
				assert.EqualError(t, err, test.errMsg, "Test case %s got an Unexpected error", test.name)
			}
		}
	}
}

func TestCreateRhpamImageStreams(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Objects: api.KieAppObjects{
				Console: &api.ConsoleObject{
					KieAppObject: api.KieAppObject{
						ImageContext: "test",
					},
				},
			},
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	image := fmt.Sprintf("rhpam-businesscentral-openshift:%s", cr.Status.Applied.Version)
	imageURL := constants.ImageRegistry + "/" + cr.Spec.Objects.Console.ImageContext + "/" + image
	err = reconciler.createLocalImageTag(image, imageURL, cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(context.TODO(), cr.Namespace+"/"+image, metav1.GetOptions{})
	assert.Nil(t, err)

	assert.NotNil(t, isTag)
	assert.Equal(t, imageURL, isTag.Tag.From.Name)
	assert.Equal(t, image, isTag.Name)
	assert.Equal(t, cr.Status.Applied.Version, isTag.Tag.Name)
	assert.Equal(t, cr.Namespace, isTag.Namespace)
	assert.False(t, isTag.Tag.ImportPolicy.Scheduled)
}

func TestCreateRhdmImageStreams(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	err = reconciler.createLocalImageTag(fmt.Sprintf("rhpam%s-businesscentral-openshift:1.0", cr.Status.Applied.Version), "", cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(context.TODO(), fmt.Sprintf("test-ns/rhpam%s-businesscentral-openshift:1.0", cr.Status.Applied.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhpam-7/rhpam%s-businesscentral-openshift:1.0", cr.Status.Applied.Version), isTag.Tag.From.Name)
}

// TODO remove after 7.12.1 is not a supported version for the current operator version and point to rhpam images
func TestCreateRhdmImageStreamsFor7121(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Version:     "7.12.1",
			Environment: api.RhdmTrial,
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	err = reconciler.createLocalImageTag(fmt.Sprintf("rhdm%s-decisioncentral-openshift:1.0", cr.Status.Applied.Version), "", cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(context.TODO(), fmt.Sprintf("test-ns/rhdm%s-decisioncentral-openshift:1.0", cr.Status.Applied.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhdm-7/rhdm%s-decisioncentral-openshift:1.0", cr.Status.Applied.Version), isTag.Tag.From.Name)
	assert.False(t, isTag.Tag.ImportPolicy.Scheduled)
}

func getISTag(mockSvc *test.MockPlatformService, cr *api.KieApp, tagRefName string, imageName string) (*oimagev1.ImageStreamTag, error) {
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	if err != nil {
		return nil, err
	}
	reconciler := Reconciler{
		Service: mockSvc,
	}
	err = reconciler.createLocalImageTag(tagRefName, "", cr)
	isTag, err := isTagMock.Get(context.TODO(), imageName, metav1.GetOptions{})
	return isTag, err
}

func TestCreateRhpamImageStreamsUsingImageTags(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment:  api.RhpamAuthoring,
			UseImageTags: true,
		},
	}
	mockSvc := test.MockService()
	isTag, err := getISTag(mockSvc, cr, fmt.Sprintf("rhpam%s-kieserver-openshift:1.0", cr.Status.Applied.Version), fmt.Sprintf("test-ns/rhpam%s-kieserver-openshift:1.0", cr.Status.Applied.Version))
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.False(t, isTag.Tag.ImportPolicy.Scheduled)
}

func TestCreateRhpamImageStreamsUsingImageTagsWithScheduledImportPolicy(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Upgrades: api.KieAppUpgrades{
				ScheduledImportPolicy: true,
			},
			Environment:  api.RhpamProduction,
			UseImageTags: true,
		},
	}
	mockSvc := test.MockService()
	isTag, err := getISTag(mockSvc, cr, fmt.Sprintf("rhpam%s-kieserver-openshift:1.0", cr.Status.Applied.Version), fmt.Sprintf("test-ns/rhpam%s-kieserver-openshift:1.0", cr.Status.Applied.Version))
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.True(t, isTag.Tag.ImportPolicy.Scheduled)
}

func TestCreateTagVersionImageStreams(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	err = reconciler.createLocalImageTag(fmt.Sprintf("%s:%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), "", cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(context.TODO(), fmt.Sprintf("test-ns/%s:%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("%s/jboss-datagrid-7/%s:%s", constants.ImageRegistry, constants.VersionConstants[cr.Status.Applied.Version].DatagridImage, constants.VersionConstants[cr.Status.Applied.Version].DatagridImageTag), isTag.Tag.From.Name)
	assert.False(t, isTag.Tag.ImportPolicy.Scheduled)
}

func TestCreateImageStreamsLatest(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	err = reconciler.createLocalImageTag(fmt.Sprintf("%s", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage), "", cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(context.TODO(), fmt.Sprintf("test-ns/%s:latest", constants.VersionConstants[cr.Status.Applied.Version].DatagridImage), metav1.GetOptions{})
	assert.Nil(t, err)
	fmt.Print(isTag)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("%s/jboss-datagrid-7/%s:latest", constants.ImageRegistry, constants.VersionConstants[cr.Status.Applied.Version].DatagridImage), isTag.Tag.From.Name)
	assert.False(t, isTag.Tag.ImportPolicy.Scheduled)
}

func TestStatusDeploymentsProgression(t *testing.T) {
	crNamespacedName := getNamespacedName("namespace", "cr")
	cr := getInstance(crNamespacedName)
	cr.Spec = api.KieAppSpec{
		Environment: api.RhpamTrial,
	}
	service := test.MockService()
	err := service.Create(context.TODO(), cr)
	assert.Nil(t, err)
	reconciler := Reconciler{Service: service}
	result, err := reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Duration(500) * time.Millisecond}, result, "Routes should be created, requeued for hostname detection before other resources are created")

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "All other resources created, custom Resource status set to provisioning, and requeued")
	assert.Nil(t, err)

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment status set, and requeued")
	assert.Nil(t, err)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.NotEmpty(t, cr.Status.Conditions)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 2, "Expect 2 stopped deployments")

	//Let's now assume console pod is starting
	service.ListFunc = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
		err := service.Client.List(ctx, list, opts...)
		if err == nil && reflect.TypeOf(list) == reflect.TypeOf(&oappsv1.DeploymentConfigList{}) {
			for index := range list.(*oappsv1.DeploymentConfigList).Items {
				dc := &list.(*oappsv1.DeploymentConfigList).Items[index]
				if dc.Name == "cr-rhpamcentr" {
					dc.Status.Replicas = 1
				}
			}
		}
		return err
	}

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result, depMessage)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 1, "Expect 1 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 1, "Expect 1 deployment starting up")

	//Let's now assume both pods have started
	service.ListFunc = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
		err := service.Client.List(ctx, list, opts...)
		if err == nil && reflect.TypeOf(list) == reflect.TypeOf(&oappsv1.DeploymentConfigList{}) {
			for index := range list.(*oappsv1.DeploymentConfigList).Items {
				dc := &list.(*oappsv1.DeploymentConfigList).Items[index]
				dc.Status.Replicas = 1
				dc.Status.ReadyReplicas = 1
			}
		}
		return err
	}

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result, depMessage)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 0, "Expect 0 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 0, "Expect 0 deployment starting up")
	assert.Len(t, cr.Status.Deployments.Ready, 2, "Expect 2 deployment to be ready")
}

func TestConsoleLinkCreation(t *testing.T) {
	crNamespacedName := getNamespacedName("testns", "cr")
	cr := getInstance(crNamespacedName)
	cr.Spec = api.KieAppSpec{
		Environment: api.RhpamAuthoring,
	}
	service := test.MockService()
	err := service.Create(context.TODO(), cr)
	assert.Nil(t, err)
	reconciler := Reconciler{Service: service}
	result, err := reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Duration(500) * time.Millisecond}, result, "Routes should be created, requeued for hostname detection before other resources are created")

	// Simulate server setting the host
	bcRoute := &routev1.Route{}
	reconciler.Service.Get(context.TODO(), getNamespacedName(cr.Namespace, "cr-rhpamcentr"), bcRoute)
	bcRoute.Spec.Host = "example"
	reconciler.Service.Update(context.TODO(), bcRoute)

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "All other resources created, custom Resource status set to provisioning, and requeued")
	assert.Nil(t, err)

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment status set, and requeued")
	assert.Nil(t, err)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 2, "Expect 2 stopped deployments")

	//Let's now assume console pod is starting
	service.ListFunc = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
		err := service.Client.List(ctx, list, opts...)
		if err == nil && reflect.TypeOf(list) == reflect.TypeOf(&oappsv1.DeploymentConfigList{}) {
			for index := range list.(*oappsv1.DeploymentConfigList).Items {
				dc := &list.(*oappsv1.DeploymentConfigList).Items[index]
				if dc.Name == "cr-rhpamcentr" {
					dc.Status.Replicas = 1
				}
			}
		}
		return err
	}

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result, depMessage)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 1, "Expect 1 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 1, "Expect 1 deployment starting up")

	//Let's now assume both pods have started
	service.ListFunc = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
		err := service.Client.List(ctx, list, opts...)
		if err == nil && reflect.TypeOf(list) == reflect.TypeOf(&oappsv1.DeploymentConfigList{}) {
			for index := range list.(*oappsv1.DeploymentConfigList).Items {
				dc := &list.(*oappsv1.DeploymentConfigList).Items[index]
				dc.Status.Replicas = 1
				dc.Status.ReadyReplicas = 1
			}
		}
		return err
	}

	result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result, depMessage)

	cr, err = reloadCR(t, service, crNamespacedName)
	assert.Nil(t, err)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 0, "Expect 0 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 0, "Expect 0 deployment starting up")
	assert.Len(t, cr.Status.Deployments.Ready, 2, "Expect 2 deployment to be ready")

	consoleLink := &consolev1.ConsoleLink{}
	reconciler.Service.Get(context.TODO(), getNamespacedName("", getConsoleLinkName(cr)), consoleLink)
	assert.Equal(t, "cr: Business Central", consoleLink.Spec.Text)
	assert.Equal(t, "https://example", consoleLink.Spec.Href)
	assert.Equal(t, "testns-link-cr", consoleLink.GetName())
	assert.Len(t, consoleLink.Spec.NamespaceDashboard.Namespaces, 1)
	assert.Contains(t, consoleLink.Spec.NamespaceDashboard.Namespaces, "testns")

	//Delete kieapp
	deletionTimestamp := metav1.Now()
	cr.SetDeletionTimestamp(&deletionTimestamp)
	err = service.Update(context.TODO(), cr)
	assert.Nil(t, err)

	_, err = reconciler.Reconcile(context.TODO(), reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	cr, err = reloadCR(t, service, crNamespacedName)
	// it now returns the error.StatusError, in this case:
	// Status:"Failure", Message:"kieapps.app.kiegroup.org \"cr\" not found", Reason:"NotFound",
	// Details:(*v1.StatusDetails)(0xc0003d1320), Code:404}
	assert.NotNil(t, err)
	assert.Len(t, cr.GetFinalizers(), 0)
	consoleLink = &consolev1.ConsoleLink{}
	err = reconciler.Service.Get(context.TODO(), getNamespacedName("", "testns-cr-0"), consoleLink)
	assert.Error(t, err, "ConsoleLink must have been removed by the Finalizer")
}

func getNamespacedName(namespace string, name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func reloadCR(t *testing.T, service *test.MockPlatformService, namespacedName types.NamespacedName) (*api.KieApp, error) {
	cr := getInstance(namespacedName)
	err := service.Get(context.TODO(), namespacedName, cr)
	return cr, err
}

func getInstance(namespacedName types.NamespacedName) *api.KieApp {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			UID:       types.UID(uuid.New().String()),
		},
	}
	return cr
}

func TestGetComparatorConfigMap(t *testing.T) {
	comparator := getComparator()

	type args struct {
		deployed  map[reflect.Type][]client.Object
		requested map[reflect.Type][]client.Object
	}
	tests := []struct {
		name string
		args args
		want compare.ResourceDelta
	}{
		{
			"NoChange",
			args{
				deployed: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
				requested: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
			},
			compare.ResourceDelta{},
		},
		{
			"Add",
			args{
				deployed: map[reflect.Type][]client.Object{},
				requested: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
			},
			compare.ResourceDelta{
				Added: []client.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
							Labels: map[string]string{
								"test": "test",
							},
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Data: map[string]string{
							"test": "test",
						},
						BinaryData: map[string][]byte{
							"test": {'t', 'e', 's', 't'},
						},
					},
				},
			},
		},
		{
			"Removed",
			args{
				deployed: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
				requested: map[reflect.Type][]client.Object{},
			},
			compare.ResourceDelta{
				Removed: []client.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
							Labels: map[string]string{
								"test": "test",
							},
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Data: map[string]string{
							"test": "test",
						},
						BinaryData: map[string][]byte{
							"test": {'t', 'e', 's', 't'},
						},
					},
				},
			},
		},
		{
			"UpdatedData",
			args{
				deployed: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
				requested: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test1",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
			},
			compare.ResourceDelta{
				Updated: []client.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
							Labels: map[string]string{
								"test": "test",
							},
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Data: map[string]string{
							"test": "test1",
						},
						BinaryData: map[string][]byte{
							"test": {'t', 'e', 's', 't'},
						},
					},
				},
			},
		},
		{
			"UpdatedBinaryData",
			args{
				deployed: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
				requested: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't', '1'},
							},
						},
					},
				},
			},
			compare.ResourceDelta{
				Updated: []client.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
							Labels: map[string]string{
								"test": "test",
							},
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Data: map[string]string{
							"test": "test",
						},
						BinaryData: map[string][]byte{
							"test": {'t', 'e', 's', 't', '1'},
						},
					},
				},
			},
		},
		{
			"UpdatedLabels",
			args{
				deployed: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
				requested: map[reflect.Type][]client.Object{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
								Labels: map[string]string{
									"test": "test1",
								},
								Annotations: map[string]string{
									"test": "test",
								},
							},
							Data: map[string]string{
								"test": "test",
							},
							BinaryData: map[string][]byte{
								"test": {'t', 'e', 's', 't'},
							},
						},
					},
				},
			},
			compare.ResourceDelta{
				Updated: []client.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
							Labels: map[string]string{
								"test": "test1",
							},
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Data: map[string]string{
							"test": "test",
						},
						BinaryData: map[string][]byte{
							"test": {'t', 'e', 's', 't'},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := comparator.Compare(tt.args.deployed, tt.args.requested)[reflect.TypeOf(corev1.ConfigMap{})]; !ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getComparator_ConfigMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
