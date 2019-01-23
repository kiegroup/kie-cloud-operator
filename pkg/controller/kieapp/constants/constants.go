package constants

const (
	// RhpamcentrServicePrefix prefix to use for the console
	RhpamcentrServicePrefix = "rhpamcentr"
	// RhpamcentrMonitoringServicePrefix prefix to use for the monitoring console
	RhpamcentrMonitoringServicePrefix = "rhpamcentrmon"
	// RhpamcentrImageName image name of the console
	RhpamcentrImageName = "businesscentral"
	// RhpamcentrMonitoringImageName image name of the monitoring console
	RhpamcentrMonitoringImageName = "businesscentral-monitoring"
	// KieServerServicePrefix prefix to use for the servers
	KieServerServicePrefix = "kieserver"
	// RhpamRegistry default registry
	RhpamRegistry = "registry.redhat.io"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
	// RhpamVersion default version
	RhpamVersion = "7.2"
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "1.0"
	// ConfigMapPrefix prefix to use for the configmaps
	ConfigMapPrefix = "kieconfigs"
	// DefaultPassword default password to use for test environments
	DefaultPassword = "RedHat"
	// SSODefaultPrincipalAttribute default PrincipalAttribute to use for SSO integration
	SSODefaultPrincipalAttribute = "preferred_username"
)

// MonitoringEnvs Type of environments that will deploy the Monitoring console.
// The console resources will be suffixed as -monitoring as well
var MonitoringEnvs = map[string]bool{
	"production":           true,
	"production-immutable": true,
}
