package v1

import (
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KieAppSpec defines the desired state of KieApp
type KieAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// KIE environment type to deploy (prod, authoring, trial, etc)
	Environment string `json:"environment,omitempty"`
	// Number of KieServer DeploymentConfigs (defaults to 1)
	KieDeployments int           `json:"kieDeployments"`
	Objects        KieAppObjects `json:"objects,omitempty"`
	Template       Template      `json:"template,omitempty"`
}

// KieAppStatus defines the observed state of KieApp
type KieAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Status      string   `json:"status,omitempty"`
	Deployments []string `json:"deployments,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KieApp is the Schema for the kieapps API
// +k8s:openapi-gen=true
type KieApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KieAppSpec   `json:"spec,omitempty"`
	Status KieAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KieAppList contains a list of KieApp
type KieAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KieApp `json:"items"`
}

type KieAppObjects struct {
	// KIE Server container configs
	Console KieAppObject `json:"console,omitempty"`
	// Business Central container configs
	Server KieAppObject `json:"server,omitempty"`
}

type KieAppObject struct {
	Env       []corev1.EnvVar             `json:"env"`
	Resources corev1.ResourceRequirements `json:"resources"`
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
	RoleBindings           []rbacv1.RoleBinding           `json:"roleBindings,omitempty"`
	DeploymentConfigs      []appsv1.DeploymentConfig      `json:"deploymentConfigs,omitempty"`
	Services               []corev1.Service               `json:"services,omitempty"`
	Routes                 []routev1.Route                `json:"routes,omitempty"`
}

type EnvTemplate struct {
	Template    `json:",inline"`
	ServerCount []Template `json:"serverCount,omitempty"`
}

type Template struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	Version            string `json:"version,omitempty"`
	ImageTag           string `json:"imageTag,omitempty"`
	KeyStorePassword   string `json:"keyStorePassword,omitempty"`
	AdminPassword      string `json:"adminPassword,omitempty"`
	ControllerPassword string `json:"controllerPassword,omitempty"`
	ServerPassword     string `json:"serverPassword,omitempty"`
	MavenPassword      string `json:"mavenPassword,omitempty"`
}

func init() {
	SchemeBuilder.Register(&KieApp{}, &KieAppList{})
}
