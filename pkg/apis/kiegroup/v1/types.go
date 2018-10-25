package v1

import (
	appsv1 "github.com/openshift/api/apps/v1"
	authv1 "github.com/openshift/api/authorization/v1"
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
	Spec              AppSpec `json:"spec"`
	Status            string  `json:"status,omitempty"`
}

type AppSpec struct {
	// KIE environment type to deploy (prod, authoring, trial, etc)
	Environment string `json:"environment,omitempty"`
	// RHPAM version to deploy
	Version string `json:"version,omitempty"`
	// Image tag to use
	ImageTag string `json:"imageTag,omitempty"`
	// Number of KieServer DeploymentConfigs (defaults to 1)
	KieDeployments int        `json:"kieDeployments"`
	Objects        AppObjects `json:"objects,omitempty"`
}

type AppObjects struct {
	// KIE Server container configs
	Console AppObject `json:"console,omitempty"`
	// Business Central container configs
	Server AppObject `json:"server,omitempty"`
}

type AppObject struct {
	Env       []corev1.EnvVar             `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type Environment struct {
	Console CustomObject   `json:"console,omitempty"`
	Others  []CustomObject `json:"others,omitempty"`
	Servers []CustomObject `json:"servers,omitempty"`
}

type CustomObject struct {
	PersistentVolumeClaims []corev1.PersistentVolumeClaim `json:"persistentVolumeClaims,omitempty"`
	ServiceAccounts        []corev1.ServiceAccount        `json:"serviceAccounts,omitempty"`
	Secrets                []corev1.Secret                `json:"secrets,omitempty"`
	RoleBindings           []authv1.RoleBinding           `json:"roleBindings,omitempty"`
	DeploymentConfigs      []appsv1.DeploymentConfig      `json:"deploymentConfigs,omitempty"`
	Services               []corev1.Service               `json:"services,omitempty"`
	Routes                 []routev1.Route                `json:"routes,omitempty"`
}

type EnvTemplate struct {
	Template    `json:",inline"`
	ServerCount []Template `json:"serverCount,omitempty"`
}

type Template struct {
	ApplicationName  string `json:"applicationName,omitempty"`
	Version          string `json:"version,omitempty"`
	ImageTag         string `json:"imageTag,omitempty"`
	KeyStorePassword string `json:"keyStorePassword,omitempty"`
	AppPassword      string `json:"appPassword,omitempty"`
}
