package kieapp

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
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
		&corev1.ConfigMap{},
		// Watch for changes to primary resource KieApp
		&v1.KieApp{},
	}
	objectHandler := &handler.EnqueueRequestForObject{}
	for _, watchObject := range watchObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, objectHandler)
		if err != nil {
			return err
		}
	}

	watchOwnedObjects := []runtime.Object{
		&oappsv1.DeploymentConfig{},
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
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}
	return nil
}
