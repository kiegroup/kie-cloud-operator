package test

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestCompareServices(t *testing.T) {
	svcs := test.GetServices(2)
	svcs[0].Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					IP:       "127.0.0.1",
					Hostname: "localhost",
				},
			},
		},
	}
	svcs[1].Name = svcs[0].Name

	assert.False(t, reflect.DeepEqual(svcs[0], svcs[1]), "Inconsequential differences between two services should make equality test fail")
	assert.True(t, compare.SimpleComparator().Compare(&svcs[0], &svcs[1]), "Expected resources to be deemed equal")
	assert.True(t, compare.DefaultComparator().Compare(&svcs[0], &svcs[1]), "Expected resources to be deemed equal based on service comparator")
}

func TestCompareDeploymentConfigs(t *testing.T) {
	dcs := test.GetDeploymentConfigs(2)
	dcs[1].Name = dcs[0].Name
	dcs[1].Status = oappsv1.DeploymentConfigStatus{
		ReadyReplicas: 1,
	}

	assert.False(t, reflect.DeepEqual(dcs[0], dcs[1]), "Inconsequential differences between two DCs should make equality test fail")
	assert.True(t, compare.SimpleComparator().Compare(&dcs[0], &dcs[1]), "Expected resources to be deemed equal")
	assert.True(t, compare.DefaultComparator().Compare(&dcs[0], &dcs[1]), "Expected resources to be deemed equal based on DC comparator")
}

func TestCompareCombined(t *testing.T) {
	dcs := test.GetDeploymentConfigs(6)
	dc1a := dcs[0]
	dc1b := dcs[1]
	dc2a := dcs[2]
	dc2b := dcs[3]
	dc3a := dcs[4]
	dc4b := dcs[5]
	dc1a.Name = "dc1"
	dc1b.Name = "dc1"
	dc2a.Name = "dc2"
	dc2b.Name = "dc2"
	dc3a.Name = "dc3"
	dc4b.Name = "dc4"
	dc1b.Status = oappsv1.DeploymentConfigStatus{
		Replicas: 2,
	}
	dc2b.Spec.Replicas = 2

	svcs := test.GetServices(6)
	service1a := svcs[0]
	service1b := svcs[1]
	service2a := svcs[2]
	service2b := svcs[3]
	service3a := svcs[4]
	service4b := svcs[5]
	service1a.Name = "service1"
	service1b.Name = "service1"
	service2a.Name = "service2"
	service2b.Name = "service2"
	service3a.Name = "service3"
	service4b.Name = "service4"
	service1b.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					IP:       "127.0.0.1",
					Hostname: "localhost",
				},
			},
		},
	}
	service2b.Spec = corev1.ServiceSpec{
		ClusterIP: "127.0.0.1",
	}

	serviceType := reflect.TypeOf(corev1.Service{})
	dcType := reflect.TypeOf(oappsv1.DeploymentConfig{})
	deployed := map[reflect.Type][]resource.KubernetesResource{
		dcType:      {&dc1a, &dc2a, &dc3a},
		serviceType: {&service1a, &service2a, &service3a},
	}
	requested := map[reflect.Type][]resource.KubernetesResource{
		dcType:      {&dc1b, &dc2b, &dc4b},
		serviceType: {&service1b, &service2b, &service4b},
	}

	mapComparator := compare.NewMapComparator()
	deltaMap := mapComparator.Compare(deployed, requested)

	assert.Len(t, deltaMap[serviceType].Added, 1, "Expected 1 added service")
	assert.Equal(t, deltaMap[serviceType].Added[0].GetName(), "service4", "Expected added service called service4")
	assert.Len(t, deltaMap[serviceType].Updated, 1, "Expected 1 updated service")
	assert.Equal(t, deltaMap[serviceType].Updated[0].GetName(), "service2", "Expected added service called service2")
	assert.Len(t, deltaMap[serviceType].Removed, 1, "Expected 1 removed service")
	assert.Equal(t, deltaMap[serviceType].Removed[0].GetName(), "service3", "Expected added service called service3")
	assert.Len(t, deltaMap[dcType].Added, 1, "Expected 1 added dc")
	assert.Equal(t, deltaMap[dcType].Added[0].GetName(), "dc4", "Expected added dc called dc4")
	assert.Len(t, deltaMap[dcType].Updated, 1, "Expected 1 updated dc")
	assert.Equal(t, deltaMap[dcType].Updated[0].GetName(), "dc2", "Expected updated dc called dc2")
	assert.Len(t, deltaMap[dcType].Removed, 1, "Expected 1 removed dc")
	assert.Equal(t, deltaMap[dcType].Removed[0].GetName(), "dc3", "Expected removed dc called dc3")
}
