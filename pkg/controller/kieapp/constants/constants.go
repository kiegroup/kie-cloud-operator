package constants

import (
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CurrentVersion product version supported
	CurrentVersion = "7.8.0"
	// PriorVersion1 product version supported
	PriorVersion1 = "7.7.0"
	// PriorVersion2 product version supported
	PriorVersion2 = "7.6.0"
)

// SupportedVersions - product versions this operator supports
var SupportedVersions = []string{CurrentVersion, PriorVersion1, PriorVersion2}

const (
	// RhpamPrefix RHPAM prefix
	RhpamPrefix = "rhpam"
	// RhdmPrefix RHDM prefix
	RhdmPrefix = "rhdm"
	// KieServerServicePrefix prefix to use for the servers
	KieServerServicePrefix = "kieserver"
	// ImageRegistry default registry
	ImageRegistry = "registry.redhat.io"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
	// ConfigMapPrefix prefix to use for the configmaps
	ConfigMapPrefix = "kieconfigs"
	// KieServerCMLabel the label to modify when replicas is set to 0
	KieServerCMLabel = "services.server.kie.org/kie-server-state"
	// DefaultAdminUser default admin user
	DefaultAdminUser = "adminUser"
	// DefaultPassword default password to use for test environments
	DefaultPassword = "RedHat"
	// SSODefaultPrincipalAttribute default PrincipalAttribute to use for SSO integration
	SSODefaultPrincipalAttribute = "preferred_username"
	// NameSpaceEnv is an environment variable of the current namespace
	// set via downward api when the code is running via deployment
	NameSpaceEnv = "WATCH_NAMESPACE"
	// OpNameEnv is an environment variable of the operator name
	// set when the code is running via deployment
	OpNameEnv = "OPERATOR_NAME"
	// OpUIEnv is an environment variable indicating whether the UI should be deployed
	// Default behavior is to deploy the UI, unless this variable is provided with a false value
	OpUIEnv = "OPERATOR_UI"
	// TrialEnvSuffix is the suffix for trial environments
	TrialEnvSuffix = "trial"
	// DefaultKieDeployments default number of Kie Server deployments
	DefaultKieDeployments = 1
	// KeystoreSecret is the default format for keystore secret names
	KeystoreSecret = "%s-app-secret"
	// KeystoreVolumeSuffix Suffix for the keystore volumes and volumeMounts name
	KeystoreVolumeSuffix = "keystore-volume"
	// DatabaseVolumeSuffix Suffix to use for any database volume and volumeMounts
	DatabaseVolumeSuffix = "pvol"
	// DefaultDatabaseSize Default Database Persistence size
	DefaultDatabaseSize = "1Gi"
	// DefaultExtensionImageInstallDir Default Extension Install Dir for JDBC drivers
	DefaultExtensionImageInstallDir = "/extensions"
	// ConsoleLinkName is how the link will be titled in an installed CSV within the marketplace
	ConsoleLinkName = "Installer"
	// ConsoleDescription is how the link will be described in an installed CSV within the marketplace
	ConsoleDescription = "**To use the guided installer to provision an environment, open the Installer link, in the links section on the left side of this page.**"
	// SmartRouterProtocol - default SmartRouter protocol
	SmartRouterProtocol = "http"
	// GitHooksDefaultDir Default path where to mount the GitHooks volume
	GitHooksDefaultDir = "/opt/kie/data/git/hooks"
	// GitHooksVolume Name of the mounted volume name when GitHooks reference is set
	GitHooksVolume = "githooks-volume"
	// RoleMapperVolume Name of the mounted volume name when RoleMapper reference is set
	RoleMapperVolume = "rolemapper-volume"
	// RoleMapperDefaultDir Default path for the rolemapping properties file
	RoleMapperDefaultDir = "/opt/eap/standalone/configuration/rolemapping"

	DmKieImageVar          = "DM_KIESERVER_IMAGE_"
	DmDecisionCentralVar   = "DM_DC_IMAGE_"
	DmControllerVar        = "DM_CONTROLLER_IMAGE_"
	PamKieImageVar         = "PAM_KIESERVER_IMAGE_"
	PamControllerVar       = "PAM_CONTROLLER_IMAGE_"
	PamBusinessCentralVar  = "PAM_BC_IMAGE_"
	PamBCMonitoringVar     = "PAM_BC_MONITORING_IMAGE_"
	PamProcessMigrationVar = "PAM_PROCESS_MIGRATION_IMAGE_"
	PamSmartRouterVar      = "PAM_SMARTROUTER_IMAGE_"

	OauthVar       = "OAUTH_PROXY_IMAGE"
	OauthImageURL  = ImageRegistry + "/openshift3/oauth-proxy:v3.11"
	OauthComponent = "golang-github-openshift-oauth-proxy-container"

	PostgreSQLVar         = "POSTGRESQL_PROXY_IMAGE_"
	PostgreSQL10ImageURL  = ImageRegistry + "/rhscl/postgresql-10-rhel7:latest"
	PostgreSQL10Component = "rh-postgresql10-container"

	MySQLVar         = "MYSQL_PROXY_IMAGE_"
	MySQL57ImageURL  = ImageRegistry + "/rhscl/mysql-57-rhel7:latest"
	MySQL57Component = "rh-mysql57-container"

	OseCliVar          = "OSE_CLI_IMAGE_"
	OseCli311ImageURL  = ImageRegistry + "/openshift3/ose-cli:v3.11"
	OseCli311Component = "openshift-enterprise-cli-container"

	BrokerVar         = "BROKER_IMAGE_"
	Broker75Image     = "amq-broker"
	Broker75ImageTag  = "7.5"
	Broker75ImageURL  = ImageRegistry + "/amq7/" + Broker75Image + ":" + Broker75ImageTag
	Broker75Component = "amq-broker-openshift-container"

	DatagridVar         = "DATAGRID_IMAGE_"
	Datagrid73Image     = "datagrid73-openshift"
	Datagrid73ImageTag  = "1.3"
	Datagrid73ImageURL  = ImageRegistry + "/jboss-datagrid-7/" + Datagrid73Image + ":" + Datagrid73ImageTag
	Datagrid73Component = "jboss-datagrid-7-datagrid73-openshift-container"

	DmContext   = ImageRegistry + "/rhdm-7/rhdm-"
	PamContext  = ImageRegistry + "/rhpam-7/rhpam-"
	RhelVersion = "-rhel8"

	ConsoleLinkFinalizer = "finalizer.console.openshift.io"
)

