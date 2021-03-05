package v2

import (
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KieAppSpec defines the desired state of KieApp
type KieAppSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rhdm-authoring-ha;rhdm-authoring;rhdm-production-immutable;rhdm-trial;rhpam-authoring-ha;rhpam-authoring;rhpam-production-immutable;rhpam-production;rhpam-standalone-dashbuilder;rhpam-trial
	// The name of the environment used as a baseline
	Environment EnvironmentType `json:"environment"`
	// If required imagestreams are missing in both the 'openshift' and local namespaces, the operator will create said imagestreams locally using the registry specified here.
	ImageRegistry *KieAppRegistry `json:"imageRegistry,omitempty"`
	// Configuration of the RHPAM components
	Objects KieAppObjects `json:"objects,omitempty"`
	// Specify the level of product upgrade that should be allowed when an older product version is detected
	Upgrades KieAppUpgrades `json:"upgrades,omitempty"`
	// Set true to enable image tags, disabled by default. This will leverage image tags instead of the image digests.
	UseImageTags bool `json:"useImageTags,omitempty"`
	// The version of the application deployment.
	Version      string            `json:"version,omitempty"`
	CommonConfig CommonConfig      `json:"commonConfig,omitempty"`
	Auth         *KieAppAuthObject `json:"auth,omitempty"`
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
	// RhpamDashbuilder RHPAM Standalone Dashbuilder environment
	RhpamStandaloneDashbuilder EnvironmentType = "rhpam-standalone-dashbuilder"
	// RhdmTrial RHDM Trial environment
	RhdmTrial EnvironmentType = "rhdm-trial"
	// RhdmAuthoring RHDM Authoring environment
	RhdmAuthoring EnvironmentType = "rhdm-authoring"
	// RhdmAuthoringHA RHDM Authoring HA environment
	RhdmAuthoringHA EnvironmentType = "rhdm-authoring-ha"
	// RhdmProductionImmutable RHDM Production Immutable environment
	RhdmProductionImmutable EnvironmentType = "rhdm-production-immutable"
)

// KieAppRegistry defines the registry that should be used for rhpam images
type KieAppRegistry struct {
	// Image registry's base 'url:port'. e.g. registry.example.com:5000. Defaults to 'registry.redhat.io'.
	Registry string `json:"registry,omitempty"`
	// A flag used to indicate the specified registry is insecure. Defaults to 'false'.
	Insecure bool `json:"insecure,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KieApp is the Schema for the kieapps API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kieapps,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`,description="The version of the application deployment"
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment`,description="The name of the environment used as a baseline"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`,description="The status of the KieApp deployment"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type KieApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec   KieAppSpec   `json:"spec"`
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
	Console *ConsoleObject `json:"console,omitempty"`
	// Configuration of the each individual KIE server
	Servers          []KieServerSet          `json:"servers,omitempty"`
	SmartRouter      *SmartRouterObject      `json:"smartRouter,omitempty"`
	ProcessMigration *ProcessMigrationObject `json:"processMigration,omitempty"`
	Dashbuilder      *DashbuilderObject      `json:"dashbuilder,omitempty"`
}

// KieAppUpgrades KIE App product upgrade flags
type KieAppUpgrades struct {
	// Set true to enable automatic micro version product upgrades, it is disabled by default.
	Enabled bool `json:"enabled,omitempty"`
	// Set true to enable automatic minor product version upgrades, it is disabled by default. Requires spec.upgrades.enabled to be true.
	Minor bool `json:"minor,omitempty"`
}

// KieServerSet KIE Server configuration for a single set, or for multiple sets if deployments is set to >1
type KieServerSet struct {
	// +kubebuilder:validation:Format:=int
	// Number of Server sets that will be deployed
	Deployments *int `json:"deployments,omitempty"` // Number of KieServer DeploymentConfigs (defaults to 1)
	// Server name
	Name string `json:"name,omitempty"`
	// Server ID
	ID           string             `json:"id,omitempty"`
	From         *ImageObjRef       `json:"from,omitempty"`
	Build        *KieAppBuildObject `json:"build,omitempty"` // S2I Build configuration
	SSOClient    *SSOAuthClient     `json:"ssoClient,omitempty"`
	KieAppObject `json:",inline"`
	Database     *DatabaseObject  `json:"database,omitempty"`
	Jms          *KieAppJmsObject `json:"jms,omitempty"`
	Jvm          *JvmObject       `json:"jvm,omitempty"`
}

// ConsoleObject configuration of the RHPAM workbench
type ConsoleObject struct {
	KieAppObject `json:",inline"`
	SSOClient    *SSOAuthClient  `json:"ssoClient,omitempty"`
	GitHooks     *GitHooksVolume `json:"gitHooks,omitempty"`
	Jvm          *JvmObject      `json:"jvm,omitempty"`
	PvSize       string          `json:"pvSize,omitempty"`
}

// DashbuilderObject configuration of the RHPAM Dashbuilder
type DashbuilderObject struct {
	KieAppObject `json:",inline"`
	SSOClient    *SSOAuthClient     `json:"ssoClient,omitempty"`
	Jvm          *JvmObject         `json:"jvm,omitempty"`
	Config       *DashbuilderConfig `json:"config,omitempty"`
}

