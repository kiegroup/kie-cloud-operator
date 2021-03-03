package constants

import (
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ocp4Versions - OpenShift minor versions used for image curation
var Ocp4Versions = []string{"4.7", "4.6", "4.5", "4.4", "4.3", "4.2", "4.1"}

const (
	// CurrentVersion product version supported
	CurrentVersion = "7.11.0"
	// PriorVersion1 product version supported
	PriorVersion1 = "7.10.1"
	// PriorVersion2 product version supported
	PriorVersion2 = "7.10.0"
)

// SupportedVersions - product versions this operator supports
var SupportedVersions = []string{CurrentVersion, PriorVersion1, PriorVersion2}

// VersionConstants ...
var VersionConstants = map[string]*api.VersionConfigs{
	CurrentVersion: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli311ImageURL,
		OseCliComponent:     OseCli311Component,
		BrokerImage:         BrokerImage,
		BrokerImageTag:      Broker78ImageTag,
		BrokerImageURL:      Broker78ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag16,
		DatagridImageURL:    Datagrid73ImageURL16,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL80ImageURL,
		MySQLComponent:      MySQL80Component,
		PostgreSQLImageURL:  PostgreSQL10ImageURL,
		PostgreSQLComponent: PostgreSQL10Component,
	},
	PriorVersion1: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli311ImageURL,
		OseCliComponent:     OseCli311Component,
		BrokerImage:         BrokerImage,
		BrokerImageTag:      Broker78ImageTag,
		BrokerImageURL:      Broker78ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag16,
		DatagridImageURL:    Datagrid73ImageURL16,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL80ImageURL,
		MySQLComponent:      MySQL80Component,
		PostgreSQLImageURL:  PostgreSQL10ImageURL,
		PostgreSQLComponent: PostgreSQL10Component,
	},
	PriorVersion2: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli311ImageURL,
		OseCliComponent:     OseCli311Component,
		BrokerImage:         BrokerImage,
		BrokerImageTag:      Broker78ImageTag,
		BrokerImageURL:      Broker78ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag16,
		DatagridImageURL:    Datagrid73ImageURL16,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL80ImageURL,
		MySQLComponent:      MySQL80Component,
		PostgreSQLImageURL:  PostgreSQL10ImageURL,
		PostgreSQLComponent: PostgreSQL10Component,
	},
}

