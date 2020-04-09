package v2

import (
	"github.com/RHsyseng/operator-utils/pkg/olm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType - type of condition
type ConditionType string

const (
	// DeployedConditionType - the kieapp is deployed
	DeployedConditionType ConditionType = "Deployed"
	// ProvisioningConditionType - the kieapp is being provisioned
	ProvisioningConditionType ConditionType = "Provisioning"
	// FailedConditionType - the kieapp is in a failed state
	FailedConditionType ConditionType = "Failed"
)

// ReasonType - type of reason
type ReasonType string

const (
	// DeploymentFailedReason - Unable to deploy the application
	DeploymentFailedReason ReasonType = "DeploymentFailed"
	// ConfigurationErrorReason - An invalid configuration caused an error
	ConfigurationErrorReason ReasonType = "ConfigurationError"
	// MissingDependenciesReason - Dependencies does not exist or cannot be found
	MissingDependenciesReason ReasonType = "MissingDependencies"
	// UnknownReason - Unable to determine the error
	UnknownReason ReasonType = "Unknown"
)

// Condition - The condition for the kie-cloud-operator
type Condition struct {
	Type               ConditionType          `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ReasonType             `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Version            string                 `json:"version,omitempty"`
}

// KieAppStatus - The status for custom resources managed by the operator-sdk.
type KieAppStatus struct {
	Conditions  []Condition          `json:"conditions"`
	ConsoleHost string               `json:"consoleHost,omitempty"`
	Deployments olm.DeploymentStatus `json:"deployments"`
	Phase       ConditionType        `json:"phase,omitempty"`
	Generated   KieAppSpec           `json:"generated,omitempty"`
	Version     string               `json:"version,omitempty"`
}