// SmartRouterObject configuration of the RHPAM smart router
type SmartRouterObject struct {
	KieAppObject `json:",inline"`
	// +kubebuilder:validation:Enum:=http;https
	// Smart Router protocol, if no value is provided, http is the default protocol.
	Protocol string `json:"protocol,omitempty"`
	// If enabled, Business Central will use the external smartrouter route to communicate with it. Note that, valid SSL certificates should be used.
	UseExternalRoute bool `json:"useExternalRoute,omitempty"`
}

// KieAppJmsObject messaging specification to be used by the KieApp
type KieAppJmsObject struct {
	// +kubebuilder:validation:Required
	// When set to true will configure the KIE Server with JMS integration, if no configuration is added, the default will be used.
	EnableIntegration bool `json:"enableIntegration"`
	// Set false to disable the JMS executor, it is enabled by default.
	Executor *bool `json:"executor,omitempty"`
	// Enable transactions for JMS executor, disabled by default.
	ExecutorTransacted bool `json:"executorTransacted,omitempty"`
	// JNDI name of request queue for JMS, example queue/CUSTOM.KIE.SERVER.REQUEST, default is queue/KIE.SERVER.REQUEST.
	QueueRequest string `json:"queueRequest,omitempty"`
	// JNDI name of response queue for JMS, example queue/CUSTOM.KIE.SERVER.RESPONSE, default is queue/KIE.SERVER.RESPONSE.
	QueueResponse string `json:"queueResponse,omitempty"`
	// JNDI name of executor queue for JMS, example queue/CUSTOM.KIE.SERVER.EXECUTOR, default is queue/KIE.SERVER.EXECUTOR.
	QueueExecutor string `json:"queueExecutor,omitempty"`
	// Enable the Signal configuration through JMS. Default is false.
	EnableSignal bool `json:"enableSignal,omitempty"`
	// JNDI name of signal queue for JMS, example queue/CUSTOM.KIE.SERVER.SIGNAL, default is queue/KIE.SERVER.SIGNAL.
	QueueSignal string `json:"queueSignal,omitempty"`
	// Enable the Audit logging through JMS. Default is false.
	EnableAudit bool `json:"enableAudit,omitempty"`
	// JNDI name of audit logging queue for JMS, example queue/CUSTOM.KIE.SERVER.AUDIT, default is queue/KIE.SERVER.AUDIT.
	QueueAudit string `json:"queueAudit,omitempty"`
	// Determines if JMS session is transacted or not - default true.
	AuditTransacted *bool `json:"auditTransacted,omitempty"`
	// AMQ broker username to connect do the AMQ, generated if empty.
	Username string `json:"username,omitempty"`
	// +kubebuilder:validation:Format:=password
	// AMQ broker password to connect do the AMQ, generated if empty.
	Password string `json:"password,omitempty"`
	// AMQ broker broker comma separated queues, if empty the values from default queues will be used.
	AMQQueues string `json:"amqQueues,omitempty"` // It will receive the default value for the Executor, Request, Response, Signal and Audit queues.
	// The name of a secret containing AMQ SSL related files.
	AMQSecretName string `json:"amqSecretName,omitempty"` // AMQ SSL parameters
	// The name of the AMQ SSL Trust Store file.
	AMQTruststoreName string `json:"amqTruststoreName,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password for the AMQ Trust Store.
	AMQTruststorePassword string `json:"amqTruststorePassword,omitempty"`
	// The name of the AMQ keystore file.
	AMQKeystoreName string `json:"amqKeystoreName,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password for the AMQ keystore and certificate.
	AMQKeystorePassword string `json:"amqKeystorePassword,omitempty"`
	// Not intended to be set by the user, if will be set to true if all required SSL parameters are set.
	AMQEnableSSL bool `json:"amqEnableSSL,omitempty"` // flag will be set to true if all AMQ SSL parameters are correctly set.
}

