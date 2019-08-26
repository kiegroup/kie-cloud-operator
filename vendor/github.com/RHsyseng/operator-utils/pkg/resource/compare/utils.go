package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"reflect"
)

type mapBuilder struct {
	resourceMap map[reflect.Type][]resource.KubernetesResource
}

func NewMapBuilder() *mapBuilder {
	this := &mapBuilder{resourceMap: make(map[reflect.Type][]resource.KubernetesResource)}
	return this
}

func (this *mapBuilder) Map() map[reflect.Type][]resource.KubernetesResource {
	return this.resourceMap
}

func (this *mapBuilder) SameTypeItems(items ...resource.KubernetesResource) *mapBuilder {
	if len(items) == 0 {
		return this
	}
	resourceType := reflect.ValueOf(items[0]).Elem().Type()
	this.resourceMap[resourceType] = append(this.resourceMap[resourceType], items...)
	return this
}

func (this *mapBuilder) DisparateTypeItems(items ...resource.KubernetesResource) *mapBuilder {
	for index := range items {
		this.Item(items[index])
	}
	return this
}

func (this *mapBuilder) Item(item resource.KubernetesResource) *mapBuilder {
	resourceType := reflect.ValueOf(item).Elem().Type()
	this.resourceMap[resourceType] = append(this.resourceMap[resourceType], item)
	return this
}
