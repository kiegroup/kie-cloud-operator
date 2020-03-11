package status

import (
	"github.com/RHsyseng/operator-utils/pkg/logs"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = logs.GetLogger("kieapp.controller")

const maxBuffer = 30

// SetProvisioning - Sets the condition type to Provisioning and status True if not yet set.
func SetProvisioning(cr *api.KieApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == api.ProvisioningConditionType {
		log.Debug("Status: unchanged status [provisioning].")
		return false
	}
	log.Debug("Status: set provisioning")
	condition := api.Condition{
		Type:               api.ProvisioningConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr, condition)
	return true
}

// SetDeployed - Updates the condition with the DeployedCondition and True status
func SetDeployed(cr *api.KieApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == api.DeployedConditionType {
		log.Debug("Status: unchanged status [deployed].")
		return false
	}
	log.Debugf("Status: changed status [deployed].")
	condition := api.Condition{
		Type:               api.DeployedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr, condition)
	return true
}

// SetFailed - Sets the failed condition with the error reason and message
func SetFailed(cr *api.KieApp, reason api.ReasonType, err error) {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	log.Debug("Status: set failed")
	condition := api.Condition{
		Type:               api.FailedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            err.Error(),
	}
	cr.Status.Conditions = addCondition(cr, condition)
}

func addCondition(cr *api.KieApp, condition api.Condition) []api.Condition {
	conditions := cr.Status.Conditions
	size := len(conditions) + 1
	first := 0
	if size > maxBuffer {
		first = size - maxBuffer
	}
	cr.Status.Phase = condition.Type
	return append(conditions, condition)[first:size]
}