const (
	// ProductName used for metering labels
	ProductName = "Red_Hat_Process_Automation"
	// LabelRHproductName used as metering label
	LabelRHproductName = "rht.prod_name"
	// LabelRHproductVersion used as metering label
	LabelRHproductVersion = "rht.prod_ver"
	// LabelRHcomponentName used as metering label
	LabelRHcomponentName = "rht.comp"
	// LabelRHcomponentVersion used as metering label
	LabelRHcomponentVersion = "rht.comp_ver"
	// LabelRHsubcomponentName used as metering label
	LabelRHsubcomponentName = "rht.subcomp"
	// LabelRHsubcomponentType used as metering label
	LabelRHsubcomponentType = "rht.subcomp_t"
	// LabelRHcompany used as metering label
	LabelRHcompany = "com.company"
	// RhpamPrefix RHPAM prefix
	RhpamPrefix = "rhpam"
	// RhdmPrefix RHDM prefix
	RhdmPrefix = "rhdm"
	// KieServerServicePrefix prefix to use for the servers
	KieServerServicePrefix = "kieserver"
	// ImageRegistry default registry
	ImageRegistry = "registry.redhat.io"
	// ImageRegistryStage default registry
	ImageRegistryStage = "registry.stage.redhat.io"
	// ImageRegistryBrew default registry
	ImageRegistryBrew = "registry-proxy.engineering.redhat.com"
	// ImageContextBrew default context
	ImageContextBrew = "rh-osbs"
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
	// KeystoreAlias used when creating entry in Keystore
	KeystoreAlias = "jboss"
	// KeystoreName used when creating Secret
	KeystoreName = "keystore.jks"
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
	// GitHooksSSHSecret Name of the mounted volume name when GitHooks SSH Secret reference is set
	GitHooksSSHSecret = "githooks-ssh-volume"
	// RoleMapperVolume Name of the mounted volume name when RoleMapper reference is set
	RoleMapperVolume = "rolemapper-volume"
	// RoleMapperDefaultDir Default path for the rolemapping properties file
	RoleMapperDefaultDir = "/opt/eap/standalone/configuration/rolemapping"
	// DefaultKieServerDatabaseName Default database name for Kie Server
	DefaultKieServerDatabaseName = "rhpam7"
	// DefaultKieServerDatabaseUsername Default database username for Kie Server
	DefaultKieServerDatabaseUsername = "rhpam"
	// DefaultProcessMigrationDatabaseType Default database type for Process Migration
	DefaultProcessMigrationDatabaseType = api.DatabaseH2
	// DefaultProcessMigrationDatabaseName Default database name for Process Migration
	DefaultProcessMigrationDatabaseName = "pimdb"
	// DefaultProcessMigrationDatabaseUsername Default database username for Process Migration
	DefaultProcessMigrationDatabaseUsername = "pim"
	// ProcessMigrationDefaultImageURL Process Migration Image
	ProcessMigrationDefaultImageURL = ImageRegistry + PamContext + "process-migration" + RhelVersion

	relatedImageVar = "RELATED_IMAGE_"

	DmKieImageVar          = relatedImageVar + "DM_KIESERVER_IMAGE_"
	DmDecisionCentralVar   = relatedImageVar + "DM_DC_IMAGE_"
	DmControllerVar        = relatedImageVar + "DM_CONTROLLER_IMAGE_"
	PamKieImageVar         = relatedImageVar + "PAM_KIESERVER_IMAGE_"
	PamControllerVar       = relatedImageVar + "PAM_CONTROLLER_IMAGE_"
	PamBusinessCentralVar  = relatedImageVar + "PAM_BC_IMAGE_"
	PamBCMonitoringVar     = relatedImageVar + "PAM_BC_MONITORING_IMAGE_"
	PamProcessMigrationVar = relatedImageVar + "PAM_PROCESS_MIGRATION_IMAGE_"
	PamDashbuilderVar      = relatedImageVar + "PAM_DASHBUILDER_IMAGE_"
	PamSmartRouterVar      = relatedImageVar + "PAM_SMARTROUTER_IMAGE_"

	OauthVar             = relatedImageVar + "OAUTH_PROXY_IMAGE_"
	Oauth3ImageLatestURL = ImageRegistry + "/openshift3/oauth-proxy:latest"
	Oauth4ImageURL       = ImageRegistry + "/openshift4/ose-oauth-proxy"
	Oauth4ImageLatestURL = Oauth4ImageURL + ":latest"
	OauthComponent       = "golang-github-openshift-oauth-proxy-container"

	PostgreSQLVar         = relatedImageVar + "POSTGRESQL_PROXY_IMAGE_"
	PostgreSQL10ImageURL  = ImageRegistry + "/rhscl/postgresql-10-rhel7:latest"
	PostgreSQL10Component = "rh-postgresql10-container"

	MySQLVar         = relatedImageVar + "MYSQL_PROXY_IMAGE_"
	MySQL57ImageURL  = ImageRegistry + "/rhscl/mysql-57-rhel7:latest"
	MySQL57Component = "rh-mysql57-container"
	MySQL80ImageURL  = ImageRegistry + "/rhscl/mysql-80-rhel7:latest"
	MySQL80Component = "rh-mysql80-container"

	OseCliVar          = relatedImageVar + "OSE_CLI_IMAGE_"
	OseCli311ImageURL  = ImageRegistry + "/openshift3/ose-cli:v3.11"
	OseCli311Component = "openshift-enterprise-cli-container"

	BrokerComponent = "amq-broker-openshift-container"
	BrokerVar       = relatedImageVar + "BROKER_IMAGE_"
	BrokerImage     = "amq-broker"
	BrokerImageURL  = ImageRegistry + "/amq7/" + BrokerImage + ":"

	Broker77ImageTag = "7.7"
	Broker78ImageTag = "7.8"
	Broker77ImageURL = BrokerImageURL + Broker77ImageTag
	Broker78ImageURL = BrokerImageURL + Broker78ImageTag

	DatagridVar         = relatedImageVar + "DATAGRID_IMAGE_"
	Datagrid73Image     = "datagrid73-openshift"
	Datagrid73Component = "jboss-datagrid-7-datagrid73-openshift-container"

	Datagrid73ImageTag15 = "1.5"
	Datagrid73ImageURL15 = ImageRegistry + "/jboss-datagrid-7/" + Datagrid73Image + ":" + Datagrid73ImageTag15

	Datagrid73ImageTag16 = "1.6"
	Datagrid73ImageURL16 = ImageRegistry + "/jboss-datagrid-7/" + Datagrid73Image + ":" + Datagrid73ImageTag16

	DmContext   = "/" + RhdmPrefix + "-7/" + RhdmPrefix + "-"
	PamContext  = "/" + RhpamPrefix + "-7/" + RhpamPrefix + "-"
	RhelVersion = "-rhel8"

	//Resources Limits and Requests
	ConsoleProdCPULimit         = "1"
	ConsoleProdMemLimit         = "2Gi"
	ConsoleAuthoringCPULimit    = "2"
	ConsoleAuthoringMemLimit    = "4Gi"
	ConsoleAuthoringCPURequests = "1500m"
	ConsoleAuthoringMemRequests = "3Gi"
	ConsoleProdCPURequests      = "500m"
	ConsoleProdMemRequests      = "1536Mi"
	ConsolePvSize               = "1Gi"
	ConsoleProdPvSize           = "64Mi"
	DashbuilderCPULimit         = "1"
	DashbuilderCPURequests      = "750m"
	DashbuilderMemLimit         = "2Gi"
	DashbuilderMemRequests      = "1536Mi"
	ServersCPULimit             = "1"
	ServersMemLimit             = "2Gi"
	ServersCPURequests          = "750m"
	ServersMemRequests          = "1Gi"
	SmartRouterCPULimit         = "500m"
	SmartRouterMemLimit         = "1Gi"
	SmartRouterCPURequests      = "250m"
	SmartRouterMemRequests      = "1Gi"

	//ImageNames for metering labels
	RhpamSmartRouterImageName = RhpamPrefix + "-smartrouter-" + RhelVersion
	RhpamControllerImageName  = RhpamPrefix + "-controller-" + RhelVersion
	RhdmSmartRouterImageName  = RhdmPrefix + "-smartrouter-" + RhelVersion
	RhdmControllerImageName   = RhdmPrefix + "-controller-" + RhelVersion

	RhdmDecisionCentral     = "decisioncentral"
	RhpamBusinessCentral    = "businesscentral"
	RhpamBusinessCentralMon = "businesscentral-monitoring"

	DashBuilder      = "dashbuilder"
	Smartrouter      = "smartrouter"
	ProcessMigration = "process-migration"
	Production       = "production"

	SUBCOMPONENT_TYPE_APP   = "application"
	SUBCOMPONENT_TYPE_INFRA = "infrastructure"

	//Labels for Kube_ping
	ClusterLabel       = "cluster"
	ClusterLabelPrefix = "jgrp.k8s."

	//Envs
	KubeNS     = "KUBERNETES_NAMESPACE"
	KubeLabels = "KUBERNETES_LABELS"
)

