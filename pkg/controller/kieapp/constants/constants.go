package constants

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
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
)

var rhpamAppConstants = &v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentr", ImageName: "businesscentral", MavenRepo: "RHPAMCENTR", ConsoleProbePage: "kie-wb.jsp"}
var rhpamMonitorAppConstants = &v1.AppConstants{Product: RhpamPrefix, Prefix: "rhpamcentrmon", ImageName: "businesscentral-monitoring", MavenRepo: "RHPAMCENTR", ConsoleProbePage: "kie-wb.jsp"}
var rhdmAppConstants = &v1.AppConstants{Product: RhdmPrefix, Prefix: "rhdmcentr", ImageName: "decisioncentral", MavenRepo: "RHDMCENTR", ConsoleProbePage: "kie-wb.jsp"}

// EnvironmentConstants contains
var EnvironmentConstants = map[v1.EnvironmentType]*v1.AppConstants{
	v1.RhpamProduction:          rhpamMonitorAppConstants,
	v1.RhpamProductionImmutable: rhpamMonitorAppConstants,
	v1.RhpamTrial:               rhpamAppConstants,
	v1.RhpamAuthoring:           rhpamAppConstants,
	v1.RhpamAuthoringHA:         rhpamAppConstants,
	v1.RhdmTrial:                rhdmAppConstants,
	v1.RhdmAuthoring:            rhdmAppConstants,
	v1.RhdmAuthoringHA:          rhdmAppConstants,
	v1.RhdmOptawebTrial:         rhdmAppConstants,
	v1.RhdmProductionImmutable:  rhdmAppConstants,
}
