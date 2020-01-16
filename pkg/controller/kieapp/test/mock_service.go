package test

import (
	"context"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logs.GetLogger("kieapp.test")

type MockPlatformService struct {
	Client              clientv1.Client
	scheme              *runtime.Scheme
	CreateFunc          func(ctx context.Context, obj runtime.Object, opts ...clientv1.CreateOption) error
	DeleteFunc          func(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error
	GetFunc             func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error
	ListFunc            func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error
	UpdateFunc          func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error
	UpdateStatusFunc    func(ctx context.Context, obj runtime.Object) error
	PatchFunc           func(ctx context.Context, obj runtime.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error
	DeleteAllOfFunc     func(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteAllOfOption) error
	GetCachedFunc       func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error
	ImageStreamTagsFunc func(namespace string) imagev1.ImageStreamTagInterface
	GetSchemeFunc       func() *runtime.Scheme
}

func MockService() *MockPlatformService {
	return MockServiceWithExtraScheme()
}

func MockServiceWithExtraScheme(objs ...runtime.Object) *MockPlatformService {
	registerObjs := []runtime.Object{&api.KieApp{}, &api.KieAppList{}, &corev1.PersistentVolumeClaim{}, &corev1.ServiceAccount{}, &corev1.Secret{}, &rbacv1.Role{}, &rbacv1.RoleBinding{}, &oappsv1.DeploymentConfig{}, &corev1.Service{}, &appsv1.StatefulSet{}, &routev1.Route{}, &oimagev1.ImageStream{}, &buildv1.BuildConfig{}, &oappsv1.DeploymentConfigList{}, &buildv1.BuildConfigList{}, &corev1.PersistentVolumeClaimList{}, &corev1.ServiceAccountList{}, &rbacv1.RoleList{}, &rbacv1.RoleBindingList{}, &corev1.ServiceList{}, &appsv1.StatefulSetList{}, &routev1.RouteList{}, &oimagev1.ImageStreamList{}}
	registerObjs = append(registerObjs, objs...)
	api.SchemeBuilder.Register(registerObjs...)
	scheme, _ := api.SchemeBuilder.Build()
	client := fake.NewFakeClientWithScheme(scheme)
	log.Debugf("Fake client created as %v", client)
	mockImageStreamTag := &MockImageStreamTag{}
	return &MockPlatformService{
		Client: client,
		scheme: scheme,
		CreateFunc: func(ctx context.Context, obj runtime.Object, opts ...clientv1.CreateOption) error {
			return client.Create(ctx, obj, opts...)
		},
		DeleteFunc: func(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error {
			return client.Delete(ctx, obj, opts...)
		},
		GetFunc: func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			return client.Get(ctx, key, obj)
		},
		ListFunc: func(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
			return client.List(ctx, list, opts...)
		},
		UpdateFunc: func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
			return client.Update(ctx, obj, opts...)
		},
		PatchFunc: func(ctx context.Context, obj runtime.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
			return client.Patch(ctx, obj, patch, opts...)
		},
		DeleteAllOfFunc: func(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteAllOfOption) error {
			return client.DeleteAllOf(ctx, obj, opts...)
		},
		UpdateStatusFunc: func(ctx context.Context, obj runtime.Object) error {
			return client.Status().Update(ctx, obj)
		},
		GetCachedFunc: func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			return client.Get(ctx, key, obj)
		},
		ImageStreamTagsFunc: func(namespace string) imagev1.ImageStreamTagInterface {
			return mockImageStreamTag
		},
		GetSchemeFunc: func() *runtime.Scheme {
			return scheme
		},
	}
}

func (service *MockPlatformService) Create(ctx context.Context, obj runtime.Object, opts ...clientv1.CreateOption) error {
	return service.CreateFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) Delete(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error {
	return service.DeleteFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.GetFunc(ctx, key, obj)
}

func (service *MockPlatformService) List(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
	return service.ListFunc(ctx, list, opts...)
}

func (service *MockPlatformService) Update(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
	return service.UpdateFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) UpdateStatus(ctx context.Context, obj runtime.Object) error {
	return service.UpdateStatusFunc(ctx, obj)
}

func (service *MockPlatformService) Patch(ctx context.Context, obj runtime.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	return service.PatchFunc(ctx, obj, patch, opts...)
}

func (service *MockPlatformService) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteAllOfOption) error {
	return service.DeleteAllOfFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.GetCachedFunc(ctx, key, obj)
}

func (service *MockPlatformService) ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface {
	return service.ImageStreamTagsFunc(namespace)
}

func (service *MockPlatformService) GetScheme() *runtime.Scheme {
	return service.GetSchemeFunc()
}

func (service *MockPlatformService) IsMockService() bool {
	return true
}
