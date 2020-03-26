package v2

import (
	"github.com/RHsyseng/operator-utils/pkg/olm"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KieAppSpec defines the desired state of KieApp
type KieAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// KIE environment type to deploy (prod, authoring, trial, etc)
	Environment   EnvironmentType  `json:"environment,omitempty"`
	ImageRegistry *KieAppRegistry  `json:"imageRegistry,omitempty"`
	Objects       KieAppObjects    `json:"objects,omitempty"`
	CommonConfig  CommonConfig     `json:"commonConfig,omitempty"`
	Auth          KieAppAuthObject `json:"auth,omitempty"`
	Upgrades      KieAppUpgrades   `json:"upgrades,omitempty"`
	Version       string           `json:"version,omitempty"`
	UseImageTags  bool             `json:"useImageTags"`
}

// EnvironmentType describes a possible application environment
type EnvironmentType string

const (
	// RhpamTrial RHPAM Trial environment
	RhpamTrial EnvironmentType = "rhpam-trial"
	// RhpamProduction RHPAM Production environment
	RhpamProduction EnvironmentType = "rhpam-production"
	// RhpamProductionImmutable RHPAM Production Immutable environment
	RhpamProductionImmutable EnvironmentType = "rhpam-production-immutable"
	// RhpamAuthoring RHPAM Authoring environment
	RhpamAuthoring EnvironmentType = "rhpam-authoring"
	// RhpamAuthoringHA RHPAM Authoring HA environment
	RhpamAuthoringHA EnvironmentType = "rhpam-authoring-ha"
	// RhdmTrial RHDM Trial environment
	RhdmTrial EnvironmentType = "rhdm-trial"
	// RhdmAuthoring RHDM Authoring environment
	RhdmAuthoring EnvironmentType = "rhdm-authoring"
	// RhdmAuthoringHA RHDM Authoring HA environment
	RhdmAuthoringHA EnvironmentType = "rhdm-authoring-ha"
	// RhdmProductionImmutable RHDM Production Immutable environment
	RhdmProductionImmutable EnvironmentType = "rhdm-production-immutable"
)

// EnvironmentConstants stores both the App and Replica Constants for a given environment
type EnvironmentConstants struct {
	App      AppConstants     `json:"app,omitempty"`
	Replica  ReplicaConstants `json:"replica,omitempty"`
	Database *DatabaseObject  `json:"database,omitempty"`
	Jms      *KieAppJmsObject `json:"jms,omitempty"`
}

// AppConstants data type to store application deployment constants
type AppConstants struct {
	Product      string `json:"name,omitempty"`
	Prefix       string `json:"prefix,omitempty"`
	ImageName    string `json:"imageName,omitempty"`
	ImageVar     string `json:"imageVar,omitempty"`
	MavenRepo    string `json:"mavenRepo,omitempty"`
	FriendlyName string `json:"friendlyName,omitempty"`
}