// Console Resource Limits for BC Monitoring in Prod Env
var ConsoleProdLimits = map[string]string{
	"CPU": ConsoleProdCPULimit,
	"MEM": ConsoleProdMemLimit,
}

// Console Resource Limits for BC in Authoring Env
var ConsoleAuthoringLimits = map[string]string{
	"CPU": ConsoleAuthoringCPULimit,
	"MEM": ConsoleAuthoringMemLimit,
}

// Dashbuilder Resource Limits for Dasubuilder deployment
var DashbuilderLimits = map[string]string{
	"CPU": DashbuilderCPULimit,
	"MEM": DashbuilderMemLimit,
}

// Server Limits for every Env
var ServersLimits = map[string]string{
	"CPU": ServersCPULimit,
	"MEM": ServersMemLimit,
}

// SmartRouter Limits for every Env
var SmartRouterLimits = map[string]string{
	"CPU": SmartRouterCPULimit,
	"MEM": SmartRouterMemLimit,
}

// ConsoleAuthoringRequests defines requests in Authoring environment
var ConsoleAuthoringRequests = map[string]string{
	"CPU": ConsoleAuthoringCPURequests,
	"MEM": ConsoleAuthoringMemRequests,
}

// ConsoleProdRequests defines requests in Prod or Immutable environment
var ConsoleProdRequests = map[string]string{
	"CPU": ConsoleProdCPURequests,
	"MEM": ConsoleProdMemRequests,
}

// Dashbuilder defines requests Dasubuilder deployment
var DashbuilderRequests = map[string]string{
	"CPU": DashbuilderCPURequests,
	"MEM": DashbuilderMemRequests,
}

// ServerRequests defines the requests for kieserver deployment
var ServerRequests = map[string]string{
	"CPU": ServersCPURequests,
	"MEM": ServersMemRequests,
}

// SmartRouterRequests defines the requests for smart router deployment
var SmartRouterRequests = map[string]string{
	"CPU": SmartRouterCPURequests,
	"MEM": SmartRouterMemRequests,
}

