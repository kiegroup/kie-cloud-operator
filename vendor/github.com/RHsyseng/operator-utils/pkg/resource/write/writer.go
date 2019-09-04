package write

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write/hooks"
	newerror "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type UpdateHooks interface {
	Trigger(existing resource.KubernetesResource, requested resource.KubernetesResource) error
}

type resourceWriter struct {
	ownerRefs       []metav1.OwnerReference
	ownerController metav1.Object
	updateHooks     UpdateHooks
}

func New() *resourceWriter {
	return &resourceWriter{
		updateHooks: hooks.DefaultUpdateHooks(),
	}
}

func (this *resourceWriter) WithOwnerReferences(ownerRefs ...metav1.OwnerReference) *resourceWriter {
	this.ownerRefs = ownerRefs
	this.ownerController = nil
	return this
}

func (this *resourceWriter) WithOwnerController(ownerController metav1.Object) *resourceWriter {
	this.ownerController = ownerController
	this.ownerRefs = nil
	return this
}

func (this *resourceWriter) WithCustomUpdateHooks(updateHooks UpdateHooks) *resourceWriter {
	this.updateHooks = updateHooks
	return this
}

// AddResources sets ownership as/if configured, and then uses the writer to create them
// the boolean result is true if any changes were made
func (this *resourceWriter) AddResources(scheme *runtime.Scheme, writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
	var added bool
	for index := range resources {
		requested := resources[index]
		if this.ownerRefs != nil {
			requested.SetOwnerReferences(this.ownerRefs)
		} else if this.ownerController != nil {
			err := controllerutil.SetControllerReference(this.ownerController, requested, scheme)
			if err != nil {
				return added, err
			}
		}
		err := writer.Create(context.TODO(), requested)
		if err != nil {
			return added, err
		}
		added = true
	}
	return added, nil
}

// UpdateResources finds the updated counterpart for each of the provided resources in the existing array and uses it to set resource version and GVK
// It also sets ownership as/if configured, and then uses the writer to update them
// the boolean result is true if any changes were made
func (this *resourceWriter) UpdateResources(existing []resource.KubernetesResource, scheme *runtime.Scheme, writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
	var updated bool
	for index := range resources {
		requested := resources[index]
		var counterpart resource.KubernetesResource
		for _, candidate := range existing {
			if candidate.GetNamespace() == requested.GetNamespace() && candidate.GetName() == requested.GetName() {
				counterpart = candidate
				break
			}
		}
		if counterpart == nil {
			return updated, newerror.New("Failed to find a deployed counterpart to resource being updated")
		}
		err := this.updateHooks.Trigger(counterpart, requested)
		if err != nil {
			return updated, err
		}
		if this.ownerRefs != nil {
			requested.SetOwnerReferences(this.ownerRefs)
		} else if this.ownerController != nil {
			err := controllerutil.SetControllerReference(this.ownerController, requested, scheme)
			if err != nil {
				return updated, err
			}
		}
		err = writer.Update(context.TODO(), requested)
		if err != nil {
			return updated, err
		}
		updated = true
	}
	return updated, nil
}

// RemoveResources removes each of the provided resources using the provided writer
// the boolean result is true if any changes were made
func (this *resourceWriter) RemoveResources(writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
	var removed bool
	for index := range resources {
		err := writer.Delete(context.TODO(), resources[index])
		if err != nil {
			return removed, err
		}
		removed = true
	}
	return removed, nil
}
