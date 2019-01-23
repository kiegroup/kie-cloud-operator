package v1

import (
	"context"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KieAppSpec defines the desired state of KieApp
type KieAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// KIE environment type to deploy (prod, authoring, trial, etc)
	Environment    string           `json:"environment,omitempty"`
	KieDeployments int              `json:"kieDeployments"` // Number of KieServer DeploymentConfigs (defaults to 1)
	RhpamRegistry  KieAppRegistry   `json:"rhpamRegistry,omitempty"`
	Objects        KieAppObjects    `json:"objects,omitempty"`
	CommonConfig   CommonConfig     `json:"commonConfig,omitempty"`
	Auth           KieAppAuthObject `json:"auth,omitempty"`
}

// KieAppRegistry defines the registry that should be used for rhpam images
type KieAppRegistry struct {
	Registry string `json:"registry,omitempty"` // Registry to use, can also be set w/ "REGISTRY" env variable
	Insecure bool   `json:"insecure"`           // Specify whether registry is insecure, can also be set w/ "INSECURE" env variable
}

// KieAppStatus defines the observed state of KieApp
type KieAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Status      string   `json:"status,omitempty"`
	ConsoleHost string   `json:"consoleHost,omitempty"`
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
	// Business Central container configs
	Console KieAppObject `json:"console,omitempty"`
	// KIE Server container configs
	Server KieAppObject `json:"server,omitempty"`
	// Smartrouter container configs
	Smartrouter KieAppObject `json:"smartrouter,omitempty"`
	// S2I Build configuration
	Builds []KieAppBuildObject `json:"builds,omitempty"`
}

type KieAppObject struct {
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources"`
}

type Environment struct {
	Console     CustomObject   `json:"console,omitempty"`
	Smartrouter CustomObject   `json:"smartrouter,omitempty"`
	Servers     []CustomObject `json:"servers,omitempty"`
	Others      []CustomObject `json:"others,omitempty"`
}

type CustomObject struct {
	Omit                   bool                           `json:"omit,omitempty"`
	PersistentVolumeClaims []corev1.PersistentVolumeClaim `json:"persistentVolumeClaims,omitempty"`
	ServiceAccounts        []corev1.ServiceAccount        `json:"serviceAccounts,omitempty"`
	Secrets                []corev1.Secret                `json:"secrets,omitempty"`
	Roles                  []rbacv1.Role                  `json:"roles,omitempty"`
	RoleBindings           []rbacv1.RoleBinding           `json:"roleBindings,omitempty"`
	DeploymentConfigs      []appsv1.DeploymentConfig      `json:"deploymentConfigs,omitempty"`
	BuildConfigs           []buildv1.BuildConfig          `json:"buildConfigs,omitempty"`
	ImageStreams           []oimagev1.ImageStream         `json:"imageStreams,omitempty"`
	Services               []corev1.Service               `json:"services,omitempty"`
	Routes                 []routev1.Route                `json:"routes,omitempty"`
}

// KieAppBuildObject Data to define how to build an application from source
type KieAppBuildObject struct {
	KieServerContainerDeployment string          `json:"kieServerContainerDeployment,omitempty"`
	GitSource                    GitSource       `json:"gitSource,omitempty"`
	Webhooks                     []WebhookSecret `json:"webhooks,omitempty"`
}

// GitSource Git coordinates to locate the source code to build
type GitSource struct {
	URI        string `json:"uri,omitempty"`
	Reference  string `json:"reference,omitempty"`
	ContextDir string `json:"contextDir,omitempty"`
}

// WebhookType literal type to distinguish between different types of Webhooks
type WebhookType string

const (
	// GitHubWebhook GitHub webhook
	GitHubWebhook WebhookType = "GitHub"
	// GenericWebhook Generic webhook
	GenericWebhook WebhookType = "Generic"
)

// WebhookSecret Secret to use for a given webhook
type WebhookSecret struct {
	Type   WebhookType `json:"type,omitempty"`
	Secret string      `json:"secret,omitempty"`
}

// KieAppAuthObject Authentication specification to be used by the KieApp
type KieAppAuthObject struct {
	SSO        *SSOAuthConfig        `json:"sso,omitempty"`
	LDAP       *LDAPAuthConfig       `json:"ldap,omitempty"`
	RoleMapper *RoleMapperAuthConfig `json:"roleMapper,omitempty"`
}

// SSOAuthConfig Authentication configuration for SSO
type SSOAuthConfig struct {
	URL                      string         `json:"url,omitempty"`
	Realm                    string         `json:"realm,omitempty"`
	AdminUser                string         `json:"adminUser,omitempty"`
	AdminPassword            string         `json:"adminPassword,omitempty"`
	DisableSSLCertValidation bool           `json:"disableSSLCertValication,omitempty"`
	PrincipalAttribute       string         `json:"principalAttribute,omitempty"`
	Clients                  SSOAuthClients `json:"clients,omitempty"`
}

