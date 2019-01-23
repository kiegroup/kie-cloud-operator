package controller

import (
	"context"

	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type KubernetesPlatformService struct {
	client      clientv1.Client
	cache       cachev1.Cache
	imageClient *imagev1.ImageV1Client
	scheme      *runtime.Scheme
}

func GetInstance(mgr manager.Manager) KubernetesPlatformService {
	imageClient, err := imagev1.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("Error getting image client: %v", err)
		return KubernetesPlatformService{}
	}

	return KubernetesPlatformService{
		client:      mgr.GetClient(),
		cache:       mgr.GetCache(),
		imageClient: imageClient,
		scheme:      mgr.GetScheme(),
	}
}

func (service *KubernetesPlatformService) Create(ctx context.Context, obj runtime.Object) error {
	return service.client.Create(ctx, obj)
}

func (service *KubernetesPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.client.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) List(ctx context.Context, opts *clientv1.ListOptions, list runtime.Object) error {
	return service.client.List(ctx, opts, list)
}

func (service *KubernetesPlatformService) Update(ctx context.Context, obj runtime.Object) error {
	return service.client.Update(ctx, obj)
}

func (service *KubernetesPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.cache.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface {
	return service.imageClient.ImageStreamTags(namespace)
}

func (service *KubernetesPlatformService) GetScheme() *runtime.Scheme {
	return service.scheme
}

func (service *KubernetesPlatformService) IsMockService() bool {
	return false
}
