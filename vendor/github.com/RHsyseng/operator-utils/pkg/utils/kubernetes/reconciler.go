package kubernetes

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ExtendedReconciler struct {
	Service    PlatformService
	Reconciler reconcile.Reconciler
	Resource   client.Object
	Finalizers map[string]Finalizer
}

func NewExtendedReconciler(service PlatformService, reconciler reconcile.Reconciler, resource client.Object) ExtendedReconciler {
	return ExtendedReconciler{
		Service:    service,
		Reconciler: reconciler,
		Resource:   resource,
		Finalizers: map[string]Finalizer{},
	}
}

func (e *ExtendedReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := e.Resource.DeepCopyObject().(client.Object)
	err := e.Service.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	err = e.finalizeOnDelete(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return e.Reconciler.Reconcile(context.TODO(), request)
}