// JvmObject JVM specification to be used by the KieApp
type JvmObject struct {
	// User specified Java options to be appended to generated options in JAVA_OPTS. e.g. '-Dsome.property=foo'
	JavaOptsAppend string `json:"javaOptsAppend,omitempty"`
	// Is used when no '-Xmx' option is given in JAVA_OPTS. This is used to calculate a default maximal heap memory based on a containers restriction. If used in a container without any memory constraints for the container then this option has no effect. If there is a memory constraint then '-Xmx' is set to a ratio of the container available memory as set here. The default is '50' which means 50% of the available memory is used as an upper boundary. You can skip this mechanism by setting this value to '0' in which case no '-Xmx' option is added.
	JavaMaxMemRatio *int32 `json:"javaMaxMemRatio,omitempty"`
	// Is used when no '-Xms' option is given in JAVA_OPTS. This is used to calculate a default initial heap memory based on the maximum heap memory. If used in a container without any memory constraints for the container then this option has no effect. If there is a memory constraint then '-Xms' is set to a ratio of the '-Xmx' memory as set here. The default is '25' which means 25% of the '-Xmx' is used as the initial heap size. You can skip this mechanism by setting this value to '0' in which case no '-Xms' option is added. e.g. '25'
	JavaInitialMemRatio *int32 `json:"javaInitialMemRatio,omitempty"`
	// Is used when no '-Xms' option is given in JAVA_OPTS. This is used to calculate the maximum value of the initial heap memory. If used in a container without any memory constraints for the container then this option has no effect. If there is a memory constraint then '-Xms' is limited to the value set here. The default is 4096Mb which means the calculated value of '-Xms' never will be greater than 4096Mb. The value of this variable is expressed in MB. e.g. '4096'
	JavaMaxInitialMem *int32 `json:"javaMaxInitialMem,omitempty"`
	// Set this to get some diagnostics information to standard output when things are happening. Disabled by default. e.g. 'true'
	JavaDiagnostics *bool `json:"javaDiagnostics,omitempty"`
	// If set remote debugging will be switched on. Disabled by default. e.g. 'true'
	JavaDebug *bool `json:"javaDebug,omitempty"`
	// Port used for remote debugging. Defaults to 5005. e.g. '8787'
	JavaDebugPort *int32 `json:"javaDebugPort,omitempty"`
	// Minimum percentage of heap free after GC to avoid expansion. e.g. '20'
	GcMinHeapFreeRatio *int32 `json:"gcMinHeapFreeRatio,omitempty"`
	// Maximum percentage of heap free after GC to avoid shrinking. e.g. '40'
	GcMaxHeapFreeRatio *int32 `json:"gcMaxHeapFreeRatio,omitempty"`
	// Specifies the ratio of the time spent outside the garbage collection (for example, the time spent for application execution) to the time spent in the garbage collection, it's desirable that not more than 1 / (1 + n) e.g. 99 and means 1% spent on gc, 4 means spent 20% on gc.
	GcTimeRatio *int32 `json:"gcTimeRatio,omitempty"`
	// The weighting given to the current GC time versus previous GC times  when determining the new heap size. e.g. '90'
	GcAdaptiveSizePolicyWeight *int32 `json:"gcAdaptiveSizePolicyWeight,omitempty"`
	// The maximum metaspace size in Mega bytes unit e.g. 400
	GcMaxMetaspaceSize *int32 `json:"gcMaxMetaspaceSize,omitempty"`
	// Specify Java GC to use. The value of this variable should contain the necessary JRE command-line options to specify the required GC, which will override the default of '-XX:+UseParallelOldGC'. e.g. '-XX:+UseG1GC'
	GcContainerOptions string `json:"gcContainerOptions,omitempty"`
}

// KieAppObject generic object definition
type KieAppObject struct {
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Replicas to set for the DeploymentConfig
	Replicas  *int32                       `json:"replicas,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Keystore secret name
	KeystoreSecret string `json:"keystoreSecret,omitempty"`
	// The image context to use  e.g. rhpam-7, this param is optional for custom image.
	ImageContext string `json:"imageContext,omitempty"`
	// The image to use e.g. rhpam-<app>-rhel8, this param is optional for custom image.
	Image string `json:"image,omitempty"`
	// The image tag to use e.g. 7.9.0, this param is optional for custom image.
	ImageTag string `json:"imageTag,omitempty"`
	// The storageClassName to use
	StorageClassName string `json:"storageClassName,omitempty"`
}

// KieAppBuildObject Data to define how to build an application from source
type KieAppBuildObject struct {
	// Env set environment variables for BuildConfigs
	Env []corev1.EnvVar `json:"env,omitempty"`
	// The Maven GAV to deploy, e.g., rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.5.0-SNAPSHOT
	KieServerContainerDeployment string `json:"kieServerContainerDeployment,omitempty"`
	// Disable Maven pull dependencies for immutable KIE Server configurations for S2I and pre built kjars. Useful for pre-compiled kjar.
	DisablePullDeps bool `json:"disablePullDeps,omitempty"`
	// Disable Maven KIE Jar verification. It is recommended to test the kjar manually before disabling this verification.
	DisableKCVerification bool      `json:"disableKCVerification,omitempty"`
	GitSource             GitSource `json:"gitSource,omitempty"`
	// Maven mirror to use for S2I builds
	MavenMirrorURL string `json:"mavenMirrorURL,omitempty"`
	// List of directories from which archives will be copied into the deployment folder. If unspecified, all archives in /target will be copied.
	ArtifactDir string `json:"artifactDir,omitempty"`
	// +kubebuilder:validation:MinItems:=1
	Webhooks []WebhookSecret `json:"webhooks,omitempty"`
	From     *ImageObjRef    `json:"from,omitempty"`
	// ImageStreamTag definition for the image containing the drivers and configuration. For example, custom-driver-image:7.7.0.
	ExtensionImageStreamTag string `json:"extensionImageStreamTag,omitempty"`
	// Namespace within which the ImageStream definition for the image containing the drivers and configuration is located. Defaults to openshift namespace.
	ExtensionImageStreamTagNamespace string `json:"extensionImageStreamTagNamespace,omitempty"`
	// Full path to the directory within the extensions image where the extensions are located (e.g. install.sh, modules/, etc.).
	ExtensionImageInstallDir string `json:"extensionImageInstallDir,omitempty"`
}

// GitSource Git coordinates to locate the source code to build
type GitSource struct {
	// +kubebuilder:validation:Required
	// Git URI for the s2i source
	URI string `json:"uri"`
	// +kubebuilder:validation:Required
	// Branch to use in the git repository
	Reference string `json:"reference"`
	// Context/subdirectory where the code is located, relatively to repo root
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
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=GitHub;Generic
	Type WebhookType `json:"type"`
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`
}

// GitHooksVolume GitHooks volume configuration
type GitHooksVolume struct {
	// Absolute path where the gitHooks folder will be mounted.
	MountPath string  `json:"mountPath,omitempty"`
	From      *ObjRef `json:"from,omitempty"`
	// Secret to use for ssh key and known hosts file. The secret must contain two files: id_rsa and known_hosts.
	SSHSecret string `json:"sshSecret,omitempty"`
}

