package shared

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestResourceRequirements(t *testing.T) {
	reqs := GetResourceRequirements(map[string]map[v1.ResourceName]string{"Limits": {v1.ResourceMemory: "1Gi", v1.ResourceCPU: "2"}, "Requests": {v1.ResourceMemory: "500Mi"}})
	logrus.Debugf("Resource Requirements: %v", reqs)
	assert.Equal(t, *reqs.Limits.Memory(), resource.MustParse("1Gi"))
	assert.Equal(t, *reqs.Limits.Cpu(), resource.MustParse("2"))
	assert.Equal(t, *reqs.Requests.Memory(), resource.MustParse("500Mi"))
}
