package v2

import (
	"reflect"
	"testing"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoundTripFromV1ToV2(t *testing.T) {
	patch := true
	testObj := v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-new-object",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "app.kiegroup.org/v1",
			Kind:       "KieApp",
		},
		Spec: v1.KieAppSpec{
			CommonConfig: v1.CommonConfig{
				Version: "7.4.0",
			},
			Upgrades: v1.KieAppUpgrades{
				Patch: &patch,
			},
		},
	}
	testRoundTripFromV1(t, testObj)
}

func testRoundTripFromV1(t *testing.T, v1Object v1.KieApp) {
	v2Object, err := convertKieAppV1toV2(&v1Object)
	if err != nil {
		t.Fatalf("failed to convert v1 crontab to v2: %v", err)
	}
	assert.Equal(t, v1Object.Spec.CommonConfig.Version, v2Object.Spec.Version)
	assert.Equal(t, v1Object.Spec.Upgrades.Patch, &v2Object.Spec.Upgrades.Enabled)

	v2Object.Spec.Upgrades.Enabled = false
	v2Object.Spec.Version = "7.4.1"

	v1Object2, err := convertKieAppV2toV1(v2Object)
	if err != nil {
		t.Fatalf("failed to convert v2 crontab to v1: %v", err)
	}
	if !reflect.DeepEqual(v1Object2, v1Object2) {
		t.Errorf("round tripping failed for v1 crontab. v1Object: %v, v2Object: %v, v1ObjectConverted: %v",
			v1Object, v2Object, v1Object2)
	}
	assert.Equal(t, v1Object2.Spec.CommonConfig.Version, v2Object.Spec.Version)
	assert.Equal(t, *v1Object2.Spec.Upgrades.Patch, v2Object.Spec.Upgrades.Enabled)

	assert.Equal(t, v1Object.APIVersion, v1Object2.APIVersion)
}
