package defaults

import (
	"fmt"
	"testing"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/kiegroup/kie-cloud-operator/version"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpgradesTrue(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
			Upgrades:    api.KieAppUpgrades{Enabled: true},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err)
	assert.True(t, cr.Spec.Upgrades.Enabled, "Spec.Upgrades.Enabled should be true")
}

func TestCheckProductUpgrade(t *testing.T) {
	// Incompatible version
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     "6.3.1",
			Upgrades:    api.KieAppUpgrades{Minor: true, Enabled: true},
		},
	}
	minor, micro, err := checkProductUpgrade(cr)
	assert.Error(t, err, "Incompatible product versions should throw an error")
	assert.Equal(t, fmt.Sprintf("Product version %s is not supported in operator version %s. The following versions are supported - %s", cr.Spec.Version, version.Version, constants.SupportedVersions), err.Error())
	assert.False(t, minor)
	assert.False(t, micro)

	// Upgrades default to false
	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     constants.PastMicroVersion,
		},
	}
	minor, micro, err = checkProductUpgrade(cr)
	assert.Nil(t, err)
	assert.False(t, minor)
	assert.False(t, micro)

	// Micro set to true
	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     constants.PastMicroVersion,
			Upgrades:    api.KieAppUpgrades{Enabled: true},
		},
	}
	minor, micro, err = checkProductUpgrade(cr)
	assert.Nil(t, err)
	assert.False(t, minor)
	assert.True(t, micro)

	// Past version, all upgrades true
	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     constants.PastMinorVersion,
			Upgrades:    api.KieAppUpgrades{Minor: true, Enabled: true},
		},
	}
	minor, micro, err = checkProductUpgrade(cr)
	assert.Nil(t, err)
	assert.True(t, minor)
	assert.True(t, micro)

	// Current version, no upgrades
	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     constants.CurrentVersion,
			Upgrades:    api.KieAppUpgrades{Minor: true, Enabled: true},
		},
	}
	minor, micro, err = checkProductUpgrade(cr)
	assert.Nil(t, err)
	assert.False(t, minor)
	assert.False(t, micro)

	// Upgrades disabled with minor true
	cr = &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamProduction,
			Version:     constants.PastMinorVersion,
			Upgrades:    api.KieAppUpgrades{Minor: true, Enabled: false},
		},
	}
	minor, micro, err = checkProductUpgrade(cr)
	assert.Nil(t, err)
	assert.False(t, minor)
	assert.False(t, micro)
}
