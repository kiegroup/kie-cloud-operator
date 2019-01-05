package constants

const (
	RhpamcentrServicePrefix           = "rhpamcentr"
	RhpamcentrMonitoringServicePrefix = "rhpamcentrmon"
	RhpamcentrImageName               = "businesscentral"
	RhpamcentrMonitoringImageName     = "businesscentral-monitoring"
	KieServerServicePrefix            = "kieserver"
	RhpamRegistry                     = "registry.access.redhat.com"
	ImageStreamNamespace              = "openshift"
	RhpamVersion                      = "7.2"
	ImageStreamTag                    = "1.0"
	ConfigMapPrefix                   = "kieconfigs"
)

// MonitoringEnvs Type of environments that will deploy the Monitoring console.
// The console resources will be suffixed as -monitoring as well
var MonitoringEnvs = map[string]bool{
	"production":           true,
	"production-immutable": true,
}
