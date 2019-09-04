package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeploymentConfigGeneration(t *testing.T) {
	dcs := GetDeploymentConfigs(3)
	assert.Len(t, dcs, 3, "Expected 3 DCs in the array")
	for i := 0; i < len(dcs); i++ {
		assert.Equal(t, fmt.Sprintf("%s%d", "dc", (i+1)), dcs[i].Name, "DeploymentConfig name does not have the expected value")
	}
}
