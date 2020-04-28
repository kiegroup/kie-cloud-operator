package status

import (
	"fmt"
	"testing"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetDeployed(t *testing.T) {
	now := metav1.Now()
	cr := &api.KieApp{Spec: api.KieAppSpec{Version: constants.CurrentVersion}}

	assert.True(t, SetDeployed(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	assert.Equal(t, api.DeployedConditionType, cr.Status.Conditions[0].Type)
	assert.Equal(t, api.DeployedConditionType, cr.Status.Phase)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[0].Status)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Conditions[0].Version)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Version)
	assert.True(t, now.Before(&cr.Status.Conditions[0].LastTransitionTime))
}

func TestSetDeployedSkipUpdate(t *testing.T) {
	cr := &api.KieApp{}
	SetDeployed(cr)

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]

	assert.False(t, SetDeployed(cr))
	assert.Equal(t, 1, len(cr.Status.Conditions))
	assert.Equal(t, condition, cr.Status.Conditions[0])
	assert.Equal(t, condition.Type, cr.Status.Phase)
}

func TestSetProvisioning(t *testing.T) {
	now := metav1.Now()
	cr := &api.KieApp{Spec: api.KieAppSpec{Version: constants.CurrentVersion}}
	assert.True(t, SetProvisioning(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Phase)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[0].Status)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Conditions[0].Version)
	assert.NotEqual(t, constants.CurrentVersion, cr.Status.Version)
	assert.True(t, now.Before(&cr.Status.Conditions[0].LastTransitionTime))
}

func TestSetProvisioningSkipUpdate(t *testing.T) {
	cr := &api.KieApp{}
	assert.True(t, SetProvisioning(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]

	assert.False(t, SetProvisioning(cr))
	assert.Equal(t, 1, len(cr.Status.Conditions))
	assert.Equal(t, condition, cr.Status.Conditions[0])
	assert.Equal(t, condition.Type, cr.Status.Phase)
}

func TestSetProvisioningAndThenDeployed(t *testing.T) {
	now := metav1.Now()
	cr := &api.KieApp{Spec: api.KieAppSpec{Version: constants.PriorVersion1}}

	assert.True(t, SetProvisioning(cr))
	assert.True(t, SetDeployed(cr))
	defaults.SetVersion(cr, constants.CurrentVersion)
	assert.True(t, SetProvisioning(cr))
	assert.True(t, SetDeployed(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]
	assert.Equal(t, 4, len(cr.Status.Conditions))

	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[0].Status)
	assert.Equal(t, constants.PriorVersion1, cr.Status.Conditions[0].Version)
	assert.True(t, now.Before(&condition.LastTransitionTime))

	assert.Equal(t, api.DeployedConditionType, cr.Status.Conditions[1].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[1].Status)
	assert.True(t, condition.LastTransitionTime.Before(&cr.Status.Conditions[1].LastTransitionTime))
	assert.Equal(t, constants.PriorVersion1, cr.Status.Conditions[1].Version)

	assert.Equal(t, api.ProvisioningConditionType, cr.Status.Conditions[2].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[2].Status)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Conditions[2].Version)
	assert.True(t, condition.LastTransitionTime.Before(&cr.Status.Conditions[2].LastTransitionTime))

	assert.Equal(t, api.DeployedConditionType, cr.Status.Conditions[3].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[3].Status)
	assert.True(t, condition.LastTransitionTime.Before(&cr.Status.Conditions[3].LastTransitionTime))
	assert.Equal(t, constants.CurrentVersion, cr.Status.Conditions[3].Version)
	assert.Equal(t, constants.CurrentVersion, cr.Status.Version)
	assert.Equal(t, api.DeployedConditionType, cr.Status.Phase)
}

func TestBuffer(t *testing.T) {
	cr := &api.KieApp{Spec: api.KieAppSpec{Version: constants.CurrentVersion}}
	for i := 0; i < maxBuffer+2; i++ {
		SetFailed(cr, api.UnknownReason, fmt.Errorf("Error %d", i))
	}
	size := len(cr.Status.Conditions)
	assert.Equal(t, maxBuffer, size)
	assert.Equal(t, "Error 31", cr.Status.Conditions[size-1].Message)
}
