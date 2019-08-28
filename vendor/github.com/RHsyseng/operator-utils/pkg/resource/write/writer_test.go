package write

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFluentAPI(t *testing.T) {
	noOwnership := New()
	assert.Nil(t, noOwnership.ownerRefs, "Do not expect ownerRefs to be set")
	assert.Nil(t, noOwnership.ownerController, "Do not expect ownerController to be set")

	ownerRefs := New().WithOwnerController(&corev1.Service{})
	assert.Nil(t, ownerRefs.ownerRefs, "Do not expect ownerRefs to be set")
	assert.NotNil(t, ownerRefs.ownerController, "Expect ownerController to be set")

	controler := New().WithOwnerReferences(v1.OwnerReference{})
	assert.NotNil(t, controler.ownerRefs, "Expect ownerRefs to be set")
	assert.Nil(t, controler.ownerController, "Do not expect ownerController to be set")
}