var Images = []ImageEnv{
	{
		Var:       DmKieImageVar,
		Component: RhdmPrefix + "-7-kieserver-rhel8-container",
		Registry:  ImageRegistry,
		Context:   DmContext + "kieserver" + RhelVersion,
	},
	{
		Var:       DmControllerVar,
		Component: RhdmPrefix + "-7-controller-rhel8-container",
		Registry:  ImageRegistry,
		Context:   DmContext + "controller" + RhelVersion,
	},
	{
		Var:       DmDecisionCentralVar,
		Component: RhdmPrefix + "-7-decisioncentral-rhel8-container",
		Registry:  ImageRegistry,
		Context:   DmContext + "decisioncentral" + RhelVersion,
	},
	{
		Var:       PamKieImageVar,
		Component: RhpamPrefix + "-7-kieserver-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "kieserver" + RhelVersion,
	},
	{
		Var:       PamControllerVar,
		Component: RhpamPrefix + "-7-controller-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "controller" + RhelVersion,
	},
	{
		Var:       PamBusinessCentralVar,
		Component: RhpamPrefix + "-7-businesscentral-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "businesscentral" + RhelVersion,
	},
	{
		Var:       PamBCMonitoringVar,
		Component: RhpamPrefix + "-7-businesscentral-monitoring-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "businesscentral-monitoring" + RhelVersion,
	},
	{
		Var:       PamSmartRouterVar,
		Component: RhpamPrefix + "-7-smartrouter-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "smartrouter" + RhelVersion,
	},
	{
		Var:       PamProcessMigrationVar,
		Component: RhpamPrefix + "-7-process-migration-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "process-migration" + RhelVersion,
	},
	{
		Var:       PamDashbuilderVar,
		Component: RhpamPrefix + "-7-dashbuilder-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "dashbuilder" + RhelVersion,
	},
}

type ImageEnv struct {
	Var       string
	Component string
	Registry  string
	Context   string
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

var rhpamAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", ImageVar: PamBusinessCentralVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central"}
var rhpamMonitorAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", ImageVar: PamBCMonitoringVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central Monitoring"}
var rhpamDashbuilderConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamdash", ImageName: "dashbuilder", ImageVar: PamDashbuilderVar, FriendlyName: "Dashbuilder"}
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

var replicasDashbuilder = api.ReplicaConstants{
	Dashbuilder: api.Replicas{Replicas: 1},
}

// DefaultDatabaseConfig defines the default Database to use for each environment
var databaseRhpamAuthoring = &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabaseH2, Size: DefaultDatabaseSize}}
var databaseRhpamAuthoringHA = &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabaseMySQL, Size: DefaultDatabaseSize}}
var databaseRhpamProduction = &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabasePostgreSQL, Size: DefaultDatabaseSize}}
var databaseRhpamProductionImmutable = &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabasePostgreSQL, Size: DefaultDatabaseSize}}
var databaseRhpamTrial = &api.DatabaseObject{InternalDatabaseObject: api.InternalDatabaseObject{Type: api.DatabaseH2, Size: ""}}

// EnvironmentConstants contains
var EnvironmentConstants = map[api.EnvironmentType]*api.EnvironmentConstants{
	api.RhpamProduction:            {App: rhpamMonitorAppConstants, Replica: replicasRhpamProduction, Database: databaseRhpamProduction},
	api.RhpamProductionImmutable:   {App: rhpamMonitorAppConstants, Replica: replicasRhpamProductionImmutable, Database: databaseRhpamProductionImmutable},
	api.RhpamTrial:                 {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamTrial},
	api.RhpamAuthoring:             {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamAuthoring},
	api.RhpamAuthoringHA:           {App: rhpamAppConstants, Replica: replicasAuthoringHA, Database: databaseRhpamAuthoringHA},
	api.RhpamStandaloneDashbuilder: {App: rhpamDashbuilderConstants, Replica: replicasDashbuilder},
	api.RhdmTrial:                  {App: rhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoring:              {App: rhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoringHA:            {App: rhdmAppConstants, Replica: replicasAuthoringHA},
	api.RhdmProductionImmutable:    {App: rhdmAppConstants, Replica: replicasTrial},
}

// TemplateConstants set of constant values to use in templates
var TemplateConstants = api.TemplateConstants{
	KeystoreVolumeSuffix: KeystoreVolumeSuffix,
	DatabaseVolumeSuffix: DatabaseVolumeSuffix,
	RoleMapperVolume:     RoleMapperVolume,
	GitHooksVolume:       GitHooksVolume,
	GitHooksSSHSecret:    GitHooksSSHSecret,
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
