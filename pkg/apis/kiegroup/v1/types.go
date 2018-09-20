package v1

import (
	apiv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []App `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AppSpec   `json:"spec"`
	Status            AppStatus `json:"status,omitempty"`
}

type AppSpec struct {
	Environment string           `json:"environment,omitempty"`
	Console     corev1.Container `json:"console,omitempty"`
	Server      corev1.Container `json:"server,omitempty"`
}

type AppStatus struct {
	// Fill me
}

type Environment struct {
	Console OpenShiftObject `json:"console"`
	Servers  []OpenShiftObject `json:"servers"`
}

type OpenShiftObject struct {
	DeploymentConfig apiv1.DeploymentConfig `json:"deployment"`
	Service          corev1.Service      `json:"service"`
	Route            routev1.Route       `json:"route"`
}
