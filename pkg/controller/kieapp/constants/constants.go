package constants

import (
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	corev1 "k8s.io/api/core/v1"
)

const (
	// CurrentVersion product version supported
	CurrentVersion = "7.6.0"
	// LastMicroVersion product version supported
	LastMicroVersion = "7.5.1"
	// LastMinorVersion product version supported
	LastMinorVersion = "7.5.0"
)

// SupportedVersions - product versions this operator supports
var SupportedVersions = []string{CurrentVersion, LastMicroVersion, LastMinorVersion}

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
	// ConsoleLinkName is how the link will be titled in an installed CSV within the marketplace
	ConsoleLinkName = "Installer"
	// ConsoleDescription is how the link will be described in an installed CSV within the marketplace
	ConsoleDescription = "**To use the guided installer to provision an environment, open the Installer link, in the links section on the left side of this page.**"
	// SmartRouterProtocol - default SmartRouter protocol
	SmartRouterProtocol = "http"
)

// VersionConstants ...
var VersionConstants = map[string]*api.VersionConfigs{
	CurrentVersion: {
		APIVersion:       api.SchemeGroupVersion.Version,
		ImageTag:         CurrentVersion,
		BrokerImage:      "amq-broker",
		BrokerImageTag:   "7.4",
		DatagridImage:    "datagrid73-openshift",
		DatagridImageTag: "1.2",
	},
	LastMicroVersion: {
		APIVersion:       api.SchemeGroupVersion.Version,
		ImageTag:         CurrentVersion,
		BrokerImage:      "amq-broker",
		BrokerImageTag:   "7.4",
		DatagridImage:    "datagrid73-openshift",
		DatagridImageTag: "1.2",
	},
	LastMinorVersion: {
		APIVersion:       api.SchemeGroupVersion.Version,
		ImageTag:         LastMinorVersion,
		BrokerImage:      "amq-broker",
		BrokerImageTag:   "7.4",
		DatagridImage:    "datagrid73-openshift",
		DatagridImageTag: "1.1",
	},
}

var rhpamAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", MavenRepo: "RHPAMCENTR"}
var rhpamMonitorAppConstants = api.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", MavenRepo: "RHPAMCENTR"}
var rhdmAppConstants = api.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: "decisioncentral", MavenRepo: "RHDMCENTR"}

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
