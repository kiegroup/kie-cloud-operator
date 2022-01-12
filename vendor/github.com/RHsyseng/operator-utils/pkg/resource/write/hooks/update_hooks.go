package hooks

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type hookFunc = func(existing client.Object, requested client.Object) error

type UpdateHookMap struct {
	DefaultHook hookFunc
	HookMap     map[reflect.Type]hookFunc
}

func DefaultUpdateHooks() *UpdateHookMap {
	hookMap := make(map[reflect.Type]func(existing client.Object, requested client.Object) error)
	hookMap[reflect.TypeOf(corev1.Service{})] = serviceHook
	return &UpdateHookMap{
		DefaultHook: defaultHook,
		HookMap:     hookMap,
	}
}

func (this *UpdateHookMap) Trigger(existing client.Object, requested client.Object) error {
	function := this.HookMap[reflect.ValueOf(existing).Elem().Type()]
	if function == nil {
		function = this.DefaultHook
	}
	return function(existing, requested)
}

func defaultHook(existing client.Object, requested client.Object) error {
	requested.SetResourceVersion(existing.GetResourceVersion())
	requested.GetObjectKind().SetGroupVersionKind(existing.GetObjectKind().GroupVersionKind())
	return nil
}

func serviceHook(existing client.Object, requested client.Object) error {
	existingService := existing.(*corev1.Service)
	requestedService := requested.(*corev1.Service)
	if requestedService.Spec.ClusterIP == "" {
		requestedService.Spec.ClusterIP = existingService.Spec.ClusterIP
	}
	err := defaultHook(existing, requested)
	if err != nil {
		return err
	}
	return nil
}