// SSOAuthClients Different SSO Clients to use
type SSOAuthClients struct {
	Console SSOAuthClient   `json:"console,omitempty"`
	Servers []SSOAuthClient `json:"servers,omitempty"`
}

// SSOAuthClient Auth client to use for the SSO integration
type SSOAuthClient struct {
	Name          string `json:"name,omitempty"`
	Secret        string `json:"secret,omitempty"`
	HostnameHTTP  string `json:"hostnameHTTP,omitempty"`
	HostnameHTTPS string `json:"hostnameHTTPS,omitempty"`
}

// LDAPAuthConfig Authentication configuration for LDAP
type LDAPAuthConfig struct {
	URL                            string          `json:"url,omitempty"`
	BindDN                         string          `json:"bindDN,omitempty"`
	BindCredential                 string          `json:"bindCredential,omitempty"`
	JAASSecurityDomain             string          `json:"jaasSecurityDomain,omitempty"`
	BaseCtxDN                      string          `json:"baseCtxDN,omitempty"`
	BaseFilter                     string          `json:"baseFilter,omitempty"`
	SearchScope                    SearchScopeType `json:"searchScope,omitempty"`
	SearchTimeLimit                int32           `json:"searchTimeLimit,omitempty"`
	DistinguishedNameAttribute     string          `json:"distinguishedNameAttribute,omitempty"`
	ParseUsername                  bool            `json:"parseUsername,omitempty"`
	UsernameBeginString            string          `json:"usernameBeginString,omitempty"`
	UsernameEndString              string          `json:"usernameEndString,omitempty"`
	RoleAttributeID                string          `json:"roleAttributeID,omitempty"`
	RolesCtxDN                     string          `json:"rolesCtxDN,omitempty"`
	RoleFilter                     string          `json:"roleFilter,omitempty"`
	RoleRecursion                  int16           `json:"roleRecursion,omitempty"`
	DefaultRole                    string          `json:"defaultRole,omitempty"`
	RoleNameAttributeID            string          `json:"roleNameAttributeID,omitempty"`
	ParseRoleNameFromDN            bool            `json:"parseRoleNameFromDN,omitempty"`
	RoleAttributeIsDN              bool            `json:"roleAttributeIsDN,omitempty"`
	ReferralUserAttributeIDToCheck string          `json:"referralUserAttributeIDToCheck,omitempty"`
}

// SearchScopeType Type used to define how the LDAP searches are performed
type SearchScopeType string

const (
	// SubtreeSearchScope Subtree search scope
	SubtreeSearchScope SearchScopeType = "SUBTREE_SCOPE"
	// ObjectSearchScope Object search scope
	ObjectSearchScope SearchScopeType = "OBJECT_SCOPE"
	// OneLevelSearchScope One Level search scope
	OneLevelSearchScope SearchScopeType = "ONELEVEL_SCOPE"
)

// RoleMapperAuthConfig Configuration for RoleMapper Authentication
type RoleMapperAuthConfig struct {
	RolesProperties string `json:"rolesProperties,omitempty"`
	ReplaceRole     string `json:"replaceRole,omitempty"`
}

type OpenShiftObject interface {
	metav1.Object
	runtime.Object
}

type EnvTemplate struct {
	Template    `json:",inline"`
	ServerCount []Template `json:"serverCount,omitempty"`
}

// Template contains all the variables used in the yaml templates
type Template struct {
	*CommonConfig
	ApplicationName              string       `json:"applicationName,omitempty"`
	GitSource                    GitSource    `json:"gitSource,omitempty"`
	GitHubWebhookSecret          string       `json:"githubWebhookSecret,omitempty"`
	GenericWebhookSecret         string       `json:"genericWebhookSecret,omitempty"`
	KieServerContainerDeployment string       `json:"kieServerContainerDeployment,omitempty"`
	Auth                         AuthTemplate `json:"auth,omitempty"`
}

type CommonConfig struct {
	Version            string `json:"version,omitempty"`
	ImageTag           string `json:"imageTag,omitempty"`
	ConsoleName        string `json:"consoleName,omitempty"`
	ConsoleImage       string `json:"consoleImage,omitempty"`
	KeyStorePassword   string `json:"keyStorePassword,omitempty"`
	AdminPassword      string `json:"adminPassword,omitempty"`
	ControllerPassword string `json:"controllerPassword,omitempty"`
	ServerPassword     string `json:"serverPassword,omitempty"`
	MavenPassword      string `json:"mavenPassword,omitempty"`
}

// AuthTemplate Authentication definition used in the template
type AuthTemplate struct {
	SSO        SSOAuthConfig        `json:"sso,omitempty"`
	LDAP       LDAPAuthConfig       `json:"ldap,omitempty"`
	RoleMapper RoleMapperAuthConfig `json:"roleMapper,omitempty"`
}

type PlatformService interface {
	Create(ctx context.Context, obj runtime.Object) error
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	GetCached(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface
	GetScheme() *runtime.Scheme
	IsMockService() bool
}

func init() {
	SchemeBuilder.Register(&KieApp{}, &KieAppList{})
}
