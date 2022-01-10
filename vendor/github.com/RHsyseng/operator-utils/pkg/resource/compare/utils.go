package compare

import (
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mapBuilder struct {
	resourceMap map[reflect.Type][]client.Object
}

func NewMapBuilder() *mapBuilder {
	this := &mapBuilder{resourceMap: make(map[reflect.Type][]client.Object)}
	return this
}

func (this *mapBuilder) ResourceMap() map[reflect.Type][]client.Object {
	return this.resourceMap
}

func (this *mapBuilder) Add(resources ...client.Object) *mapBuilder {
	for index := range resources {
		if resources[index] == nil || reflect.ValueOf(resources[index]).IsNil() {
			continue
		}
		resourceType := reflect.ValueOf(resources[index]).Elem().Type()
		this.resourceMap[resourceType] = append(this.resourceMap[resourceType], resources[index])
	}
	return this
}
