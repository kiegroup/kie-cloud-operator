package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var logger = logs.GetLogger("comparator")

type MapComparator struct {
	Comparator ResourceComparator
}

func NewMapComparator() MapComparator {
	return MapComparator{
		Comparator: DefaultComparator(),
	}
}

func (this *MapComparator) Compare(deployed map[reflect.Type][]client.Object, requested map[reflect.Type][]client.Object) map[reflect.Type]ResourceDelta {
	delta := make(map[reflect.Type]ResourceDelta)
	for deployedType, deployedArray := range deployed {
		requestedArray := requested[deployedType]
		delta[deployedType] = this.Comparator.CompareArrays(deployedArray, requestedArray)
	}
	for requestedType, requestedArray := range requested {
		if _, ok := deployed[requestedType]; !ok {
			//Item type in request does not exist in deployed set, needs to be added:
			delta[requestedType] = ResourceDelta{Added: requestedArray}
		}
	}
	return delta
}
