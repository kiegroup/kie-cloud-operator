package constants

import (
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	corev1 "k8s.io/api/core/v1"
)

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
	// ProductVersion default version
	ProductVersion = "7.4"
	// ImageStreamTag default tag name for the ImageStreams
<<<<<<< HEAD
	ImageStreamTag = "1.0"
	// AMQ Broker Image Name
=======
	ImageStreamTag = "1.1"
	// BrokerImage AMQ Broker Image Name
>>>>>>> 8a6aa8ce... [KIECLOUD-254] Allow default value for JNDI Name with external databases
	BrokerImage = "amq-broker-73-openshift"
	// BrokerImageTag AMQ Broker Image Tag
	BrokerImageTag = "7.3"
	// DatagridImage JBoss Datagrid Image Name
	DatagridImage = "datagrid73-openshift"
	// DatagridImageTag JBoss Datagrid  Image Tag
	DatagridImageTag = "1.1"
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
	// DefaultJNDIName to use for external databases
	DefaultJNDIName = "java:jboss/datasources/jbpmDS"
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
)

var rhpamAppConstants = v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", MavenRepo: "RHPAMCENTR"}
var rhpamMonitorAppConstants = v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", MavenRepo: "RHPAMCENTR"}
var rhdmAppConstants = v1.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: "decisioncentral", MavenRepo: "RHDMCENTR"}

var replicasTrial = v1.ReplicaConstants{
	Console:     v1.Replicas{Replicas: 1, DenyScale: true},
	Server:      v1.Replicas{Replicas: 1},
	SmartRouter: v1.Replicas{Replicas: 1},
}
var replicasRhpamProductionImmutable = v1.ReplicaConstants{
	Console:     v1.Replicas{Replicas: 1},
	Server:      v1.Replicas{Replicas: 2},
	SmartRouter: v1.Replicas{Replicas: 1},
}
var replicasRhpamProduction = v1.ReplicaConstants{
	Console:     v1.Replicas{Replicas: 3},
	Server:      v1.Replicas{Replicas: 3},
	SmartRouter: v1.Replicas{Replicas: 1},
}
var replicasAuthoringHA = v1.ReplicaConstants{
	Console:     v1.Replicas{Replicas: 2},
	Server:      v1.Replicas{Replicas: 2},
	SmartRouter: v1.Replicas{Replicas: 1},
}

// DefaultDatabaseConfig defines the default Database to use for each environment
var databaseRhpamAuthoring = &v1.DatabaseObject{Type: v1.DatabaseH2, Size: DefaultDatabaseSize}
var databaseRhpamAuthoringHA = &v1.DatabaseObject{Type: v1.DatabaseMySQL, Size: DefaultDatabaseSize}
var databaseRhpamProduction = &v1.DatabaseObject{Type: v1.DatabasePostgreSQL, Size: DefaultDatabaseSize}
var databaseRhpamProductionImmutable = &v1.DatabaseObject{Type: v1.DatabasePostgreSQL, Size: DefaultDatabaseSize}
var databaseRhpamTrial = &v1.DatabaseObject{Type: v1.DatabaseH2, Size: ""}

// EnvironmentConstants contains
var EnvironmentConstants = map[v1.EnvironmentType]*v1.EnvironmentConstants{
	v1.RhpamProduction:          {App: rhpamMonitorAppConstants, Replica: replicasRhpamProduction, Database: databaseRhpamProduction},
	v1.RhpamProductionImmutable: {App: rhpamMonitorAppConstants, Replica: replicasRhpamProductionImmutable, Database: databaseRhpamProductionImmutable},
	v1.RhpamTrial:               {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamTrial},
	v1.RhpamAuthoring:           {App: rhpamAppConstants, Replica: replicasTrial, Database: databaseRhpamAuthoring},
	v1.RhpamAuthoringHA:         {App: rhpamAppConstants, Replica: replicasAuthoringHA, Database: databaseRhpamAuthoringHA},
	v1.RhdmTrial:                {App: rhdmAppConstants, Replica: replicasTrial},
	v1.RhdmAuthoring:            {App: rhdmAppConstants, Replica: replicasTrial},
	v1.RhdmAuthoringHA:          {App: rhdmAppConstants, Replica: replicasAuthoringHA},
	v1.RhdmProductionImmutable:  {App: rhdmAppConstants, Replica: replicasTrial},
}

// TemplateConstants set of constant values to use in templates
var TemplateConstants = v1.TemplateConstants{
	KeystoreVolumeSuffix: KeystoreVolumeSuffix,
	DatabaseVolumeSuffix: DatabaseVolumeSuffix,
	BrokerImage:          BrokerImage,
	BrokerImageTag:       BrokerImageTag,
	DatagridImage:        DatagridImage,
	DatagridImageTag:     DatagridImageTag,
}

var DebugTrue = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "true",
}

var DebugFalse = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "false",
}