// KieAppAuthObject Authentication specification to be used by the KieApp
type KieAppAuthObject struct {
	SSO  *SSOAuthConfig  `json:"sso,omitempty"`
	LDAP *LDAPAuthConfig `json:"ldap,omitempty"`
	// When present, the RoleMapping Login Module will be configured.
	RoleMapper *RoleMapperAuthConfig `json:"roleMapper,omitempty"`
}

// SSOAuthConfig Authentication configuration for SSO
type SSOAuthConfig struct {
	// +kubebuilder:validation:Format:=password
	// RH-SSO Realm Admin Password used to create the Client
	AdminPassword string `json:"adminPassword,omitempty"`
	// RH-SSO Realm Admin Username used to create the Client if it doesn't exist
	AdminUser string `json:"adminUser,omitempty"`
	// +kubebuilder:validation:Required
	// RH-SSO URL
	URL string `json:"url"`
	// +kubebuilder:validation:Required
	// RH-SSO Realm name
	Realm string `json:"realm"`
	// RH-SSO Disable SSL Certificate Validation
	DisableSSLCertValidation bool `json:"disableSSLCertValidation,omitempty"`
	// RH-SSO Principal Attribute to use as username
	PrincipalAttribute string `json:"principalAttribute,omitempty"`
}

// SSOAuthClient Auth client to use for the SSO integration
type SSOAuthClient struct {
	// +kubebuilder:validation:Format:=password
	// Client secret
	Secret string `json:"secret,omitempty"`
	// Client name
	Name string `json:"name,omitempty"`
	// Hostname to set as redirect URL
	HostnameHTTP string `json:"hostnameHTTP,omitempty"`
	// Secure hostname to set as redirect URL
	HostnameHTTPS string `json:"hostnameHTTPS,omitempty"`
}

// LDAPAuthConfig Authentication configuration for LDAP
type LDAPAuthConfig struct {
	// +kubebuilder:validation:Format:=password
	// LDAP Credentials used for authentication
	BindCredential string `json:"bindCredential,omitempty"`
	// +kubebuilder:validation:Required
	// LDAP endpoint to connect for authentication. For failover set two or more LDAP endpoints separated by space
	URL string `json:"url"`
	// Bind DN used for authentication
	BindDN string `json:"bindDN,omitempty"`
	// The JMX ObjectName of the JaasSecurityDomain used to decrypt the password.
	JAASSecurityDomain string `json:"jaasSecurityDomain,omitempty"`
	// +kubebuilder:validation:Enum:=optional;required
	LoginModule LoginModuleType `json:"loginModule,omitempty"`
	// LDAP Base DN of the top-level context to begin the user search.
	BaseCtxDN string `json:"baseCtxDN,omitempty"`
	// DAP search filter used to locate the context of the user to authenticate. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. A common example for the search filter is (uid={0}).
	BaseFilter string `json:"baseFilter,omitempty"`
	// +kubebuilder:validation:Enum:=SUBTREE_SCOPE;OBJECT_SCOPE;ONELEVEL_SCOPE
	SearchScope SearchScopeType `json:"searchScope,omitempty"`
	// The timeout in milliseconds for user or role searches.
	SearchTimeLimit int32 `json:"searchTimeLimit,omitempty"`
	// The name of the attribute in the user entry that contains the DN of the user. This may be necessary if the DN of the user itself contains special characters, backslash for example, that prevent correct user mapping. If the attribute does not exist, the entry’s DN is used.
	DistinguishedNameAttribute string `json:"distinguishedNameAttribute,omitempty"`
	// A flag indicating if the DN is to be parsed for the username. If set to true, the DN is parsed for the username. If set to false the DN is not parsed for the username. This option is used together with usernameBeginString and usernameEndString.
	ParseUsername bool `json:"parseUsername,omitempty"`
	// Defines the String which is to be removed from the start of the DN to reveal the username. This option is used together with usernameEndString and only taken into account if parseUsername is set to true.
	UsernameBeginString string `json:"usernameBeginString,omitempty"`
	// Defines the String which is to be removed from the end of the DN to reveal the username. This option is used together with usernameBeginString and only taken into account if parseUsername is set to true.
	UsernameEndString string `json:"usernameEndString,omitempty"`
	// Name of the attribute containing the user roles.
	RoleAttributeID string `json:"roleAttributeID,omitempty"`
	// The fixed DN of the context to search for user roles. This is not the DN where the actual roles are, but the DN where the objects containing the user roles are. For example, in a Microsoft Active Directory server, this is the DN where the user account is.
	RolesCtxDN string `json:"rolesCtxDN,omitempty"`
	// A search filter used to locate the roles associated with the authenticated user. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. The authenticated userDN is substituted into the filter anywhere a {1} is used. An example search filter that matches on the input username is (member={0}). An alternative that matches on the authenticated userDN is (member={1}).
	RoleFilter string `json:"roleFilter,omitempty"`
	// +kubebuilder:validation:Format:=int16
	// The number of levels of recursion the role search will go below a matching context. Disable recursion by setting this to 0.
	RoleRecursion int16 `json:"roleRecursion,omitempty"`
	// A role included for all authenticated users
	DefaultRole string `json:"defaultRole,omitempty"`
	// Name of the attribute within the roleCtxDN context which contains the role name. If the roleAttributeIsDN property is set to true, this property is used to find the role object’s name attribute.
	RoleNameAttributeID string `json:"roleNameAttributeID,omitempty"`
	// A flag indicating if the DN returned by a query contains the roleNameAttributeID. If set to true, the DN is checked for the roleNameAttributeID. If set to false, the DN is not checked for the roleNameAttributeID. This flag can improve the performance of LDAP queries.
	ParseRoleNameFromDN bool `json:"parseRoleNameFromDN,omitempty"`
	// Whether or not the roleAttributeID contains the fully-qualified DN of a role object. If false, the role name is taken from the value of the roleNameAttributeId attribute of the context name. Certain directory schemas, such as Microsoft Active Directory, require this attribute to be set to true.
	RoleAttributeIsDN bool `json:"roleAttributeIsDN,omitempty"`
	// If you are not using referrals, you can ignore this option. When using referrals, this option denotes the attribute name which contains users defined for a certain role, for example member, if the role object is inside the referral. Users are checked against the content of this attribute name. If this option is not set, the check will always fail, so role objects cannot be stored in a referral tree.
	ReferralUserAttributeIDToCheck string `json:"referralUserAttributeIDToCheck,omitempty"`
}

