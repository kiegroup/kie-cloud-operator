package constants

import (
	"golang.org/x/mod/semver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
)

// Ocp4Versions - OpenShift minor versions used for image curation
var Ocp4Versions = []string{"4.14", "4.13", "4.12"}

const (
	// CurrentVersion product version supported
	CurrentVersion = "8.0.5"
	// PriorVersion product version supported
	PriorVersion = "8.0.4"
)

// SupportedVersions - product versions this operator supports
var SupportedVersions = []string{CurrentVersion, PriorVersion}

// VersionConstants ...
var VersionConstants = map[string]*api.VersionConfigs{
	CurrentVersion: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli4ImageURL,
		OseCliComponent:     OseCli4Component,
		BrokerImage:         Broker79Image,
		BrokerImageTag:      Broker79ImageTag,
		BrokerImageURL:      Broker79ImageURL,
		DatagridImage:       Datagrid8Image,
		DatagridImageTag:    Datagrid8ImageTag11,
		DatagridImageURL:    Datagrid8ImageURL11,
		DatagridComponent:   Datagrid8Component,
		MySQLImageURL:       MySQL80ImageURL,
		MySQLComponent:      MySQL80Component,
		PostgreSQLImageURL:  PostgreSQL13ImageURL,
		PostgreSQLComponent: PostgreSQL13Component,
	},
	PriorVersion: {
		APIVersion:          api.SchemeGroupVersion.Version,
		OseCliImageURL:      OseCli4ImageURL,
		OseCliComponent:     OseCli4Component,
		BrokerImage:         BrokerImage,
		BrokerImageTag:      Broker78ImageTag,
		BrokerImageURL:      Broker78ImageURL,
		DatagridImage:       Datagrid73Image,
		DatagridImageTag:    Datagrid73ImageTag16,
		DatagridImageURL:    Datagrid73ImageURL16,
		DatagridComponent:   Datagrid73Component,
		MySQLImageURL:       MySQL80ImageURL,
		MySQLComponent:      MySQL80Component,
		PostgreSQLImageURL:  PostgreSQL12ImageURL,
		PostgreSQLComponent: PostgreSQL12Component,
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
	// IBMBamoeImageContext IBM image context
	IBMBamoeImageContext = "ibm-bamoe"
	// IBMBamoeImagePrefix IBM image prefix
	IBMBamoeImagePrefix = "bamoe"
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
	// TruststoreSecret is the default format for truststore secret names
	TruststoreSecret = "-truststore"
	// TruststoreName used when creating Secret
	TruststoreName = "truststore.jks"
	// TruststorePath used when mounting Secret
	TruststorePath = "/etc/openshift-truststore-volume"
	// TruststorePwd used when creating Secret
	TruststorePwd = "changeit"
	// CaBundleKey ...
	CaBundleKey = "ca-bundle.crt"
	// HttpProtocol ...
	HttpProtocol = "http"
	// HttpsProtocol ...
	HttpsProtocol = "https"
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
	// ClusterLabel for Kube_ping
	ClusterLabel = "cluster"
	// ClusterLabelPrefix for Kube_ping
	ClusterLabelPrefix = "jgrp.k8s."
	// KubeNS Env name
	KubeNS         = "KUBERNETES_NAMESPACE"
	KubeLabels     = "KUBERNETES_LABELS"
	KafkaTopicsEnv = "KIE_SERVER_KAFKA_EXT_TOPICS"
	// ConsoleRouteEnv for backwards compatibility with Application Templates
	ConsoleRouteEnv = "BUSINESS_CENTRAL_HOSTNAME_HTTP"
	// ServersRouteEnv for backwards compatibility with Application Templates
	ServersRouteEnv = "KIE_SERVER_HOSTNAME_HTTP"
	// SmartRouterRouteEnv for backwards compatibility with Application Templates
	SmartRouterRouteEnv = "SMART_ROUTER_HOSTNAME_HTTP"
	// DashbuilderRouteEnv for backwards compatibility, similar to Console and Servers
	DashbuilderRouteEnv = "DASHBUILDER_HOSTNAME_HTTP"
	// ACFilters Default filters for CORS
	ACFilters = "AC_ALLOW_ORIGIN,AC_ALLOW_METHODS,AC_ALLOW_HEADERS,AC_ALLOW_CREDENTIALS,AC_MAX_AGE"

	relatedImageVar        = "RELATED_IMAGE_"
	PamKieImageVar         = relatedImageVar + "PAM_KIESERVER_IMAGE_"
	PamControllerVar       = relatedImageVar + "PAM_CONTROLLER_IMAGE_"
	PamBusinessCentralVar  = relatedImageVar + "PAM_BC_IMAGE_"
	PamBCMonitoringVar     = relatedImageVar + "PAM_BC_MONITORING_IMAGE_"
	PamProcessMigrationVar = relatedImageVar + "PAM_PROCESS_MIGRATION_IMAGE_"
	PamDashbuilderVar      = relatedImageVar + "PAM_DASHBUILDER_IMAGE_"
	PamSmartRouterVar      = relatedImageVar + "PAM_SMARTROUTER_IMAGE_"

	OauthVar             = relatedImageVar + "OAUTH_PROXY_IMAGE_"
	Oauth4ImageURL       = ImageRegistry + "/openshift4/ose-oauth-proxy"
	Oauth4ImageLatestURL = Oauth4ImageURL + ":latest"
	OauthComponent       = "golang-github-openshift-oauth-proxy-container"

	PostgreSQLVar         = relatedImageVar + "POSTGRESQL_PROXY_IMAGE_"
	PostgreSQL12ImageURL  = ImageRegistry + "/rhscl/postgresql-12-rhel7:latest"
	PostgreSQL12Component = "rh-postgresql12-container"
	PostgreSQL13ImageURL  = ImageRegistry + "/rhscl/postgresql-13-rhel7:latest"
	PostgreSQL13Component = "rh-postgresql13-container"

	MySQLVar         = relatedImageVar + "MYSQL_PROXY_IMAGE_"
	MySQL80ImageURL  = ImageRegistry + "/rhscl/mysql-80-rhel7:latest"
	MySQL80Component = "rh-mysql80-container"

	OseCliVar        = relatedImageVar + "OSE_CLI_IMAGE_"
	OseCli4Component = "openshift-enterprise-cli-container"

	BrokerVar            = relatedImageVar + "BROKER_IMAGE_"
	BrokerImage          = "amq-broker"
	Broker79Image        = "amq-broker-rhel8"
	BrokerImageURL       = ImageRegistry + "/amq7/" + BrokerImage + ":"
	Broker79ImageBaseURL = ImageRegistry + "/amq7/" + Broker79Image + ":"

	Broker78ImageTag = "7.8"
	Broker79ImageTag = "7.9"
	Broker78ImageURL = BrokerImageURL + Broker78ImageTag
	Broker79ImageURL = Broker79ImageBaseURL + Broker79ImageTag

	DatagridVar         = relatedImageVar + "DATAGRID_IMAGE_"
	Datagrid73Image     = "datagrid73-openshift"
	Datagrid73Component = "jboss-datagrid-7-datagrid73-openshift-container"

	Datagrid8Image     = "datagrid-8-rhel8"
	Datagrid8Component = "datagrid-datagrid-8-rhel8-container"

	Datagrid73ImageTag16 = "1.6"
	Datagrid73ImageURL16 = ImageRegistry + "/jboss-datagrid-7/" + Datagrid73Image + ":" + Datagrid73ImageTag16

	Datagrid8ImageTag11 = "1.1"
	Datagrid8ImageURL11 = ImageRegistry + "/datagrid/" + Datagrid8Image + ":" + Datagrid8ImageTag11

	PamContext  = "/" + IBMBamoeImageContext + "/" + IBMBamoeImagePrefix + "-"
	RhelVersion = "-rhel8"

	//Resources Limits and Requests
	ConsoleProdCPULimit         = "2"
	ConsoleProdMemLimit         = "2Gi"
	ConsoleAuthoringCPULimit    = "2"
	ConsoleAuthoringMemLimit    = "4Gi"
	ConsoleAuthoringCPURequests = "1500m"
	ConsoleAuthoringMemRequests = "3Gi"
	ConsoleProdCPURequests      = "1500m"
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
	ServersMemRequests          = "1536Mi"
	ServersM2PvSize             = "1Gi"
	ServersKiePvSize            = "10Mi"
	SmartRouterCPULimit         = "500m"
	SmartRouterMemLimit         = "1Gi"
	SmartRouterCPURequests      = "250m"
	SmartRouterMemRequests      = "1Gi"
	ProcessMigrationCPULimit    = "500m"
	ProcessMigrationMemLimit    = "512Mi"
	ProcessMigrationCPURequests = "250m"
	ProcessMigrationMemRequests = "512Mi"

	//ImageNames for metering labels
	RhpamSmartRouterImageName = IBMBamoeImagePrefix + "-smartrouter-" + RhelVersion
	RhpamControllerImageName  = IBMBamoeImagePrefix + "-controller-" + RhelVersion

	RhpamBusinessCentral    = "businesscentral"
	RhpamBusinessCentralMon = "businesscentral-monitoring"

	DashBuilder      = "dashbuilder"
	Smartrouter      = "smartrouter"
	ProcessMigration = "process-migration"
	Production       = "production"

	SUBCOMPONENT_TYPE_APP   = "application"
	SUBCOMPONENT_TYPE_INFRA = "infrastructure"

	DefaultDatagridUsername = "infinispan"

	USERNAME_ADMIN_SECRET_KEY = "username"
	PASSWORD_ADMIN_SECRET_KEY = "password"
)

