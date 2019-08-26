package test

import (
	"fmt"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRoutes(count int) []routev1.Route {
	var slice []routev1.Route
	for i := 0; i < count; i++ {
		rte := routev1.Route{
			TypeMeta: metav1.TypeMeta{},
			Spec:     routev1.RouteSpec{},
			Status:   routev1.RouteStatus{},
		}
		rte.Name = fmt.Sprintf("%s%d", "rte", (i + 1))
		slice = append(slice, rte)
	}
	return slice
}

func GetServices(count int) []corev1.Service {
	var slice []corev1.Service
	for i := 0; i < count; i++ {
		svc := corev1.Service{
			TypeMeta: metav1.TypeMeta{},
			Spec:     corev1.ServiceSpec{},
			Status:   corev1.ServiceStatus{},
		}
		svc.Name = fmt.Sprintf("%s%d", "svc", (i + 1))
		slice = append(slice, svc)
	}
	return slice
}

func GetDeploymentConfigs(count int) []oappsv1.DeploymentConfig {
	var slice []oappsv1.DeploymentConfig
	for i := 0; i < count; i++ {
		dc := oappsv1.DeploymentConfig{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       oappsv1.DeploymentConfigSpec{},
			Status: oappsv1.DeploymentConfigStatus{
				ReadyReplicas: 0,
			},
		}
		dc.Name = fmt.Sprintf("%s%d", "dc", (i + 1))
		slice = append(slice, dc)
	}
	return slice
}
