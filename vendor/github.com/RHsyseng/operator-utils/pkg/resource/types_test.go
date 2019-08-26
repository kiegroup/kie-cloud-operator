package resource

import (
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestInterfaceAssignability(t *testing.T) {
	assert.True(t, isAssignable(&routev1.Route{}), "Expected Route to implement KubernetesResource")
	assert.True(t, isAssignable(&corev1.Service{}), "Expected Service to implement KubernetesResource")
	assert.False(t, isAssignable(&corev1.Volume{}), "Expected Volume to NOT implement KubernetesResource")
	assert.True(t, isAssignable(&oappsv1.DeploymentConfig{}), "Expected DeploymentConfig to implement KubernetesResource")
}

func isAssignable(object interface{}) bool {
	_, assignable := interface{}(object).(KubernetesResource)
	return assignable
}