var OseCli4ImageURL = ImageRegistry + "/openshift4/ose-cli:" + highestOcpVersion(Ocp4Versions)

// ConsoleProdLimits Console Resource Limits for BC Monitoring in Prod Env
var ConsoleProdLimits = map[string]string{
	"CPU": ConsoleProdCPULimit,
	"MEM": ConsoleProdMemLimit,
}

// ConsoleAuthoringLimits Resource Limits for BC in Authoring Env
var ConsoleAuthoringLimits = map[string]string{
	"CPU": ConsoleAuthoringCPULimit,
	"MEM": ConsoleAuthoringMemLimit,
}

// DashbuilderLimits Resource Limits for Dasubuilder
var DashbuilderLimits = map[string]string{
	"CPU": DashbuilderCPULimit,
	"MEM": DashbuilderMemLimit,
}

// ServersLimits Resource Limits for KIE Servers
var ServersLimits = map[string]string{
	"CPU": ServersCPULimit,
	"MEM": ServersMemLimit,
}

// SmartRouterLimits defines resource limits for Smart Router
var SmartRouterLimits = map[string]string{
	"CPU": SmartRouterCPULimit,
	"MEM": SmartRouterMemLimit,
}

// ProcessMigrationLimits defines resource limits for PIM
var ProcessMigrationLimits = map[string]string{
	"CPU": ProcessMigrationCPULimit,
	"MEM": ProcessMigrationMemLimit,
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

// DashbuilderRequests defines requests Dasubuilder deployment
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

// ProcessMigrationRequests defines the requests for PIM deployment
var ProcessMigrationRequests = map[string]string{
	"CPU": ProcessMigrationCPURequests,
	"MEM": ProcessMigrationMemRequests,
}

var Images = []ImageEnv{
	{
		Var:       PamKieImageVar,
		Component: IBMBamoeImageContext + "-7-kieserver-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "kieserver" + RhelVersion,
	},
	{
		Var:       PamControllerVar,
		Component: IBMBamoeImageContext + "-7-controller-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "controller" + RhelVersion,
	},
	{
		Var:       PamBusinessCentralVar,
		Component: IBMBamoeImageContext + "-7-businesscentral-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "businesscentral" + RhelVersion,
	},
	{
		Var:       PamBCMonitoringVar,
		Component: IBMBamoeImageContext + "-7-businesscentral-monitoring-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "businesscentral-monitoring" + RhelVersion,
	},
	{
		Var:       PamSmartRouterVar,
		Component: IBMBamoeImageContext + "-7-smartrouter-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "smartrouter" + RhelVersion,
	},
	{
		Var:       PamProcessMigrationVar,
		Component: IBMBamoeImageContext + "-7-process-migration-rhel8-container",
		Registry:  ImageRegistry,
		Context:   PamContext + "process-migration" + RhelVersion,
	},
	{
		Var:       PamDashbuilderVar,
		Component: IBMBamoeImageContext + "-7-dashbuilder-rhel8-container",
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

var rhpamAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: RhpamBusinessCentral, ImageVar: PamBusinessCentralVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central"}
var rhpamMonitorAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: RhpamBusinessCentralMon, ImageVar: PamBCMonitoringVar, MavenRepo: "RHPAMCENTR", FriendlyName: "Business Central Monitoring"}
var rhpamDashbuilderConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamdash", ImageName: "dashbuilder", ImageVar: PamDashbuilderVar, FriendlyName: "Dashbuilder"}
var RhdmAppConstants = api.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: RhpamBusinessCentral, ImageVar: PamBusinessCentralVar, MavenRepo: "RHDMCENTR", FriendlyName: "Business Central"}

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
	api.RhdmTrial:                  {App: RhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoring:              {App: RhdmAppConstants, Replica: replicasTrial},
	api.RhdmAuthoringHA:            {App: RhdmAppConstants, Replica: replicasAuthoringHA},
	api.RhdmProductionImmutable:    {App: RhdmAppConstants, Replica: replicasRhpamProduction},
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

func highestOcpVersion(versions []string) string {
	highest := ""
	for _, ver := range versions {
		ver = "v" + ver
		if semver.IsValid(ver) && semver.Compare(ver, highest) > 0 {
			highest = ver
		}
	}
	return highest
}
