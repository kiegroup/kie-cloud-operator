package kieapp

import (
	"context"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var name = "console-cr-form"
var operatorName string

func deployConsole(reconciler *Reconciler, operator *appsv1.Deployment) {
	log.Debugf("Will deploy operator-ui")
	namespace := os.Getenv(constants.NameSpaceEnv)
	operatorName = os.Getenv(constants.OpNameEnv)
	image := getImage(operator)
	pod := getPod(namespace, image)
	err := controllerutil.SetControllerReference(operator, pod, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for pod. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), pod)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create pod as it already exists", pod)
		} else {
			log.Error("Failed to create pod. ", err)
			return
		}
	}
	service := getService(namespace)
	err = controllerutil.SetControllerReference(operator, service, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for service. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), service)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create service as it already exists", service)
		} else {
			log.Error("Failed to create service. ", err)
			return
		}
	}
	route := getRoute(namespace)
	err = controllerutil.SetControllerReference(operator, route, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for route. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), route)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create route as it already exists", route)
		} else {
			log.Error("Failed to create route. ", err)
			return
		}
	}
}

func getPod(namespace string, image string) *corev1.Pod {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: image,
					Command: []string{
						"console-cr-form",
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: 8080},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						FailureThreshold:    20,
					},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: 8080},
							},
						},
						InitialDelaySeconds: 60,
						PeriodSeconds:       60,
					},
				},
			},
		},
	}
}

func getImage(operator *appsv1.Deployment) string {
	image := operator.Spec.Template.Spec.Containers[0].Image
	log.Debugf("Own image retrieved as %v", image)
	return image
}

func getService(namespace string) *corev1.Service {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: labels,
		},
	}
}

func getRoute(namespace string) *routev1.Route {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
		},
	}
}