// A flag to set login module to optional. The default value is required
type LoginModuleType string

const (
	//OptionalLoginModule optional login module
	OptionalLoginModule LoginModuleType = "optional"
	//RequiredLoginModule required login module
	RequiredLoginModule LoginModuleType = "required"
)

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
	// +kubebuilder:validation:Required
	RolesProperties string  `json:"rolesProperties"`
	ReplaceRole     bool    `json:"replaceRole,omitempty"`
	From            *ObjRef `json:"from,omitempty"`
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
	InternalDatabaseObject `json:",inline"`
	ExternalConfig         *ExternalDatabaseObject `json:"externalConfig,omitempty"`
}

// ProcessMigrationDatabaseObject Defines how a Process Migration server will manage
// and create a new Database or connect to an existing one
type ProcessMigrationDatabaseObject struct {
	InternalDatabaseObject `json:",inline"`
	ExternalConfig         *CommonExtDBObjectRequiredURL `json:"externalConfig,omitempty"`
}

// InternalDatabaseObject Defines how a deployment will manage and create a new Database
type InternalDatabaseObject struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=mysql;postgresql;external;h2
	// Database type to use
	Type DatabaseType `json:"type"`
	// Size of the PersistentVolumeClaim to create. For example, 100Gi
	Size string `json:"size,omitempty"`
	// The storageClassName to use for database pvc's.
	StorageClassName string `json:"storageClassName,omitempty"`
}

// CommonExtDBObjectRequiredURL common configuration definition of an external database
type CommonExtDBObjectRequiredURL struct {
	// +kubebuilder:validation:Required
	// Database JDBC URL. For example, jdbc:mysql:mydb.example.com:3306/rhpam
	JdbcURL                      string `json:"jdbcURL"`
	CommonExternalDatabaseObject `json:",inline"`
}

// CommonExtDBObjectURL common configuration definition of an external database
type CommonExtDBObjectURL struct {
	// Database JDBC URL. For example, jdbc:mysql:mydb.example.com:3306/rhpam
	JdbcURL                      string `json:"jdbcURL,omitempty"`
	CommonExternalDatabaseObject `json:",inline"`
}

