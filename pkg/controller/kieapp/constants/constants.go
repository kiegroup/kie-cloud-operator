package constants

import (
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
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
	ProductVersion = "7.3"
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "1.0"
	// ConfigMapPrefix prefix to use for the configmaps
	ConfigMapPrefix = "kieconfigs"
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
	// TrialEnvSuffix is the suffix for trial environments
	TrialEnvSuffix = "trial"
	// DefaultKieDeployments default number of Kie Server deployments
	DefaultKieDeployments = 1
	// KeystoreSecret is the default format for keystore secret names
	KeystoreSecret = "%s-app-secret"
)

var rhpamAppConstants = v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", MavenRepo: "RHPAMCENTR", ConsoleProbePage: "kie-wb.jsp"}
var rhpamMonitorAppConstants = v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", MavenRepo: "RHPAMCENTR", ConsoleProbePage: "kie-wb.jsp"}
var rhdmAppConstants = v1.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: "decisioncentral", MavenRepo: "RHDMCENTR", ConsoleProbePage: "kie-wb.jsp"}

var ReplicasTrial = v1.ReplicaConstants{
	Console:     v1.Replicas{Replicas: 1, DenyScale: true},
	Server:      v1.Replicas{Replicas: 2},
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

// EnvironmentConstants contains
var EnvironmentConstants = map[v1.EnvironmentType]*v1.EnvironmentConstants{
	v1.RhpamProduction:          &v1.EnvironmentConstants{AppConstants: rhpamMonitorAppConstants, ReplicaConstants: replicasRhpamProduction},
	v1.RhpamProductionImmutable: &v1.EnvironmentConstants{AppConstants: rhpamMonitorAppConstants, ReplicaConstants: replicasRhpamProductionImmutable},
	v1.RhpamTrial:               &v1.EnvironmentConstants{AppConstants: rhpamAppConstants, ReplicaConstants: ReplicasTrial},
	v1.RhpamAuthoring:           &v1.EnvironmentConstants{AppConstants: rhpamAppConstants, ReplicaConstants: ReplicasTrial},
	v1.RhpamAuthoringHA:         &v1.EnvironmentConstants{AppConstants: rhpamAppConstants, ReplicaConstants: replicasAuthoringHA},
	v1.RhdmTrial:                &v1.EnvironmentConstants{AppConstants: rhdmAppConstants, ReplicaConstants: ReplicasTrial},
	v1.RhdmAuthoring:            &v1.EnvironmentConstants{AppConstants: rhdmAppConstants, ReplicaConstants: ReplicasTrial},
	v1.RhdmAuthoringHA:          &v1.EnvironmentConstants{AppConstants: rhdmAppConstants, ReplicaConstants: replicasAuthoringHA},
	v1.RhdmProductionImmutable:  &v1.EnvironmentConstants{AppConstants: rhdmAppConstants, ReplicaConstants: ReplicasTrial},
}
