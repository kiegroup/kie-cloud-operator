package compare

import (
	utils "github.com/RHsyseng/operator-utils/pkg/resource/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestCompareRoutes(t *testing.T) {
	routes := utils.GetRoutes(2)
	routes[0].Status = routev1.RouteStatus{
		Ingress: []routev1.RouteIngress{
			{
				Host: "localhost",
			},
		},
	}
	routes[1].Name = routes[0].Name

	assert.False(t, reflect.DeepEqual(routes[0], routes[1]), "Inconsequential differences between two routes should make equality test fail")
	assert.True(t, deepEquals(&routes[0], &routes[1]), "Expected resources to be deemed equal")
	assert.True(t, equalRoutes(&routes[0], &routes[1]), "Expected resources to be deemed equal based on route comparator")
}

func TestCompareServices(t *testing.T) {
	services := utils.GetServices(2)
	services[0].Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					IP:       "127.0.0.1",
					Hostname: "localhost",
				},
			},
		},
	}
	services[1].Name = services[0].Name

	assert.False(t, reflect.DeepEqual(services[0], services[1]), "Inconsequential differences between two services should make equality test fail")
	assert.True(t, deepEquals(&services[0], &services[1]), "Expected resources to be deemed equal")
	assert.True(t, equalServices(&services[0], &services[1]), "Expected resources to be deemed equal based on service comparator")
}

func TestCompareDeploymentConfigs(t *testing.T) {
	dcs := utils.GetDeploymentConfigs(2)
	dcs[1].Name = dcs[0].Name
	dcs[1].Status = oappsv1.DeploymentConfigStatus{
		ReadyReplicas: 1,
	}

	assert.False(t, reflect.DeepEqual(dcs[0], dcs[1]), "Inconsequential differences between two DCs should make equality test fail")
	assert.True(t, deepEquals(&dcs[0], &dcs[1]), "Expected resources to be deemed equal")
	assert.True(t, equalDeploymentConfigs(&dcs[0], &dcs[1]), "Expected resources to be deemed equal based on DC comparator")
}