// CommonExternalDatabaseObject common configuration definition of an external database
type CommonExternalDatabaseObject struct {
	// +kubebuilder:validation:Required
	// Driver name to use. For example, mysql
	Driver string `json:"driver"`
	// +kubebuilder:validation:Required
	// External database username
	Username string `json:"username"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=password
	// External database password
	Password string `json:"password"`
	// Sets xa-pool/min-pool-size for the configured datasource.
	MinPoolSize string `json:"minPoolSize,omitempty"`
	// Sets xa-pool/max-pool-size for the configured datasource.
	MaxPoolSize string `json:"maxPoolSize,omitempty"`
	// An org.jboss.jca.adapters.jdbc.ValidConnectionChecker that provides a SQLException isValidConnection(Connection e) method to validate if a connection is valid.
	ConnectionChecker string `json:"connectionChecker,omitempty"`
	// An org.jboss.jca.adapters.jdbc.ExceptionSorter that provides a boolean isExceptionFatal(SQLException e) method to validate if an exception should be broadcast to all javax.resource.spi.ConnectionEventListener as a connectionErrorOccurred.
	ExceptionSorter string `json:"exceptionSorter,omitempty"`
	// Sets the sql validation method to background-validation, if set to false the validate-on-match method will be used.
	BackgroundValidation string `json:"backgroundValidation,omitempty"`
	// Defines the interval for the background-validation check for the jdbc connections.
	BackgroundValidationMillis string `json:"backgroundValidationMillis,omitempty"`
}

// ExternalDatabaseObject configuration definition of an external database
type ExternalDatabaseObject struct {
	// +kubebuilder:validation:Required
	// Hibernate dialect class to use. For example, org.hibernate.dialect.MySQL8Dialect
	Dialect string `json:"dialect"`
	// Database Name. For example, rhpam
	Name string `json:"name,omitempty"`
	// Database Host. For example, mydb.example.com
	Host string `json:"host,omitempty"`
	// Database Port. For example, 3306
	Port string `json:"port,omitempty"`
	// Sets the datasources type. It can be XA or NONXA. For non XA set it to true. Default value is false.
	NonXA                string `json:"nonXA,omitempty"`
	CommonExtDBObjectURL `json:",inline"`
}

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

type Environment struct {
	Console          CustomObject   `json:"console,omitempty"`
	SmartRouter      CustomObject   `json:"smartRouter,omitempty"`
	Servers          []CustomObject `json:"servers,omitempty"`
	ProcessMigration CustomObject   `json:"processMigration,omitempty"`
	Dashbuilder      CustomObject   `json:"dashbuilder,omitempty"`
	Databases        []CustomObject `json:"databases,omitempty"`
	Others           []CustomObject `json:"others,omitempty"`
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
	ConfigMaps             []corev1.ConfigMap             `json:"configMaps,omitempty"`
}

type OpenShiftObject interface {
	metav1.Object
	runtime.Object
}

type EnvTemplate struct {
	*CommonConfig    `json:",inline"`
	Console          ConsoleTemplate          `json:"console,omitempty"`
	Servers          []ServerTemplate         `json:"servers,omitempty"`
	SmartRouter      SmartRouterTemplate      `json:"smartRouter,omitempty"`
	Auth             AuthTemplate             `json:"auth,omitempty"`
	ProcessMigration ProcessMigrationTemplate `json:"processMigration,omitempty"`
	Dashbuilder      DashbuilderTemplate      `json:"dashbuilder,omitempty"`
	Databases        []DatabaseTemplate       `json:"databases,omitempty"`
	Constants        TemplateConstants        `json:"constants,omitempty"`
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
	BrokerImageContext   string `json:"brokerImageContext"`
	BrokerImage          string `json:"brokerImage"`
	BrokerImageTag       string `json:"brokerImageTag"`
	DatagridImageContext string `json:"datagridImageContext"`
	DatagridImage        string `json:"datagridImage"`
	DatagridImageTag     string `json:"datagridImageTag"`
	MySQLImageURL        string `json:"mySQLImageURL"`
	PostgreSQLImageURL   string `json:"postgreSQLImageURL"`
	BrokerImageURL       string `json:"brokerImageURL,omitempty"`
	DatagridImageURL     string `json:"datagridImageURL,omitempty"`
	RoleMapperVolume     string `json:"roleMapperVolume"`
	GitHooksVolume       string `json:"gitHooksVolume,omitempty"`
	GitHooksSSHSecret    string `json:"gitHooksSSHSecret,omitempty"`
}

// ConsoleTemplate contains all the variables used in the yaml templates
type ConsoleTemplate struct {
	OmitImageStream     bool           `json:"omitImageStream"`
	SSOAuthClient       SSOAuthClient  `json:"ssoAuthClient,omitempty"`
	Name                string         `json:"name,omitempty"`
	Replicas            int32          `json:"replicas,omitempty"`
	ImageContext        string         `json:"imageContext,omitempty"`
	Image               string         `json:"image,omitempty"`
	ImageTag            string         `json:"imageTag,omitempty"`
	ImageURL            string         `json:"imageURL,omitempty"`
	KeystoreSecret      string         `json:"keystoreSecret,omitempty"`
	GitHooks            GitHooksVolume `json:"gitHooks,omitempty"`
	Jvm                 JvmObject      `json:"jvm,omitempty"`
	StorageClassName    string         `json:"storageClassName,omitempty"`
	PvSize              string         `json:"pvSize,omitempty"`
	Simplified          bool           `json:"simplifed"`
	DashbuilderLocation string         `json:"dashbuilderLocation,omitempty"`
}

// ServerTemplate contains all the variables used in the yaml templates
type ServerTemplate struct {
	OmitImageStream  bool              `json:"omitImageStream"`
	OmitConsole      bool              `json:"omitConsole"`
	KieName          string            `json:"kieName,omitempty"`
	KieServerID      string            `json:"kieServerID,omitempty"`
	Replicas         int32             `json:"replicas,omitempty"`
	SSOAuthClient    SSOAuthClient     `json:"ssoAuthClient,omitempty"`
	From             ImageObjRef       `json:"from,omitempty"`
	ImageURL         string            `json:"imageURL,omitempty"`
	Build            BuildTemplate     `json:"build,omitempty"`
	KeystoreSecret   string            `json:"keystoreSecret,omitempty"`
	Database         DatabaseObject    `json:"database,omitempty"`
	Jms              KieAppJmsObject   `json:"jms,omitempty"`
	SmartRouter      SmartRouterObject `json:"smartRouter,omitempty"`
	Jvm              JvmObject         `json:"jvm,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
}

