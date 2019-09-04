package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"reflect"
	logs "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = logs.Log.WithName("comparator")

type MapComparator struct {
	Comparator ResourceComparator
}

func NewMapComparator() MapComparator {
	return MapComparator{
		Comparator: DefaultComparator(),
	}
}

func (this *MapComparator) Compare(deployed map[reflect.Type][]resource.KubernetesResource, requested map[reflect.Type][]resource.KubernetesResource) map[reflect.Type]ResourceDelta {
	delta := make(map[reflect.Type]ResourceDelta)
	for deployedType, deployedArray := range deployed {
		requestedArray := requested[deployedType]
		delta[deployedType] = this.compareArrays(deployedType, deployedArray, requestedArray)
	}
	for requestedType, requestedArray := range requested {
		if _, ok := deployed[requestedType]; !ok {
			//Item type in request does not exist in deployed set, needs to be added:
			delta[requestedType] = ResourceDelta{Added: requestedArray}
		}
	}
	return delta
}

func (this *MapComparator) compareArrays(resourceType reflect.Type, deployed []resource.KubernetesResource, requested []resource.KubernetesResource) ResourceDelta {
	deployedMap := getObjectMap(deployed)
	requestedMap := getObjectMap(requested)
	var added []resource.KubernetesResource
	var updated []resource.KubernetesResource
	var removed []resource.KubernetesResource
	for name, requestedObject := range requestedMap {
		deployedObject := deployedMap[name]
		if deployedObject == nil {
			added = append(added, requestedObject)
		} else if !this.Comparator.Compare(deployedObject, requestedObject) {
			updated = append(updated, requestedObject)
		}
	}
	for name, deployedObject := range deployedMap {
		if requestedMap[name] == nil {
			removed = append(removed, deployedObject)
		}
	}
	return ResourceDelta{
		Added:   added,
		Updated: updated,
		Removed: removed,
	}
}

func getObjectMap(objects []resource.KubernetesResource) map[string]resource.KubernetesResource {
	objectMap := make(map[string]resource.KubernetesResource)
	for index := range objects {
		objectMap[objects[index].GetName()] = objects[index]
	}
	return objectMap
}
