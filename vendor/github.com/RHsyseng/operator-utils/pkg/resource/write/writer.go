package write

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	newerror "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AddResources sets ownership for provided resources to the provided owner, and then uses the writer to create them
// the boolean result is true if any changes were made
func AddResources(owner resource.KubernetesResource, scheme *runtime.Scheme, writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
	var added bool
	for index := range resources {
		requested := resources[index]
		err := controllerutil.SetControllerReference(owner, requested, scheme)
		if err != nil {
			return added, err
		}
		err = writer.Create(context.TODO(), requested)
		if err != nil {
			return added, err
		}
		added = true
	}
	return added, nil
}

// UpdateResources finds the updated counterpart for each of the provided resources in the existing array and uses it to set resource version and GVK
// It also sets ownership to the provided owner, and then uses the writer to update them
// the boolean result is true if any changes were made
func UpdateResources(owner resource.KubernetesResource, existing []resource.KubernetesResource, scheme *runtime.Scheme, writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
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
		requested.SetResourceVersion(counterpart.GetResourceVersion())
		requested.GetObjectKind().SetGroupVersionKind(counterpart.GetObjectKind().GroupVersionKind())
		err := controllerutil.SetControllerReference(owner, requested, scheme)
		if err != nil {
			return updated, err
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
func RemoveResources(writer clientv1.Writer, resources []resource.KubernetesResource) (bool, error) {
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