var Images = []ImageEnv{
	{
		Var:       DmKieImageVar,
		Component: "rhdm-7-kieserver-rhel8-container",
		Registry:  DmContext + "kieserver" + RhelVersion,
	},
	{
		Var:       DmControllerVar,
		Component: "rhdm-7-controller-rhel8-container",
		Registry:  DmContext + "controller" + RhelVersion,
	},
	{
		Var:       DmDecisionCentralVar,
		Component: "rhdm-7-decisioncentral-rhel8-container",
		Registry:  DmContext + "decisioncentral" + RhelVersion,
	},
	{
		Var:       PamKieImageVar,
		Component: "rhpam-7-kieserver-rhel8-container",
		Registry:  PamContext + "kieserver" + RhelVersion,
	},
	{
		Var:       PamControllerVar,
		Component: "rhpam-7-controller-rhel8-container",
		Registry:  PamContext + "controller" + RhelVersion,
	},
	{
		Var:       PamBusinessCentralVar,
		Component: "rhpam-7-businesscentral-rhel8-container",
		Registry:  PamContext + "businesscentral" + RhelVersion,
	},
	{
		Var:       PamBCMonitoringVar,
		Component: "rhpam-7-businesscentral-monitoring-rhel8-container",
		Registry:  PamContext + "businesscentral-monitoring" + RhelVersion,
	},
	{
		Var:       PamProcessMigrationVar,
		Component: "rhpam-7-process-migration-rhel8-container",
		Registry:  PamContext + "process-migration" + RhelVersion,
	},
	{
		Var:       PamSmartRouterVar,
		Component: "rhpam-7-smartrouter-rhel8-container",
		Registry:  PamContext + "smartrouter" + RhelVersion,
	},
}

type ImageEnv struct {
	Var       string
	Component string
	Registry  string
}
type ImageRef struct {
	metav1.TypeMeta `json:",inline"`
	Spec            ImageRefSpec `json:"spec"`
}
type ImageRefSpec struct {
	Tags []ImageRefTag `json:"tags"`
}
type ImageRefTag struct {
	Name string                  `json:"name"`
	From *corev1.ObjectReference `json:"from"`
}

