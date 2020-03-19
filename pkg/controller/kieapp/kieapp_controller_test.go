package kieapp

import (
	"context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/google/uuid"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	"reflect"
	"testing"
	"time"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestGenerateSecret(t *testing.T) {
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
	env = reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr))
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
				Console: api.ConsoleObject{
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
	reconciler.setEnvironmentProperties(cr, env, getRequestedRoutes(env, cr))
	assert.Equal(t, fmt.Sprintf("http://%s", cr.Spec.CommonConfig.ApplicationName), cr.Status.ConsoleHost, "spec.commonConfig.consoleHost should be URL from the resulting workbench route host")
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
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Service",
			},
		},
		errMsg: "unsupported Kind: Service",
	}, {
		name: "Not found ConfigMap",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "ConfigMap",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found Secret",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Secret",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found PersistentVolumeClaim",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "PersistentVolumeClaim",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Found ConfigMap",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "ConfigMap",
			},
		},
	}, {
		name: "Found Secret",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Secret",
			},
		},
	}, {
		name: "Found PersistentVolumeClaim",
		roleMapper: &api.RoleMapperAuthConfig{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "PersistentVolumeClaim",
			},
		},
	}}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhdmTrial,
			Auth:        api.KieAppAuthObject{},
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
		mockService.GetFunc = func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			if test.errMsg == "" {
				return nil
			}
			return fmt.Errorf("Mock: Not found")
		}
		cr.Spec.Auth.RoleMapper = test.roleMapper
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
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Service",
			},
		},
		errMsg: "unsupported Kind: Service",
	}, {
		name: "Not found ConfigMap",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "ConfigMap",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found Secret",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Secret",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Not found PersistentVolumeClaim",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "PersistentVolumeClaim",
			},
		},
		errMsg: "Mock: Not found",
	}, {
		name: "Found ConfigMap",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "ConfigMap",
			},
		},
	}, {
		name: "Found Secret",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "Secret",
			},
		},
	}, {
		name: "Found PersistentVolumeClaim",
		gitHooks: &api.GitHooksVolume{
			From: &corev1.ObjectReference{
				Name: "test",
				Kind: "PersistentVolumeClaim",
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
				Console: api.ConsoleObject{},
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
		mockService.GetFunc = func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			if test.errMsg == "" {
				return nil
			}
			return fmt.Errorf("Mock: Not found")
		}
		cr.Spec.Objects.Console.GitHooks = test.gitHooks
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
		},
	}
	mockSvc := test.MockService()
	isTagMock := mockSvc.ImageStreamTagsFunc(cr.Namespace)
	_, err := defaults.GetEnvironment(cr, mockSvc)
	assert.Nil(t, err)
	reconciler := Reconciler{
		Service: mockSvc,
	}

	err = reconciler.createLocalImageTag(fmt.Sprintf("rhpam%s-businesscentral-openshift:1.0", cr.Spec.Version), cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhpam%s-businesscentral-openshift:1.0", cr.Spec.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhpam-7/rhpam%s-businesscentral-openshift:1.0", cr.Spec.Version), isTag.Tag.From.Name)
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

	err = reconciler.createLocalImageTag(fmt.Sprintf("rhdm%s-decisioncentral-openshift:1.0", cr.Spec.Version), cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/rhdm%s-decisioncentral-openshift:1.0", cr.Spec.Version), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("registry.redhat.io/rhdm-7/rhdm%s-decisioncentral-openshift:1.0", cr.Spec.Version), isTag.Tag.From.Name)
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

	err = reconciler.createLocalImageTag(fmt.Sprintf("%s:%s", constants.VersionConstants[cr.Spec.Version].DatagridImage, constants.VersionConstants[cr.Spec.Version].DatagridImageTag), cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/%s:%s", constants.VersionConstants[cr.Spec.Version].DatagridImage, constants.VersionConstants[cr.Spec.Version].DatagridImageTag), metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("%s/jboss-datagrid-7/%s:%s", constants.ImageRegistry, constants.VersionConstants[cr.Spec.Version].DatagridImage, constants.VersionConstants[cr.Spec.Version].DatagridImageTag), isTag.Tag.From.Name)
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

	err = reconciler.createLocalImageTag(fmt.Sprintf("%s", constants.VersionConstants[cr.Spec.Version].DatagridImage), cr)
	assert.Nil(t, err)

	isTag, err := isTagMock.Get(fmt.Sprintf("test-ns/%s:latest", constants.VersionConstants[cr.Spec.Version].DatagridImage), metav1.GetOptions{})
	assert.Nil(t, err)
	fmt.Print(isTag)
	assert.NotNil(t, isTag)
	assert.Equal(t, fmt.Sprintf("%s/jboss-datagrid-7/%s:latest", constants.ImageRegistry, constants.VersionConstants[cr.Spec.Version].DatagridImage), isTag.Tag.From.Name)
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
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Duration(500) * time.Millisecond}, result, "Routes should be created, requeued for hostname detection before other resources are created")

	result, err = reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "All other resources created, custom Resource status set to provisioning, and requeued")
	assert.Nil(t, err)

	result, err = reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment status set, and requeued")
	assert.Nil(t, err)

	cr = reloadCR(t, service, crNamespacedName)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 2, "Expect 2 stopped deployments")

	//Let's now assume console pod is starting
	service.ListFunc = func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
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

	result, err = reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment should be created but requeued for status updates")

	cr = reloadCR(t, service, crNamespacedName)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 1, "Expect 1 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 1, "Expect 1 deployment starting up")

	//Let's now assume both pods have started
	service.ListFunc = func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
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

	result, err = reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment should be created but requeued for status updates")

	cr = reloadCR(t, service, crNamespacedName)
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
	extReconciler := kubernetes.NewExtendedReconciler(service, &reconciler, &api.KieApp{})
	extReconciler.RegisterFinalizer(&ConsoleLinkFinalizer{})
	result, err := extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Duration(500) * time.Millisecond}, result, "Routes should be created, requeued for hostname detection before other resources are created")

	// Simulate server setting the host
	bcRoute := &routev1.Route{}
	extReconciler.Service.Get(context.TODO(), getNamespacedName(cr.Namespace, "cr-rhpamcentr"), bcRoute)
	bcRoute.Spec.Host = "example"
	extReconciler.Service.Update(context.TODO(), bcRoute)

	result, err = extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "All other resources created, custom Resource status set to provisioning, and requeued")
	assert.Nil(t, err)

	result, err = extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment status set, and requeued")
	assert.Nil(t, err)

	cr = reloadCR(t, service, crNamespacedName)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 2, "Expect 2 stopped deployments")

	//Let's now assume console pod is starting
	service.ListFunc = func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
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

	result, err = extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment should be created but requeued for status updates")

	cr = reloadCR(t, service, crNamespacedName)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 1, "Expect 1 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 1, "Expect 1 deployment starting up")

	//Let's now assume both pods have started
	service.ListFunc = func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
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

	result, err = extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true}, result, "Deployment should be created but requeued for status updates")

	cr = reloadCR(t, service, crNamespacedName)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Len(t, cr.Status.Deployments.Stopped, 0, "Expect 0 stopped deployments")
	assert.Len(t, cr.Status.Deployments.Starting, 0, "Expect 0 deployment starting up")
	assert.Len(t, cr.Status.Deployments.Ready, 2, "Expect 2 deployment to be ready")

	assert.Len(t, cr.GetFinalizers(), 1)
	assert.Equal(t, constants.ConsoleLinkFinalizer, cr.GetFinalizers()[0])

	consoleLink := &consolev1.ConsoleLink{}
	extReconciler.Service.Get(context.TODO(), getNamespacedName("", string(cr.GetUID())), consoleLink)
	assert.Equal(t, "Business Central", consoleLink.Spec.Text)
	assert.Equal(t, "https://example", consoleLink.Spec.Href)
	assert.Equal(t, string(cr.GetUID()), consoleLink.GetName())
	assert.Len(t, consoleLink.Spec.NamespaceDashboard.Namespaces, 1)
	assert.Contains(t, consoleLink.Spec.NamespaceDashboard.Namespaces, "testns")

	//Delete kieapp
	deletionTimestamp := metav1.Now()
	cr.SetDeletionTimestamp(&deletionTimestamp)
	err = service.Update(context.TODO(), cr)
	assert.Nil(t, err)

	result, err = extReconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	cr = reloadCR(t, service, crNamespacedName)
	assert.Len(t, cr.GetFinalizers(), 0)
	consoleLink = &consolev1.ConsoleLink{}
	err = extReconciler.Service.Get(context.TODO(), getNamespacedName("", "testns-cr-0"), consoleLink)
	assert.Error(t, err, "ConsoleLink must have been removed by the Finalizer")
}

func getNamespacedName(namespace string, name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func reloadCR(t *testing.T, service *test.MockPlatformService, namespacedName types.NamespacedName) *api.KieApp {
	cr := getInstance(namespacedName)
	err := service.Get(context.TODO(), namespacedName, cr)
	assert.Nil(t, err)
	return cr
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