// DashbuilderTemplate contains all the variables used in the yaml templates
type DashbuilderTemplate struct {
	OmitImageStream  bool              `json:"omitImageStream"`
	Name             string            `json:"name,omitempty"`
	Replicas         int32             `json:"replicas,omitempty"`
	SSOAuthClient    SSOAuthClient     `json:"ssoAuthClient,omitempty"`
	ImageContext     string            `json:"imageContext,omitempty"`
	Image            string            `json:"image,omitempty"`
	ImageTag         string            `json:"imageTag,omitempty"`
	ImageURL         string            `json:"imageURL,omitempty"`
	KeystoreSecret   string            `json:"keystoreSecret,omitempty"`
	Database         DatabaseObject    `json:"database,omitempty"`
	Jvm              JvmObject         `json:"jvm,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	Config           DashbuilderConfig `json:"config,omitempty"`
}

// DashbuilderConfig holds all configurations that can be applied to the Dashbuilder env
type DashbuilderConfig struct {
	// Enables integration with Business Central
	// +optional
	EnableBusinessCentral bool `json:"enableBusinessCentral,omitempty"`
	// Enables integration with KIE Server
	// +optional
	EnableKieServer bool `json:"enableKieServer,omitempty"`
	// Allow download of external (remote) files into runtime. Default value is false
	// +optional
	AllowExternalFileRegister *bool `json:"allowExternalFileRegister,omitempty"`
	// Components will be partitioned by the Runtime Model ID. Default value is true
	// +optional
	ComponentPartition *bool `json:"componentPartition,omitempty"`
	// Datasets IDs will partitioned by the Runtime Model ID. Default value is true
	// +optional
	DataSetPartition *bool `json:"dataSetPartition,omitempty"`
	// Set a static dashboard to run with runtime. When this property is set no new imports are allowed.
	// +optional
	ImportFileLocation string `json:"importFileLocation,omitempty"`
	// Make Dashbuilder not ephemeral. If ImportFileLocation is set PersistentConfigs will be ignored.
	// Default value is true.
	// +optional
	PersistentConfigs *bool `json:"persistentConfigs,omitempty"`
	// Base Directory where dashboards ZIPs are stored. If PersistentConfigs is enabled and ImportsBaseDir is not
	// pointing to a already existing PV the /opt/kie/dashbuilder/imports directory will be used. If ImportFileLocation is set
	// ImportsBaseDir will be ignored.
	// +optional
	ImportsBaseDir string `json:"importsBaseDir,omitempty"`
	// Allows Runtime to check model last update in FS to update its content. Default value is true.
	// +optional
	ModelUpdate *bool `json:"modelUpdate,omitempty"`
	// When enabled will also remove actual model file from file system. Default value is false.
	// +optional
	ModelFileRemoval *bool `json:"modelFileRemoval,omitempty"`
	// Runtime will always allow use of new imports (multi tenancy). Default value is false.
	// +optional
	RuntimeMultipleImport *bool `json:"runtimeMultipleImport,omitempty"`
	// Limits the size of uploaded dashboards (in kb). Default value is 10485760 kb.
	// +optional
	UploadSize *int64 `json:"uploadSize,omitempty"`
	// When set to true enables external components.
	// +optional
	ComponentEnable *bool `json:"componentEnable,omitempty"`
	// Base Directory where dashboards ZIPs are stored. If PersistentConfigs is enabled and ExternalCompDir is not
	// pointing to a already existing PV the /opt/kie/dashbuilder/components directory will be used.
	// +optional
	ExternalCompDir string `json:"externalCompDir,omitempty"`
	// Properties file with Dashbuilder configurations, if set, uniq properties will be appended and, if a property
	// is set mode than once, the one from this property file will be used.
	// +optional
	ConfigMapProps string `json:"configMapProps,omitempty"`
	// Defines the KIE Server Datasets access configurations
	// +optional
	KieServerDataSets []KieServerDataSetOrTemplate `json:"kieServerDataSets,omitempty"`
	// Defines the KIE Server Templates access configurations
	// +optional
	KieServerTemplates []KieServerDataSetOrTemplate `json:"kieServerTemplates,omitempty"`
}

type KieServerDataSetOrTemplate struct {
	Name         string `json:"name,omitempty"`
	Location     string `json:"location,omitempty"`
	User         string `json:"user,omitempty"`
	Password     string `json:"password,omitempty"`
	Token        string `json:"token,omitempty"`
	ReplaceQuery string `json:"replaceQuery,omitempty"`
}

// DatabaseTemplate contains all the variables used in the yaml templates
type DatabaseTemplate struct {
	InternalDatabaseObject `json:",inline"`
	ServerName             string `json:"serverName,omitempty"`
	Username               string `json:"username,omitempty"`
	DatabaseName           string `json:"databaseName,omitempty"`
}

// SmartRouterTemplate contains all the variables used in the yaml templates
type SmartRouterTemplate struct {
	OmitImageStream  bool   `json:"omitImageStream"`
	Replicas         int32  `json:"replicas,omitempty"`
	KeystoreSecret   string `json:"keystoreSecret,omitempty"`
	Protocol         string `json:"protocol,omitempty"`
	UseExternalRoute bool   `json:"useExternalRoute,omitempty"`
	ImageContext     string `json:"imageContext,omitempty"`
	Image            string `json:"image,omitempty"`
	ImageTag         string `json:"imageTag,omitempty"`
	ImageURL         string `json:"imageURL,omitempty"`
	StorageClassName string `json:"storageClassName,omitempty"`
}

// ReplicaConstants contains the default replica amounts for a component in a given environment type
type ReplicaConstants struct {
	Console     Replicas `json:"console,omitempty"`
	Dashbuilder Replicas `json:"dashbuilder,omitempty"`
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
	From                         ImageObjRef `json:"from,omitempty"`
	GitSource                    GitSource   `json:"gitSource,omitempty"`
	GitHubWebhookSecret          string      `json:"githubWebhookSecret,omitempty"`
	GenericWebhookSecret         string      `json:"genericWebhookSecret,omitempty"`
	KieServerContainerDeployment string      `json:"kieServerContainerDeployment,omitempty"`
	DisablePullDeps              bool        `json:"disablePullDeps,omitempty"`
	DisableKCVerification        bool        `json:"disableKCVerification,omitempty"`
	MavenMirrorURL               string      `json:"mavenMirrorURL,omitempty"`
	ArtifactDir                  string      `json:"artifactDir,omitempty"`
	// Extension image configuration which provides custom jdbc drivers to be used
	// by KieServer.
	ExtensionImageStreamTag          string `json:"extensionImageStreamTag,omitempty"`
	ExtensionImageStreamTagNamespace string `json:"extensionImageStreamTagNamespace,omitempty"`
	ExtensionImageInstallDir         string `json:"extensionImageInstallDir,omitempty"`
}

// CommonConfig variables used in the templates
type CommonConfig struct {
	// The name of the application deployment.
	ApplicationName string `json:"applicationName,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password to use for keystore generation.
	KeyStorePassword string `json:"keyStorePassword,omitempty"`
	// The user to use for the admin.
	AdminUser string `json:"adminUser,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password to use for the adminUser.
	AdminPassword string `json:"adminPassword,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password to use for databases.
	DBPassword string `json:"dbPassword,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password to use for amq user.
	AMQPassword string `json:"amqPassword,omitempty"`
	// +kubebuilder:validation:Format:=password
	// The password to use for amq cluster user.
	AMQClusterPassword string `json:"amqClusterPassword,omitempty"`
}

// VersionConfigs ...
type VersionConfigs struct {
	APIVersion           string `json:"apiVersion,omitempty"`
	OseCliImageURL       string `json:"oseCliImageURL,omitempty"`
	OseCliComponent      string `json:"oseCliComponent,omitempty"`
	BrokerImageContext   string `json:"brokerImageContext,omitempty"`
	BrokerImage          string `json:"brokerImage,omitempty"`
	BrokerImageTag       string `json:"brokerImageTag,omitempty"`
	BrokerImageURL       string `json:"brokerImageURL,omitempty"`
	BrokerComponent      string `json:"brokerComponent,omitempty"`
	DatagridImageContext string `json:"datagridImageContext,omitempty"`
	DatagridImage        string `json:"datagridImage,omitempty"`
	DatagridImageTag     string `json:"datagridImageTag,omitempty"`
	DatagridImageURL     string `json:"datagridImageURL,omitempty"`
	DatagridComponent    string `json:"datagridComponent,omitempty"`
	MySQLImageURL        string `json:"mySQLImageURL,omitempty"`
	MySQLComponent       string `json:"mySQLComponent,omitempty"`
	PostgreSQLImageURL   string `json:"postgreSQLImageURL,omitempty"`
	PostgreSQLComponent  string `json:"postgreSQLComponent,omitempty"`
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

// ProcessMigrationObject configuration of the RHPAM PIM
type ProcessMigrationObject struct {
	// The image context to use for Process Instance Migration  e.g. rhpam-7, this param is optional for custom image.
	ImageContext string `json:"imageContext,omitempty"`
	// The image to use for Process Instance Migration e.g. rhpam-process-migration-rhel8, this param is optional for custom image.
	Image string `json:"image,omitempty"`
	// The image tag to use for Process Instance Migration e.g. 7.9.0, this param is optional for custom image.
	ImageTag string                         `json:"imageTag,omitempty"`
	Database ProcessMigrationDatabaseObject `json:"database,omitempty"`
}

// ProcessMigrationTemplate ...
type ProcessMigrationTemplate struct {
	OmitImageStream  bool                           `json:"omitImageStream"`
	ImageContext     string                         `json:"imageContext,omitempty"`
	Image            string                         `json:"image,omitempty"`
	ImageTag         string                         `json:"imageTag,omitempty"`
	ImageURL         string                         `json:"imageURL,omitempty"`
	KieServerClients []KieServerClient              `json:"kieServerClients,omitempty"`
	Database         ProcessMigrationDatabaseObject `json:"database,omitempty"`
}

// KieServerClient ...
type KieServerClient struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ObjRef contains enough information to let you inspect or modify the referred object.
type ObjRef struct {
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=ConfigMap;Secret;PersistentVolumeClaim
	Kind            string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	ObjectReference `json:",inline"`
}

// ImageObjRef contains enough information to let you inspect or modify the referred object.
type ImageObjRef struct {
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=ImageStreamTag;DockerImage
	Kind            string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	ObjectReference `json:",inline"`
}

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
	// UID of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
	// +optional
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,4,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,5,opt,name=apiVersion"`
	// Specific resourceVersion to which this reference is made, if any.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
	// +optional
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`

	// If referring to a piece of an object instead of an entire object, this string
	// should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
	// For example, if the object reference is to a container within a pod, this would take on a value like:
	// "spec.containers{name}" (where "name" refers to the name of the container that triggered
	// the event) or if no container name is specified "spec.containers[2]" (container with
	// index 2 in this pod). This syntax is chosen only to have some well-defined way of
	// referencing a part of an object.
	// TODO: this design is not final and this field is subject to change in the future.
	// +optional
	FieldPath string `json:"fieldPath,omitempty" protobuf:"bytes,7,opt,name=fieldPath"`
}

func init() {
	SchemeBuilder.Register(&KieApp{}, &KieAppList{})
}
