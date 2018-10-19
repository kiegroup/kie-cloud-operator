package shared

import (
	"encoding/json"

	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	authv1 "github.com/openshift/api/authorization/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
)

func GetDeploymentTypeMeta() metav1.TypeMeta {
	kind := "DeploymentConfig"
	version := appsv1.GroupVersion.String()
	return getTypeMeta(kind, version)
}

func GetServiceTypeMeta() metav1.TypeMeta {
	kind := "Service"
	version := "v1"
	return getTypeMeta(kind, version)
}

func GetRouteTypeMeta() metav1.TypeMeta {
	kind := "Route"
	version := routev1.GroupVersion.String()
	return getTypeMeta(kind, version)
}

func getTypeMeta(kind string, version string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: version,
	}
}

func GetObjectMeta(service string, cr *opv1.App, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            service,
		Namespace:       cr.Namespace,
		OwnerReferences: GetOwnerReferences(cr),
		Labels:          labels,
	}
}

func GetOwnerReferences(cr *opv1.App) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(cr, opv1.SchemeGroupVersion.WithKind(cr.GetObjectKind().GroupVersionKind().Kind)),
	}
}

func GetServiceSpec(selector map[string]string, ports map[string]int) corev1.ServiceSpec {
	servicePorts := make([]corev1.ServicePort, len(ports))
	index := 0
	for key, value := range ports {
		servicePorts[index] = corev1.ServicePort{
			Name:       key,
			Port:       int32(value),
			Protocol:   "TCP",
			TargetPort: intstr.FromInt(value),
		}
		index++
	}
	return corev1.ServiceSpec{
		Ports:    servicePorts,
		Selector: selector,
	}
}

func GetRouteSpec(service string) routev1.RouteSpec {
	return routev1.RouteSpec{
		To: routev1.RouteTargetReference{
			Name: service,
		},
		Port: &routev1.RoutePort{
			TargetPort: intstr.FromString("http"),
		},
	}
}

func GetContainerPorts(ports map[string]int) []corev1.ContainerPort {
	containerPorts := make([]corev1.ContainerPort, len(ports))
	index := 0
	for key, value := range ports {
		containerPorts[index] = corev1.ContainerPort{
			Name:          key,
			ContainerPort: int32(value),
			Protocol:      "TCP",
		}
		index++
	}
	return containerPorts
}

func GetProbe(probeInts map[string]int, probeScript map[string]string) *corev1.Probe {
	curl := "curl --fail --silent -u '" + probeScript["username"] + ":" + probeScript["password"] + "' " + probeScript["url"]
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/bash", "-c", curl},
			},
		},
		InitialDelaySeconds: int32(probeInts["InitialDelaySeconds"]),
		TimeoutSeconds:      int32(probeInts["TimeoutSeconds"]),
		PeriodSeconds:       int32(probeInts["PeriodSeconds"]),
		FailureThreshold:    int32(probeInts["FailureThreshold"]),
	}
}

func GetResourceRequirements(resourceReqs map[string]map[corev1.ResourceName]string) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	if len(resourceReqs) > 0 {
		limits := resourceReqs["Limits"]
		if len(limits) > 0 {
			reqs.Limits = corev1.ResourceList{}
			for resourceName, value := range limits {
				reqs.Limits[resourceName] = resource.MustParse(value)
			}
		}
		requests := resourceReqs["Requests"]
		if len(requests) > 0 {
			reqs.Requests = corev1.ResourceList{}
			for resourceName, value := range requests {
				reqs.Requests[resourceName] = resource.MustParse(value)
			}
		}
	}
	return reqs
}

func GetDeploymentTrigger(containerName string, isNamespace string, isName string, isTag string) appsv1.DeploymentTriggerPolicies {
	return appsv1.DeploymentTriggerPolicies{
		{
			Type: appsv1.DeploymentTriggerOnImageChange,
			ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
				Automatic:      true,
				ContainerNames: []string{containerName},
				From: corev1.ObjectReference{
					Kind:      "ImageStreamTag",
					Namespace: isNamespace,
					Name:      isName + ":" + isTag,
				},
			},
		},
	}
}

func ObjectAppend(objs []runtime.Object, object opv1.CustomObject, cr *opv1.App) []runtime.Object {
	var o []runtime.Object
	for _, obj := range object.PersistentVolumeClaims {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.ServiceAccounts {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.Secrets {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.RoleBindings {
		obj.SetGroupVersionKind(authv1.SchemeGroupVersion.WithKind("RoleBinding"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.DeploymentConfigs {
		obj.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.Services {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		o = append(o, obj.DeepCopyObject())
	}
	for _, obj := range object.Routes {
		obj.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
		// check if tls route already exists
		if CheckTLS(obj.Spec.TLS) {
			if !CheckRouteOwnerRef(obj, cr) {
				o = append(o, obj.DeepCopyObject())
			}
		} else {
			o = append(o, obj.DeepCopyObject())
		}
	}

	return append(objs, o...)
}

func SetReferences(objs []runtime.Object, cr *opv1.App) []runtime.Object {
	for i, common := range objs {
		common.(metav1.Object).SetNamespace(cr.Namespace)
		common.(metav1.Object).SetOwnerReferences(GetOwnerReferences(cr))

		objs[i] = common
	}
	return objs
}

func CreateObject(resourceInterface dynamic.ResourceInterface, object runtime.Object) {
	unstructObj, err := k8sutil.UnstructuredFromRuntimeObject(object)
	if err != nil {
		logrus.Error(err)
	}
	_, err = resourceInterface.Create(unstructObj)
	if err != nil {
		logrus.Error(err)
	}
}

func GetObjectByte(resourceInterface dynamic.ResourceInterface, objName string) ([]byte, error) {
	unstructObj, err := resourceInterface.Get(objName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	objectByte, err := unstructObj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return objectByte, nil
}

func CheckRouteOwnerRef(route routev1.Route, cr *opv1.App) bool {
	groupVK := routev1.SchemeGroupVersion.WithKind("Route")
	apiVersion, kind := groupVK.ToAPIVersionAndKind()
	routeClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, cr.Namespace)
	if err != nil {
		logrus.Error(err)
	}
	objectByte, _ := GetObjectByte(routeClient, route.Name)
	json.Unmarshal(objectByte, &route)

	for _, r := range route.GetOwnerReferences() {
		if r.UID == cr.UID {
			return true
		}
	}
	return false
}

func CheckTLS(tls *routev1.TLSConfig) bool {
	if tls != nil {
		return true
	}
	return false
}
