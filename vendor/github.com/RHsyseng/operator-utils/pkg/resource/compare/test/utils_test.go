package test

import (
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestEmptyArray(t *testing.T) {
	builder := compare.NewMapBuilder()
	assert.Empty(t, builder.ResourceMap(), "Expected empty map")
}

func TestMapBuilder(t *testing.T) {
	resMap := compare.NewMapBuilder().Add(&routev1.Route{}, &routev1.Route{}, &corev1.Service{}).Add(&oappsv1.DeploymentConfig{}).ResourceMap()
	assert.Len(t, resMap, 3, "Expect map to have 3 entries")
	assert.Len(t, resMap[reflect.TypeOf(routev1.Route{})], 2, "Expect map to have 2 routes")
	assert.Len(t, resMap[reflect.TypeOf(corev1.Service{})], 1, "Expect map to have 1 service")
	assert.Len(t, resMap[reflect.TypeOf(oappsv1.DeploymentConfig{})], 1, "Expect map to have 1 deployment config")
}
