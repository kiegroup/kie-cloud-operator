package status

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = logs.GetLogger("kieapp.controller")

const maxBuffer = 30

// SetProvisioning - Sets the condition type to Provisioning and status True if not yet set.
func SetProvisioning(cr *v1.KieApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == v1.ProvisioningConditionType {
		log.Debug("Status: unchanged status [provisioning].")
		return false
	}
	log.Debug("Status: set provisioning")
	condition := v1.Condition{
		Type:               v1.ProvisioningConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
	return true
}

// SetDeployed - Updates the condition with the DeployedCondition and True status
func SetDeployed(cr *v1.KieApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == v1.DeployedConditionType {
		log.Debug("Status: unchanged status [deployed].")
		return false
	}
	log.Debugf("Status: changed status [deployed].")
	condition := v1.Condition{
		Type:               v1.DeployedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
	return true
}

// SetFailed - Sets the failed condition with the error reason and message
func SetFailed(cr *v1.KieApp, reason v1.ReasonType, err error) {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	log.Debug("Status: set failed")
	condition := v1.Condition{
		Type:               v1.FailedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            err.Error(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
}

func addCondition(conditions []v1.Condition, condition v1.Condition) []v1.Condition {
	size := len(conditions) + 1
	first := 0
	if size > maxBuffer {
		first = size - maxBuffer
	}
	return append(conditions, condition)[first:size]
}
