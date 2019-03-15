package kieapp

import (
	"fmt"
	"testing"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

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
