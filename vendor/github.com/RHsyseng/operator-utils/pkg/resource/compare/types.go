package compare

import (
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceDelta struct {
	Added   []client.Object
	Updated []client.Object
	Removed []client.Object
}

func (delta *ResourceDelta) HasChanges() bool {
	if len(delta.Added) > 0 {
		return true
	}
	if len(delta.Updated) > 0 {
		return true
	}
	if len(delta.Removed) > 0 {
		return true
	}
	return false
}

type ResourceComparator interface {
	SetDefaultComparator(compFunc func(deployed client.Object, requested client.Object) bool)
	GetDefaultComparator() func(deployed client.Object, requested client.Object) bool
	SetComparator(resourceType reflect.Type, compFunc func(deployed client.Object, requested client.Object) bool)
	GetComparator(resourceType reflect.Type) func(deployed client.Object, requested client.Object) bool
	Compare(deployed client.Object, requested client.Object) bool
	CompareArrays(deployed []client.Object, requested []client.Object) ResourceDelta
}

func DefaultComparator() ResourceComparator {
	return &resourceComparator{
		deepEquals,
		defaultMap(),
	}
}

func SimpleComparator() ResourceComparator {
	return &resourceComparator{
		deepEquals,
		make(map[reflect.Type]func(client.Object, client.Object) bool),
	}
}