// KieAppRegistry defines the registry that should be used for rhpam images
type KieAppRegistry struct {
	Registry string `json:"registry,omitempty"` // Registry to use, can also be set w/ "REGISTRY" env variable
	Insecure bool   `json:"insecure"`           // Specify whether registry is insecure, can also be set w/ "INSECURE" env variable
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KieApp is the Schema for the kieapps API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kieapps,scope=Namespaced
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

// KieAppObjects KIE App deployment objects
type KieAppObjects struct {
	// Business Central container configs
	Console ConsoleObject `json:"console,omitempty"`
	// KIE Server configuration for individual sets
	Servers []KieServerSet `json:"servers,omitempty"`
	// SmartRouter container configs
	SmartRouter *SmartRouterObject `json:"smartRouter,omitempty"`
}

// KieAppUpgrades KIE App product upgrade flags
type KieAppUpgrades struct {
	Enabled bool `json:"enabled"`
	Minor   bool `json:"minor"`
}

// KieServerSet KIE Server configuration for a single set, or for multiple sets if deployments is set to >1
type KieServerSet struct {
	Deployments  *int                    `json:"deployments,omitempty"` // Number of KieServer DeploymentConfigs (defaults to 1)
	Name         string                  `json:"name,omitempty"`
	ID           string                  `json:"id,omitempty"`
	From         *corev1.ObjectReference `json:"from,omitempty"`
	Build        *KieAppBuildObject      `json:"build,omitempty"` // S2I Build configuration
	SSOClient    *SSOAuthClient          `json:"ssoClient,omitempty"`
	KieAppObject `json:",inline"`
	Database     *DatabaseObject  `json:"database,omitempty"`
	Jms          *KieAppJmsObject `json:"jms,omitempty"`
	Jvm          *JvmObject       `json:"jvm,omitempty"`
}

// ConsoleObject Console deployment object
type ConsoleObject struct {
	KieAppObject `json:",inline"`
	SSOClient    *SSOAuthClient  `json:"ssoClient,omitempty"`
	GitHooks     *GitHooksVolume `json:"gitHooks,omitempty"`
	Jvm          *JvmObject      `json:"jvm,omitempty"`
}

// SmartRouterObject deployment object
type SmartRouterObject struct {
	KieAppObject     `json:",inline"`
	Protocol         string `json:"protocol,omitempty"`
	UseExternalRoute bool   `json:"useExternalRoute,omitempty"`
}

// KieAppJmsObject messaging specification to be used by the KieApp
type KieAppJmsObject struct {
	EnableIntegration     bool   `json:"enableIntegration,omitempty"`
	Executor              *bool  `json:"executor,omitempty"`
	ExecutorTransacted    bool   `json:"executorTransacted,omitempty"`
	QueueRequest          string `json:"queueRequest,omitempty"`
	QueueResponse         string `json:"queueResponse,omitempty"`
	QueueExecutor         string `json:"queueExecutor,omitempty"`
	EnableSignal          bool   `json:"enableSignal,omitempty"`
	QueueSignal           string `json:"queueSignal,omitempty"`
	EnableAudit           bool   `json:"enableAudit,omitempty"`
	QueueAudit            string `json:"queueAudit,omitempty"`
	AuditTransacted       *bool  `json:"auditTransacted,omitempty"`
	Username              string `json:"username,omitempty"`
	Password              string `json:"password,omitempty"`
	AMQQueues             string `json:"amqQueues,omitempty"`     // It will receive the default value for the Executor, Request, Response, Signal and Audit queues.
	AMQSecretName         string `json:"amqSecretName,omitempty"` // AMQ SSL parameters
	AMQTruststoreName     string `json:"amqTruststoreName,omitempty"`
	AMQTruststorePassword string `json:"amqTruststorePassword,omitempty"`
	AMQKeystoreName       string `json:"amqKeystoreName,omitempty"`
	AMQKeystorePassword   string `json:"amqKeystorePassword,omitempty"`
	AMQEnableSSL          bool   `json:"amqEnableSSL,omitempty"` // flag will be set to true if all AMQ SSL parameters are correctly set.
}

// JvmObject JVM specification to be used by the KieApp
type JvmObject struct {
	JavaOptsAppend             string `json:"javaOptsAppend,omitempty"`
	JavaMaxMemRatio            *int32 `json:"javaMaxMemRatio,omitempty"`
	JavaInitialMemRatio        *int32 `json:"javaInitialMemRatio,omitempty"`
	JavaMaxInitialMem          *int32 `json:"javaMaxInitialMem,omitempty"`
	JavaDiagnostics            *bool  `json:"javaDiagnostics,omitempty"`
	JavaDebug                  *bool  `json:"javaDebug,omitempty"`
	JavaDebugPort              *int32 `json:"javaDebugPort,omitempty"`
	GcMinHeapFreeRatio         *int32 `json:"gcMinHeapFreeRatio,omitempty"`
	GcMaxHeapFreeRatio         *int32 `json:"gcMaxHeapFreeRatio,omitempty"`
	GcTimeRatio                *int32 `json:"gcTimeRatio,omitempty"`
	GcAdaptiveSizePolicyWeight *int32 `json:"gcAdaptiveSizePolicyWeight,omitempty"`
	GcMaxMetaspaceSize         *int32 `json:"gcMaxMetaspaceSize,omitempty"`
	GcContainerOptions         string `json:"gcContainerOptions,omitempty"`
}

// KieAppObject Generic object definition
type KieAppObject struct {
	Env            []corev1.EnvVar             `json:"env,omitempty"`
	Replicas       *int32                      `json:"replicas,omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources"`
	KeystoreSecret string                      `json:"keystoreSecret,omitempty"`
	Image          string                      `json:"image,omitempty"`
	ImageTag       string                      `json:"imageTag,omitempty"`
}

type Environment struct {
	Console     CustomObject   `json:"console,omitempty"`
	SmartRouter CustomObject   `json:"smartRouter,omitempty"`
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
	DeploymentConfigs      []oappsv1.DeploymentConfig     `json:"deploymentConfigs,omitempty"`
	StatefulSets           []appsv1.StatefulSet           `json:"statefulSets,omitempty"`
	BuildConfigs           []buildv1.BuildConfig          `json:"buildConfigs,omitempty"`
	ImageStreams           []oimagev1.ImageStream         `json:"imageStreams,omitempty"`
	Services               []corev1.Service               `json:"services,omitempty"`
	Routes                 []routev1.Route                `json:"routes,omitempty"`
}

// KieAppBuildObject Data to define how to build an application from source
type KieAppBuildObject struct {
	KieServerContainerDeployment     string                  `json:"kieServerContainerDeployment,omitempty"`
	GitSource                        GitSource               `json:"gitSource,omitempty"`
	MavenMirrorURL                   string                  `json:"mavenMirrorURL,omitempty"`
	ArtifactDir                      string                  `json:"artifactDir,omitempty"`
	Webhooks                         []WebhookSecret         `json:"webhooks,omitempty"`
	From                             *corev1.ObjectReference `json:"from,omitempty"`
	ExtensionImageStreamTag          string                  `json:"extensionImageStreamTag,omitempty"`
	ExtensionImageStreamTagNamespace string                  `json:"extensionImageStreamTagNamespace,omitempty"`
	ExtensionImageInstallDir         string                  `json:"extensionImageInstallDir,omitempty"`
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

// GitHooksVolume GitHooks volume configuration
type GitHooksVolume struct {
	MountPath string                  `json:"mountPath,omitempty"`
	From      *corev1.ObjectReference `json:"from,omitempty"`
}

// KieAppAuthObject Authentication specification to be used by the KieApp
type KieAppAuthObject struct {
	SSO        *SSOAuthConfig        `json:"sso,omitempty"`
	LDAP       *LDAPAuthConfig       `json:"ldap,omitempty"`
	RoleMapper *RoleMapperAuthConfig `json:"roleMapper,omitempty"`
}

// SSOAuthConfig Authentication configuration for SSO
type SSOAuthConfig struct {
	URL                      string `json:"url,omitempty"`
	Realm                    string `json:"realm,omitempty"`
	AdminUser                string `json:"adminUser,omitempty"`
	AdminPassword            string `json:"adminPassword,omitempty"`
	DisableSSLCertValidation bool   `json:"disableSSLCertValidation,omitempty"`
	PrincipalAttribute       string `json:"principalAttribute,omitempty"`
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
	RolesProperties string                  `json:"rolesProperties"`
	ReplaceRole     bool                    `json:"replaceRole,omitempty"`
	From            *corev1.ObjectReference `json:"from,omitempty"`
}

// DatabaseType to define what kind of database will be used for the Kie Servers
type DatabaseType string

const (
	// DatabaseH2 H2 Embedded Database deployment
	DatabaseH2 DatabaseType = "h2"
	// DatabaseMySQL MySQL Database deployment
	DatabaseMySQL DatabaseType = "mysql"
	// DatabasePostgreSQL PostgreSQL Database deployment
	DatabasePostgreSQL DatabaseType = "postgresql"
	// DatabaseExternal External Database
	DatabaseExternal DatabaseType = "external"
)

// DatabaseObject Defines how a KieServer will manage and create a new Database
// or connect to an existing one
type DatabaseObject struct {
	Type           DatabaseType            `json:"type,omitempty"`
	Size           string                  `json:"size,omitempty"`
	ExternalConfig *ExternalDatabaseObject `json:"externalConfig,omitempty"`
}

// ExternalDatabaseObject configuration definition of an external database
type ExternalDatabaseObject struct {
	Driver                     string `json:"driver,omitempty"`
	Dialect                    string `json:"dialect,omitempty"`
	Name                       string `json:"name,omitempty"`
	Host                       string `json:"host,omitempty"`
	Port                       string `json:"port,omitempty"`
	JdbcURL                    string `json:"jdbcURL,omitempty"`
	NonXA                      string `json:"nonXA,omitempty"`
	Username                   string `json:"username,omitempty"`
	Password                   string `json:"password,omitempty"`
	MinPoolSize                string `json:"minPoolSize,omitempty"`
	MaxPoolSize                string `json:"maxPoolSize,omitempty"`
	ConnectionChecker          string `json:"connectionChecker,omitempty"`
	ExceptionSorter            string `json:"exceptionSorter,omitempty"`
	BackgroundValidation       string `json:"backgroundValidation,omitempty"`
	BackgroundValidationMillis string `json:"backgroundValidationMillis,omitempty"`
}

type OpenShiftObject interface {
	metav1.Object
	runtime.Object
}

type EnvTemplate struct {
	*CommonConfig `json:",inline"`
	Console       ConsoleTemplate     `json:"console,omitempty"`
	Servers       []ServerTemplate    `json:"servers,omitempty"`
	SmartRouter   SmartRouterTemplate `json:"smartRouter,omitempty"`
	Auth          AuthTemplate        `json:"auth,omitempty"`
	Constants     TemplateConstants   `json:"constants,omitempty"`
}

// TemplateConstants constant values that are used within the different configuration templates
type TemplateConstants struct {
	Product              string `json:"product,omitempty"`
	Major                string `json:"major,omitempty"`
	Minor                string `json:"minor,omitempty"`
	Micro                string `json:"micro,omitempty"`
	MavenRepo            string `json:"mavenRepo,omitempty"`
	KeystoreVolumeSuffix string `json:"keystoreVolumeSuffix"`
	DatabaseVolumeSuffix string `json:"databaseVolumeSuffix"`
	OseCliImageURL       string `json:"oseCliImageURL,omitempty"`
	BrokerImage          string `json:"brokerImage"`
	BrokerImageTag       string `json:"brokerImageTag"`
	DatagridImage        string `json:"datagridImage"`
	DatagridImageTag     string `json:"datagridImageTag"`
	MySQLImageURL        string `json:"mySQLImageURL"`
	PostgreSQLImageURL   string `json:"postgreSQLImageURL"`
	BrokerImageURL       string `json:"brokerImageURL,omitempty"`
	DatagridImageURL     string `json:"datagridImageURL,omitempty"`
	RoleMapperVolume     string `json:"roleMapperVolume"`
	GitHooksVolume       string `json:"gitHooksVolume,omitempty"`
}

// ConsoleTemplate contains all the variables used in the yaml templates
type ConsoleTemplate struct {
	OmitImageStream bool           `json:"omitImageStream"`
	SSOAuthClient   SSOAuthClient  `json:"ssoAuthClient,omitempty"`
	Name            string         `json:"name,omitempty"`
	Replicas        int32          `json:"replicas,omitempty"`
	Image           string         `json:"image,omitempty"`
	ImageTag        string         `json:"imageTag,omitempty"`
	ImageURL        string         `json:"imageURL,omitempty"`
	KeystoreSecret  string         `json:"keystoreSecret,omitempty"`
	GitHooks        GitHooksVolume `json:"gitHooks,omitempty"`
	Jvm             JvmObject      `json:"jvm,omitempty"`
}

// ServerTemplate contains all the variables used in the yaml templates
type ServerTemplate struct {
	OmitImageStream bool                   `json:"omitImageStream"`
	KieName         string                 `json:"kieName,omitempty"`
	KieServerID     string                 `json:"kieServerID,omitempty"`
	Replicas        int32                  `json:"replicas,omitempty"`
	SSOAuthClient   SSOAuthClient          `json:"ssoAuthClient,omitempty"`
	From            corev1.ObjectReference `json:"from,omitempty"`
	ImageURL        string                 `json:"imageURL,omitempty"`
	Build           BuildTemplate          `json:"build,omitempty"`
	KeystoreSecret  string                 `json:"keystoreSecret,omitempty"`
	Database        DatabaseObject         `json:"database,omitempty"`
	Jms             KieAppJmsObject        `json:"jms,omitempty"`
	SmartRouter     SmartRouterObject      `json:"smartRouter,omitempty"`
	Jvm             JvmObject              `json:"jvm,omitempty"`
}

// SmartRouterTemplate contains all the variables used in the yaml templates
type SmartRouterTemplate struct {
	OmitImageStream  bool   `json:"omitImageStream"`
	Replicas         int32  `json:"replicas,omitempty"`
	KeystoreSecret   string `json:"keystoreSecret,omitempty"`
	Protocol         string `json:"protocol,omitempty"`
	UseExternalRoute bool   `json:"useExternalRoute,omitempty"`
	Image            string `json:"image,omitempty"`
	ImageTag         string `json:"imageTag,omitempty"`
	ImageURL         string `json:"imageURL,omitempty"`
}

// ReplicaConstants contains the default replica amounts for a component in a given environment type
type ReplicaConstants struct {
	Console     Replicas `json:"console,omitempty"`
	Server      Replicas `json:"server,omitempty"`
	SmartRouter Replicas `json:"smartRouter,omitempty"`
}

// Replicas contains replica settings
type Replicas struct {
	Replicas  int32 `json:"replicas,omitempty"`
	DenyScale bool  `json:"denyScale,omitempty"`
}

// BuildTemplate build variables used in the templates
type BuildTemplate struct {
	From                         corev1.ObjectReference `json:"from,omitempty"`
	GitSource                    GitSource              `json:"gitSource,omitempty"`
	GitHubWebhookSecret          string                 `json:"githubWebhookSecret,omitempty"`
	GenericWebhookSecret         string                 `json:"genericWebhookSecret,omitempty"`
	KieServerContainerDeployment string                 `json:"kieServerContainerDeployment,omitempty"`
	MavenMirrorURL               string                 `json:"mavenMirrorURL,omitempty"`
	ArtifactDir                  string                 `json:"artifactDir,omitempty"`
	// Extension image configuration which provides custom jdbc drivers to be used
	// by KieServer.
	ExtensionImageStreamTag          string `json:"extensionImageStreamTag,omitempty"`
	ExtensionImageStreamTagNamespace string `json:"extensionImageStreamTagNamespace,omitempty"`
	ExtensionImageInstallDir         string `json:"extensionImageInstallDir,omitempty"`
}

// CommonConfig variables used in the templates
type CommonConfig struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	KeyStorePassword   string `json:"keyStorePassword,omitempty"`
	AdminUser          string `json:"adminUser,omitempty"`
	AdminPassword      string `json:"adminPassword,omitempty"`
	DBPassword         string `json:"dbPassword,omitempty"`
	AMQPassword        string `json:"amqPassword,omitempty"`
	AMQClusterPassword string `json:"amqClusterPassword,omitempty"`
	// Deprecated - Remove when 7.7.0 is the oldest version supported
	ControllerPassword string `json:"controllerPassword,omitempty"`
	// Deprecated - Remove when 7.7.0 is the oldest version supported
	ServerPassword string `json:"serverPassword,omitempty"`
	// Deprecated - Remove when 7.7.0 is the oldest version supported
	MavenPassword string `json:"mavenPassword,omitempty"`
}

// VersionConfigs ...
type VersionConfigs struct {
	APIVersion          string `json:"apiVersion,omitempty"`
	OseCliImageURL      string `json:"oseCliImageURL,omitempty"`
	OseCliComponent     string `json:"oseCliComponent,omitempty"`
	BrokerImage         string `json:"brokerImage,omitempty"`
	BrokerImageTag      string `json:"brokerImageTag,omitempty"`
	BrokerImageURL      string `json:"brokerImageURL,omitempty"`
	BrokerComponent     string `json:"brokerComponent,omitempty"`
	DatagridImage       string `json:"datagridImage,omitempty"`
	DatagridImageTag    string `json:"datagridImageTag,omitempty"`
	DatagridImageURL    string `json:"datagridImageURL,omitempty"`
	DatagridComponent   string `json:"datagridComponent,omitempty"`
	MySQLImageURL       string `json:"mySQLImageURL,omitempty"`
	MySQLComponent      string `json:"mySQLComponent,omitempty"`
	PostgreSQLImageURL  string `json:"postgreSQLImageURL,omitempty"`
	PostgreSQLComponent string `json:"postgreSQLComponent,omitempty"`
}

// AuthTemplate Authentication definition used in the template
type AuthTemplate struct {
	SSO        SSOAuthConfig      `json:"sso,omitempty"`
	LDAP       LDAPAuthConfig     `json:"ldap,omitempty"`
	RoleMapper RoleMapperTemplate `json:"roleMapper,omitempty"`
}

// RoleMapperTemplate RoleMapper definition used in the template
type RoleMapperTemplate struct {
	MountPath            string `json:"mountPath,omitempty"`
	RoleMapperAuthConfig `json:",inline"`
}

// ConditionType - type of condition
type ConditionType string

const (
	// DeployedConditionType - the kieapp is deployed
	DeployedConditionType ConditionType = "Deployed"
	// ProvisioningConditionType - the kieapp is being provisioned
	ProvisioningConditionType ConditionType = "Provisioning"
	// FailedConditionType - the kieapp is in a failed state
	FailedConditionType ConditionType = "Failed"
)

// ReasonType - type of reason
type ReasonType string

const (
	// DeploymentFailedReason - Unable to deploy the application
	DeploymentFailedReason ReasonType = "DeploymentFailed"
	// ConfigurationErrorReason - An invalid configuration caused an error
	ConfigurationErrorReason ReasonType = "ConfigurationError"
	// MissingDependenciesReason - Dependencies does not exist or cannot be found
	MissingDependenciesReason ReasonType = "MissingDependencies"
	// UnknownReason - Unable to determine the error
	UnknownReason ReasonType = "Unknown"
)

// Condition - The condition for the kie-cloud-operator
type Condition struct {
	Type               ConditionType          `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ReasonType             `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// KieAppStatus - The status for custom resources managed by the operator-sdk.
type KieAppStatus struct {
	Conditions  []Condition          `json:"conditions"`
	ConsoleHost string               `json:"consoleHost,omitempty"`
	Deployments olm.DeploymentStatus `json:"deployments"`
	Phase       ConditionType        `json:"phase,omitempty"`
}

func init() {
	SchemeBuilder.Register(&KieApp{}, &KieAppList{})
}