// VersionConstants ...
var VersionConstants = map[string]*api.VersionConfigs{
	CurrentVersion: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli311ImageURL,
		OseCliComponent:     OseCli311Component,
		BrokerImage:         Broker75Image,
		BrokerImageTag:      Broker75ImageTag,
		BrokerImageURL:      Broker75ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag,
		DatagridImageURL:    Datagrid73ImageURL,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL57ImageURL,
		MySQLComponent:      MySQL57Component,
		PostgreSQLImageURL:  PostgreSQL10ImageURL,
		PostgreSQLComponent: PostgreSQL10Component,
	},
	PriorVersion1: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli311ImageURL,
		OseCliComponent:     OseCli311Component,
		BrokerImage:         Broker75Image,
		BrokerImageTag:      Broker75ImageTag,
		BrokerImageURL:      Broker75ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag,
		DatagridImageURL:    Datagrid73ImageURL,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL57ImageURL,
		MySQLComponent:      MySQL57Component,
		PostgreSQLImageURL:  PostgreSQL10ImageURL,
		PostgreSQLComponent: PostgreSQL10Component,
	},
	PriorVersion2: {
		APIVersion:       api.SchemeGroupVersion.Version,
		BrokerImage:      "amq-broker",
		BrokerImageTag:   "7.5",
		DatagridImage:    "datagrid73-openshift",
		DatagridImageTag: "1.3",
	},
}

var rhpamAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", ImageVar: PamBusinessCentralVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central"}
var rhpamMonitorAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", ImageVar: PamBCMonitoringVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central Monitoring"}
var rhdmAppConstants = api.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: "decisioncentral", ImageVar: DmDecisionCentralVar, MavenRepo: "RHDMCENTR", FriendlyName: "Decision Central"}

var replicasTrial = api.ReplicaConstants{
	Console:     api.Replicas{Replicas: 1, DenyScale: true},
	Server:      api.Replicas{Replicas: 1},
	SmartRouter: api.Replicas{Replicas: 1},
}
var replicasRhpamProductionImmutable = api.ReplicaConstants{
	Console:     api.Replicas{Replicas: 1},
	Server:      api.Replicas{Replicas: 2},
	SmartRouter: api.Replicas{Replicas: 1},
}
var replicasRhpamProduction = api.ReplicaConstants{
	Console:     api.Replicas{Replicas: 3},
	Server:      api.Replicas{Replicas: 3},
	SmartRouter: api.Replicas{Replicas: 1},
}
var replicasAuthoringHA = api.ReplicaConstants{
	Console:     api.Replicas{Replicas: 2},
	Server:      api.Replicas{Replicas: 2},
	SmartRouter: api.Replicas{Replicas: 1},
}

// DefaultDatabaseConfig defines the default Database to use for each environment
var databaseRhpamAuthoring = &api.DatabaseObject{Type: api.DatabaseH2, Size: DefaultDatabaseSize}
var databaseRhpamAuthoringHA = &api.DatabaseObject{Type: api.DatabaseMySQL, Size: DefaultDatabaseSize}
var databaseRhpamProduction = &api.DatabaseObject{Type: api.DatabasePostgreSQL, Size: DefaultDatabaseSize}
var databaseRhpamProductionImmutable = &api.DatabaseObject{Type: api.DatabasePostgreSQL, Size: DefaultDatabaseSize}
var databaseRhpamTrial = &api.DatabaseObject{Type: api.DatabaseH2, Size: ""}

// EnvironmentConstants contains
var EnvironmentConstants = map[api.EnvironmentType]*api.EnvironmentConstants{
	api.RhpamProduction:          {App: rhpamMonitorAppConstants, Replica: replicasRhpamProduction, Database: databaseRhpamProduction},
	api.RhpamProductionImmutable: {App: rhpamMonitorAppConstants, Replica: replicasRhpamProductionImmutable, Database: databaseRhpamProductionImmutable},
	api.RhpamTrial:               {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamTrial},
	api.RhpamAuthoring:           {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamAuthoring},
	api.RhpamAuthoringHA:         {App: rhpamAppConstants, Replica: replicasAuthoringHA, Database: databaseRhpamAuthoringHA},
	api.RhdmTrial:                {App: rhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoring:            {App: rhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoringHA:          {App: rhdmAppConstants, Replica: replicasAuthoringHA},
	api.RhdmProductionImmutable:  {App: rhdmAppConstants, Replica: replicasTrial},
}

// TemplateConstants set of constant values to use in templates
var TemplateConstants = api.TemplateConstants{
	KeystoreVolumeSuffix: KeystoreVolumeSuffix,
	DatabaseVolumeSuffix: DatabaseVolumeSuffix,
	RoleMapperVolume:     RoleMapperVolume,
	GitHooksVolume:       GitHooksVolume,
}

// DebugTrue - used to enable debug logs in objects
var DebugTrue = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "true",
}

// DebugFalse - used to disable debug logs in objects
var DebugFalse = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "false",
}
