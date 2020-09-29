package kieapp

import (
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add Creates a new controller and starts watching resources
func Add(mgr manager.Manager, reconciler reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kieapp-controller", mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	watchObjects := []runtime.Object{
		// Watch for changes to primary resource KieApp
		&api.KieApp{},
		&appsv1.Deployment{},
	}
	objectHandler := &handler.EnqueueRequestForObject{}
	for _, watchObject := range watchObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, objectHandler)
		if err != nil {
			return err
		}
	}

	watchOwnedObjects := []runtime.Object{
		&corev1.ConfigMap{},
		&corev1.Pod{},
		&rbacv1.RoleBinding{},
		&rbacv1.Role{},
		&corev1.Service{},
		&routev1.Route{},
		&corev1.ServiceAccount{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		OwnerType: &operatorsv1alpha1.ClusterServiceVersion{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	watchOwnedObjects = []runtime.Object{
		&oappsv1.DeploymentConfig{},
		&appsv1.StatefulSet{},
		&corev1.PersistentVolumeClaim{},
		&rbacv1.RoleBinding{},
		&rbacv1.Role{},
		&corev1.ServiceAccount{},
		&corev1.Secret{},
		&corev1.Service{},
		&routev1.Route{},
		&buildv1.BuildConfig{},
		&oimagev1.ImageStream{},
	}
	ownerHandler = &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.KieApp{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	watchOwnedObjects = []runtime.Object{
		&corev1.ConfigMap{},
	}
	ownerHandler = &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &oappsv1.DeploymentConfig{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	return nil
}
